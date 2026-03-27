CREATE TABLE IF NOT EXISTS media (
    id BIGSERIAL PRIMARY KEY,
    original_name VARCHAR(255) NOT NULL,
    stored_name VARCHAR(255) NOT NULL UNIQUE,
    content_type VARCHAR(64) NOT NULL,
    size_bytes BIGINT NOT NULL,
    uploader_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_media_uploader_id ON media(uploader_id);

CREATE TABLE IF NOT EXISTS post_images (
    post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    media_id BIGINT NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    position INT NOT NULL DEFAULT 0,
    PRIMARY KEY (post_id, media_id)
);

CREATE INDEX IF NOT EXISTS idx_post_images_media_id ON post_images(media_id);
