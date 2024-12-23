package model

// ListParams defines optional list attributes.
type ListParams struct {
	Search string
	Sort   string
	Order  string
	Limit  int64
	Offset int64
}

// TeamProjectParams defines parameters for team projects.
type TeamProjectParams struct {
	ListParams

	TeamID    string
	ProjectID string
	Perm      string
}

// UserTeamParams defines parameters for user teams.
type UserTeamParams struct {
	ListParams

	UserID string
	TeamID string
	Perm   string
}

// UserProjectParams defines parameters for user projects.
type UserProjectParams struct {
	ListParams

	UserID    string
	ProjectID string
	Perm      string
}
