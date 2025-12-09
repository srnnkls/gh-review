package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/srnnkls/gh-review/internal/api"
	"github.com/srnnkls/gh-review/internal/output"
)

var editCmd = &cobra.Command{
	Use:   "edit <number>",
	Short: "Edit a draft comment",
	Long:  `Edit an existing comment in your pending review.`,
	Example: `  gh review edit 123 -c PRRC_xxx -b "Updated comment body"
  gh review edit 123 -R owner/repo -c PRRC_xxx -b "Updated"`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

var (
	editCommentID string
	editBody      string
)

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.Flags().StringVarP(&editCommentID, "comment", "c", "", "Comment ID (GraphQL node ID, required)")
	editCmd.Flags().StringVarP(&editBody, "body", "b", "", "New comment body (required)")

	editCmd.MarkFlagRequired("comment")
	editCmd.MarkFlagRequired("body")
}

func runEdit(cmd *cobra.Command, args []string) error {
	_, err := resolvePR(args[0])
	if err != nil {
		return err
	}

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	err = client.UpdateComment(api.UpdateCommentInput{
		CommentID: editCommentID,
		Body:      editBody,
	})
	if err != nil {
		return err
	}

	result := output.EditResult{
		CommentID: editCommentID,
	}

	formatter, err := output.NewFormatter(outputFormat(), os.Stdout)
	if err != nil {
		return err
	}

	return formatter.Format(result)
}
