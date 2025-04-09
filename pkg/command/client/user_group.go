package command

import (
	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

var (
	userGroupCmd = &cobra.Command{
		Use:   "group",
		Short: "Group assignments",
		Args:  cobra.NoArgs,
	}
)

func init() {
	userCmd.AddCommand(userGroupCmd)
}

func userGroupPerm(val string) v1.UserGroupPerm {
	res, err := v1.ToUserGroupPerm(val)

	if err != nil {
		return v1.UserGroupPermUser
	}

	return res
}
