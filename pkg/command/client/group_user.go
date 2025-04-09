package command

import (
	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

var (
	groupUserCmd = &cobra.Command{
		Use:   "user",
		Short: "User assignments",
		Args:  cobra.NoArgs,
	}
)

func init() {
	groupCmd.AddCommand(groupUserCmd)
}

func groupUserPerm(val string) v1.UserGroupPerm {
	res, err := v1.ToUserGroupPerm(val)

	if err != nil {
		return v1.UserGroupPermUser
	}

	return res
}
