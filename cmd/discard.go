package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/srnnkls/gh-review/internal/api"
	"github.com/srnnkls/gh-review/internal/output"
)

var discardCmd = &cobra.Command{
	Use:   "discard <number>",
	Short: "Discard a pending review",
	Long: `Discard your pending review and all its comments.

This action cannot be undone.`,
	Example: `  gh review discard 123
  gh review discard 123 -R owner/repo`,
	Args: cobra.ExactArgs(1),
	RunE: runDiscard,
}

var discardReviewID string

func init() {
	rootCmd.AddCommand(discardCmd)
	discardCmd.Flags().StringVar(&discardReviewID, "review-id", "", "Explicit review ID (GraphQL node ID)")
}

func runDiscard(cmd *cobra.Command, args []string) error {
	pr, err := resolvePR(args[0])
	if err != nil {
		return err
	}

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	var reviewID string

	if discardReviewID != "" {
		reviewID = discardReviewID
	} else {
		review, err := client.LatestPendingReview(pr, api.PendingReviewsOptions{})
		if err != nil {
			return err
		}
		reviewID = review.ID
	}

	err = client.DeleteReview(reviewID)
	if err != nil {
		return err
	}

	result := output.DiscardResult{
		ReviewID: reviewID,
	}

	formatter, err := output.NewFormatter(outputFormat(), os.Stdout)
	if err != nil {
		return err
	}

	return formatter.Format(result)
}
