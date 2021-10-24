package dbaseModels

// gen:
type FunctionalRolePermission struct {
	ID               uint `gen:"uindex"`
	FunctionalRoleID uint `gen:"index"`
	PermissionID     uint
}
