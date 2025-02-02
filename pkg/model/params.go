package model

// ListParams defines optional list attributes.
type ListParams struct {
	Search string
	Sort   string
	Order  string
	Limit  int64
	Offset int64
}

// GroupProjectParams defines parameters for group projects.
type GroupProjectParams struct {
	ListParams

	GroupID   string
	ProjectID string
	Perm      string
}

// UserGroupParams defines parameters for user groups.
type UserGroupParams struct {
	ListParams

	UserID  string
	GroupID string
	Perm    string
}

// UserProjectParams defines parameters for user projects.
type UserProjectParams struct {
	ListParams

	UserID    string
	ProjectID string
	Perm      string
}
