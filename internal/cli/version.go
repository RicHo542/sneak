package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd(info BuildInfo) *cobra.Command {
	var short bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the sneak version",
		Long: `Print version, commit, and build date information for sneak.

Useful for bug reports and confirming which build you're running.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if short {
				fmt.Println(info.Version)
				return nil
			}
			fmt.Printf("sneak version %s\n", info.Version)
			fmt.Printf("  commit: %s\n", info.Commit)
			fmt.Printf("  built:  %s\n", info.Date)
			return nil
		},
	}

	cmd.Flags().BoolVar(&short, "short", false, "print only the version number")

	return cmd
}
