package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/srnnkls/gh-review/internal/api"
	"github.com/srnnkls/gh-review/internal/output"
)

var submitCmd = &cobra.Command{
	Use:   "submit <number>",
	Short: "Submit a pending review",
	Long: `Submit your pending review with a verdict.

Available verdicts:
  approve  - Approve the pull request
  comment  - Submit general feedback
  request_changes - Request changes before merge`,
	Example: `  gh review submit 123 -v approve
  gh review submit 123 -R owner/repo -v comment -b "Looks good overall"
  gh review submit 123 -v request_changes -b "Please fix the issues"`,
	Args: cobra.ExactArgs(1),
	RunE: runSubmit,
}

var (
	submitVerdict  string
	submitBody     string
	submitReviewID string
)

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().StringVarP(&submitVerdict, "verdict", "v", "", "Review verdict: approve, comment, request_changes (required)")
	submitCmd.Flags().StringVarP(&submitBody, "body", "b", "", "Review body/summary")
	submitCmd.Flags().StringVar(&submitReviewID, "review-id", "", "Explicit review ID (GraphQL node ID)")

	submitCmd.MarkFlagRequired("verdict")
}

func runSubmit(cmd *cobra.Command, args []string) error {
	pr, err := resolvePR(args[0])
	if err != nil {
		return err
	}

	event := strings.ToUpper(strings.TrimSpace(submitVerdict))
	switch event {
	case "APPROVE", "COMMENT", "REQUEST_CHANGES":
		// valid
	default:
		return fmt.Errorf("invalid verdict %q: use approve, comment, or request_changes", submitVerdict)
	}

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	var reviewID string

	if submitReviewID != "" {
		reviewID = submitReviewID
	} else {
		review, err := client.LatestPendingReview(pr, api.PendingReviewsOptions{})
		if err != nil {
			return err
		}
		reviewID = review.ID
	}

	err = client.SubmitReview(api.SubmitReviewInput{
		ReviewID: reviewID,
		Event:    event,
		Body:     submitBody,
	})
	if err != nil {
		return err
	}

	result := output.SubmitResult{
		Verdict: strings.ToLower(event),
	}

	formatter, err := output.NewFormatter(outputFormat(), os.Stdout)
	if err != nil {
		return err
	}

	return formatter.Format(result)
}
