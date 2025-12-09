package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/srnnkls/gh-review/internal/api"
	"github.com/srnnkls/gh-review/internal/output"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <number>",
	Short: "Delete a draft comment",
	Long:  `Delete a comment from your pending review.`,
	Example: `  gh review delete 123 -c PRRC_xxx
  gh review delete 123 -R owner/repo -c PRRC_xxx`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

var deleteCommentID string

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVarP(&deleteCommentID, "comment", "c", "", "Comment ID (GraphQL node ID, required)")

	deleteCmd.MarkFlagRequired("comment")
}

func runDelete(cmd *cobra.Command, args []string) error {
	_, err := resolvePR(args[0])
	if err != nil {
		return err
	}

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	err = client.DeleteComment(deleteCommentID)
	if err != nil {
		return err
	}

	result := output.DeleteResult{
		CommentID: deleteCommentID,
	}

	formatter, err := output.NewFormatter(outputFormat(), os.Stdout)
	if err != nil {
		return err
	}

	return formatter.Format(result)
}
