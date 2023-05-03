package main

import (
	"io"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
)

const ecnetDescription = `
This command consists of multiple subcommands related to managing instances of
ecnet installations. Each installation receives a unique ecnet name.
`

func newEcnetCmd(config *action.Configuration, _ io.Reader, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ecnet",
		Short: "manage ecnet installations",
		Long:  ecnetDescription,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newEcentList(out))

	if !settings.IsManaged() {
		cmd.AddCommand(newEcnetUpgradeCmd(config, out))
	}

	return cmd
}
