package command

import (
	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

var (
	projectGroupCmd = &cobra.Command{
		Use:   "group",
		Short: "Group assignments",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectCmd.AddCommand(projectGroupCmd)
}

func projectGroupPerm(val string) v1.GroupProjectPerm {
	res, err := v1.ToGroupProjectPerm(val)

	if err != nil {
		return v1.GroupProjectPermUser
	}

	return res
}
