package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/srnnkls/gh-review/internal/api"
	"github.com/srnnkls/gh-review/internal/output"
)

var viewCmd = &cobra.Command{
	Use:   "view <number>",
	Short: "View PR review threads",
	Long: `View review threads for a pull request.

Shows threads with their comments in hierarchical structure.`,
	Example: `  gh review view 123
  gh review view 123 --unresolved
  gh review view 123 --states=pending,changes_requested
  gh review view 123 --ids`,
	Args: cobra.ExactArgs(1),
	RunE: runView,
}

var (
	viewUnresolved bool
	viewIDs        bool
	viewLimit      int
	viewStates     []string
)

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().BoolVar(&viewUnresolved, "unresolved", false, "Show only unresolved threads")
	viewCmd.Flags().BoolVar(&viewIDs, "ids", false, "Include thread/comment IDs in output")
	viewCmd.Flags().IntVar(&viewLimit, "limit", 100, "Maximum threads to fetch")
	viewCmd.Flags().StringSliceVar(&viewStates, "states", nil, "Filter by review state: pending, approved, changes_requested, commented")
}

func runView(cmd *cobra.Command, args []string) error {
	pr, err := resolvePR(args[0])
	if err != nil {
		return err
	}

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	threads, err := client.ReviewThreads(pr, api.ReviewThreadsOptions{
		Limit:          viewLimit,
		UnresolvedOnly: viewUnresolved,
		States:         viewStates,
	})
	if err != nil {
		return err
	}

	result := output.ViewResult{
		PRRef:      pr.String(),
		IncludeIDs: viewIDs,
	}

	for _, t := range threads.Threads {
		thread := output.ViewThread{
			ID:       t.ID,
			Path:     t.Path,
			Line:     t.Line,
			Resolved: t.IsResolved,
		}

		for _, c := range t.Comments {
			thread.Comments = append(thread.Comments, output.ViewThreadComment{
				ID:     c.ID,
				Author: c.Author,
				Body:   c.Body,
			})
		}

		result.Threads = append(result.Threads, thread)
	}

	formatter, err := output.NewFormatter(outputFormat(), os.Stdout)
	if err != nil {
		return err
	}

	if err := formatter.Format(result); err != nil {
		return err
	}

	if threads.Truncated {
		fmt.Fprintf(os.Stderr, "Warning: results may be truncated. Use --limit to fetch more.\n")
	}

	return nil
}
