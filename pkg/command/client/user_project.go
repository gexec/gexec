package command

import (
	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

var (
	userProjectCmd = &cobra.Command{
		Use:   "project",
		Short: "Project assignments",
		Args:  cobra.NoArgs,
	}
)

func init() {
	userCmd.AddCommand(userProjectCmd)
}

func userProjectPerm(val string) v1.UserProjectPerm {
	res, err := v1.ToUserProjectPerm(val)

	if err != nil {
		return v1.UserProjectPermUser
	}

	return res
}
