package domain

type Role string

const (
	RoleWriter     Role = "writer"
	RoleAdmin      Role = "admin"
	RoleSuperAdmin Role = "super_admin"
)

func (r Role) IsAdminLike() bool {
	return r == RoleAdmin || r == RoleSuperAdmin
}
