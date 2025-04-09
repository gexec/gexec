package command

import (
	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

var (
	projectUserCmd = &cobra.Command{
		Use:   "user",
		Short: "User assignments",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectCmd.AddCommand(projectUserCmd)
}

func projectUserPerm(val string) v1.UserProjectPerm {
	res, err := v1.ToUserProjectPerm(val)

	if err != nil {
		return v1.UserProjectPermUser
	}

	return res
}
