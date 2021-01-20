package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCompletionCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate completion script",
		Long:                  "Generate completion script for various shells",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			switch args[0] {
			case "bash":
				err = cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				err = cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				err = cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				err = cmd.Root().GenPowerShellCompletion(cmd.OutOrStdout())
			}
			if err != nil {
				return fmt.Errorf("error while generating completion script: %s", err.Error())
			}
			return nil
		},
	}
	return cmd
}
