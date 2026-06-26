package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/srnnkls/gh-review/internal/api"
	"github.com/srnnkls/gh-review/internal/output"
)

var replyCmd = &cobra.Command{
	Use:   "reply <number>",
	Short: "Reply to a review thread",
	Long: `Post a reply to an existing review thread.

Identify the thread by a comment ID from 'comments --ids' (--comment) or by
its thread node ID (--thread).`,
	Example: `  gh review reply 123 -c PRRC_xxx -b "Done in abc1234"
  gh review reply 123 --thread PRRT_xxx -b "Fixed, thanks"`,
	Args: cobra.ExactArgs(1),
	RunE: runReply,
}

var (
	replyComment string
	replyThread  string
	replyBody    string
)

func init() {
	rootCmd.AddCommand(replyCmd)
	replyCmd.Flags().StringVarP(&replyComment, "comment", "c", "", "Comment node ID to reply under (from 'comments --ids')")
	replyCmd.Flags().StringVar(&replyThread, "thread", "", "Thread node ID to reply to")
	replyCmd.Flags().StringVarP(&replyBody, "body", "b", "", "Reply body (required)")

	replyCmd.MarkFlagRequired("body")
}

func runReply(cmd *cobra.Command, args []string) error {
	pr, err := resolvePR(args[0])
	if err != nil {
		return err
	}

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	threadID, err := resolveThreadID(client, pr, replyThread, replyComment)
	if err != nil {
		return err
	}

	result, err := client.ReplyThread(api.ReplyThreadInput{
		ThreadID: threadID,
		Body:     replyBody,
	})
	if err != nil {
		return err
	}

	formatter, err := output.NewFormatter(outputFormat(), os.Stdout)
	if err != nil {
		return err
	}

	return formatter.Format(output.ReplyResult{
		ThreadID:  threadID,
		CommentID: result.ID,
		URL:       result.URL,
	})
}
