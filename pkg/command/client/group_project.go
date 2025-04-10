package command

import (
	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

var (
	groupProjectCmd = &cobra.Command{
		Use:   "project",
		Short: "Project assignments",
		Args:  cobra.NoArgs,
	}
)

func init() {
	groupCmd.AddCommand(groupProjectCmd)
}

func groupProjectPerm(val string) v1.GroupProjectPerm {
	res, err := v1.ToGroupProjectPerm(val)

	if err != nil {
		return v1.GroupProjectPermUser
	}

	return res
}
