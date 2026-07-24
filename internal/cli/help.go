package cli

import "github.com/spf13/cobra"

func newHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "help",
		Short: "Get help for any command of sneak",
		Long: `Help provides help for any command in sneak.

		Simply type sneak help [path to command] for full details.`,
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()

			if len(args) == 0 {
				return root.Help()
			}

			target, _, err := root.Find(args)
			if err != nil {
				return err
			}

			return target.Help()
		},
	}
}
