package constants

type Role int

const (
	ROLE_UNAUTHORIZED   Role = -1 // Unauthorized
	ROLE_ANY_AUTHORIZED Role = 0  // Any Authorized
	ROLE_USER           Role = 1  // User

	ROLE_ADMIN Role = 100 // System Admin

)

var DescOrderedFamilyRoles = []Role{
	ROLE_USER,
	ROLE_ANY_AUTHORIZED,
	ROLE_UNAUTHORIZED,
}

var DescOrderedAdminRoles = []Role{
	ROLE_ADMIN,
	ROLE_ANY_AUTHORIZED,
	ROLE_UNAUTHORIZED,
}
