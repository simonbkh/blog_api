const { createApp, reactive, computed } = Vue;

createApp({
  setup() {
    const state = reactive({
      loading: false,
      error: "",
      accessToken: localStorage.getItem("access_token") || "",
      refreshToken: localStorage.getItem("refresh_token") || "",
      me: {},
      posts: [],
    });

    const form = reactive({
      id: null,
      email: "",
      password: "",
      title: "",
      content: "",
      status: "draft",
      author_id: null,
    });

    const canSetAuthor = computed(() => ["admin", "super_admin"].includes(state.me.role));

    async function api(path, options = {}, retry = true) {
      const headers = Object.assign({ "Content-Type": "application/json" }, options.headers || {});
      if (state.accessToken) {
        headers.Authorization = `Bearer ${state.accessToken}`;
      }
      const response = await fetch(path, { ...options, headers });
      if (response.status === 401 && retry && state.refreshToken) {
        const refreshed = await refresh();
        if (refreshed) {
          return api(path, options, false);
        }
      }
      return response;
    }

    async function login() {
      state.loading = true;
      state.error = "";
      try {
        const response = await fetch("/auth/login", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email: form.email, password: form.password }),
        });
        if (!response.ok) {
          throw new Error("Invalid credentials");
        }
        const payload = await response.json();
        state.accessToken = payload.tokens.access_token;
        state.refreshToken = payload.tokens.refresh_token;
        localStorage.setItem("access_token", state.accessToken);
        localStorage.setItem("refresh_token", state.refreshToken);
        await fetchMe();
        await fetchPosts();
      } catch (err) {
        state.error = err.message || "Login failed";
      } finally {
        state.loading = false;
      }
    }

    async function refresh() {
      if (!state.refreshToken) {
        return false;
      }
      const response = await fetch("/auth/refresh", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: state.refreshToken }),
      });
      if (!response.ok) {
        logout();
        return false;
      }
      const payload = await response.json();
      state.accessToken = payload.access_token;
      state.refreshToken = payload.refresh_token;
      localStorage.setItem("access_token", state.accessToken);
      localStorage.setItem("refresh_token", state.refreshToken);
      return true;
    }

    async function fetchMe() {
      const response = await api("/me");
      if (!response.ok) {
        throw new Error("Session expired");
      }
      state.me = await response.json();
    }

    async function fetchPosts() {
      const response = await api("/posts");
      if (!response.ok) {
        throw new Error("Failed to load posts");
      }
      state.posts = await response.json();
    }

    async function submitPost() {
      state.loading = true;
      state.error = "";
      try {
        const body = {
          title: form.title,
          content: form.content,
          status: form.status,
        };
        if (canSetAuthor.value && form.author_id) {
          body.author_id = Number(form.author_id);
        }

        const path = form.id ? `/posts/${form.id}` : "/posts";
        const method = form.id ? "PUT" : "POST";
        const response = await api(path, { method, body: JSON.stringify(body) });
        if (!response.ok) {
          throw new Error("Failed to save post");
        }
        resetForm();
        await fetchPosts();
      } catch (err) {
        state.error = err.message || "Failed to save";
      } finally {
        state.loading = false;
      }
    }

    function editPost(post) {
      form.id = post.id;
      form.title = post.title;
      form.content = post.content;
      form.status = post.status;
      form.author_id = post.author_id;
    }

    async function deletePost(id) {
      state.error = "";
      const response = await api(`/posts/${id}`, { method: "DELETE" });
      if (!response.ok) {
        state.error = "Failed to delete";
        return;
      }
      await fetchPosts();
    }

    function resetForm() {
      form.id = null;
      form.title = "";
      form.content = "";
      form.status = "draft";
      form.author_id = null;
    }

    async function logout() {
      if (state.refreshToken) {
        await fetch("/auth/logout", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${state.accessToken}`,
          },
          body: JSON.stringify({ refresh_token: state.refreshToken }),
        });
      }
      state.accessToken = "";
      state.refreshToken = "";
      state.me = {};
      state.posts = [];
      localStorage.removeItem("access_token");
      localStorage.removeItem("refresh_token");
    }

    if (state.accessToken) {
      fetchMe().then(fetchPosts).catch(logout);
    }

    return {
      state,
      form,
      canSetAuthor,
      login,
      submitPost,
      editPost,
      deletePost,
      resetForm,
      logout,
    };
  },
}).mount("#app");
