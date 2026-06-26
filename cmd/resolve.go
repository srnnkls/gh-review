package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/srnnkls/gh-review/internal/api"
	"github.com/srnnkls/gh-review/internal/output"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve <number>",
	Short: "Resolve a review thread",
	Long: `Mark a review thread as resolved.

Identify the thread by a comment ID from 'comments --ids' (--comment) or by
its thread node ID (--thread).`,
	Example: `  gh review resolve 123 -c PRRC_xxx
  gh review resolve 123 --thread PRRT_xxx`,
	Args: cobra.ExactArgs(1),
	RunE: runResolve,
}

var (
	resolveComment string
	resolveThread  string
)

func init() {
	rootCmd.AddCommand(resolveCmd)
	resolveCmd.Flags().StringVarP(&resolveComment, "comment", "c", "", "Comment node ID whose thread to resolve (from 'comments --ids')")
	resolveCmd.Flags().StringVar(&resolveThread, "thread", "", "Thread node ID to resolve")
}

func runResolve(cmd *cobra.Command, args []string) error {
	pr, err := resolvePR(args[0])
	if err != nil {
		return err
	}

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	threadID, err := resolveThreadID(client, pr, resolveThread, resolveComment)
	if err != nil {
		return err
	}

	result, err := client.ResolveThread(threadID)
	if err != nil {
		return err
	}

	formatter, err := output.NewFormatter(outputFormat(), os.Stdout)
	if err != nil {
		return err
	}

	return formatter.Format(output.ResolveResult{
		ThreadID: result.ThreadID,
		Resolved: result.IsResolved,
	})
}
