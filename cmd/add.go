package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/srnnkls/gh-review/internal/api"
	"github.com/srnnkls/gh-review/internal/output"
	"github.com/srnnkls/gh-review/internal/templates"
)

var addCmd = &cobra.Command{
	Use:   "add <number>",
	Short: "Add a draft comment",
	Long: `Add a comment to your pending review.

Creates a new pending review if none exists.`,
	Example: `  gh review add 123 -p src/main.go -l 42 -b "Consider error handling"
  gh review add 123 -R owner/repo -p src/main.go -l 42 -t naming
  gh review add 123 -p src/main.go -l 50 --start-line 45 -b "Multi-line comment"`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

var (
	addPath      string
	addLine      int
	addBody      string
	addSide      string
	addTemplate  string
	addStartLine int
	addStartSide string
	addReviewID  string
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&addPath, "path", "p", "", "File path (required)")
	addCmd.Flags().IntVarP(&addLine, "line", "l", 0, "Line number (required)")
	addCmd.Flags().StringVarP(&addBody, "body", "b", "", "Comment body")
	addCmd.Flags().StringVarP(&addSide, "side", "s", "RIGHT", "Diff side: LEFT or RIGHT")
	addCmd.Flags().StringVarP(&addTemplate, "template", "t", "", "Use a predefined template")
	addCmd.Flags().IntVar(&addStartLine, "start-line", 0, "Start line for multi-line comment")
	addCmd.Flags().StringVar(&addStartSide, "start-side", "", "Start side for multi-line comment")
	addCmd.Flags().StringVar(&addReviewID, "review-id", "", "Explicit review ID (GraphQL node ID)")

	addCmd.MarkFlagRequired("path")
	addCmd.MarkFlagRequired("line")
}

func runAdd(cmd *cobra.Command, args []string) error {
	pr, err := resolvePR(args[0])
	if err != nil {
		return err
	}

	body := addBody
	if addTemplate != "" {
		tpl, ok := templates.Get(addTemplate)
		if !ok {
			return fmt.Errorf("unknown template %q (available: %v)", addTemplate, templates.List())
		}
		body = tpl
	}
	if body == "" {
		return fmt.Errorf("body is required (use -b or -t)")
	}

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	var reviewID string

	if addReviewID != "" {
		reviewID = addReviewID
	} else {
		review, err := client.LatestPendingReview(pr, api.PendingReviewsOptions{})
		if err != nil {
			prIdentity, resolveErr := client.ResolvePR(pr)
			if resolveErr != nil {
				return resolveErr
			}
			result, createErr := client.CreateReview(api.CreateReviewInput{
				PRNodeID:  prIdentity.NodeID,
				CommitOID: prIdentity.HeadRefOID,
			})
			if createErr != nil {
				return createErr
			}
			reviewID = result.ID
		} else {
			reviewID = review.ID
		}
	}

	input := api.AddThreadInput{
		ReviewID: reviewID,
		Path:     addPath,
		Line:     addLine,
		Side:     addSide,
		Body:     body,
	}
	if addStartLine > 0 {
		input.StartLine = &addStartLine
	}
	if addStartSide != "" {
		input.StartSide = &addStartSide
	}

	_, err = client.AddThread(input)
	if err != nil {
		return err
	}

	result := output.AddResult{
		Path: addPath,
		Line: addLine,
	}

	formatter, err := output.NewFormatter(outputFormat(), os.Stdout)
	if err != nil {
		return err
	}

	return formatter.Format(result)
}
