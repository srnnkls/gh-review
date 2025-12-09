package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/srnnkls/gh-review/internal/api"
	"github.com/srnnkls/gh-review/internal/output"
)

var commentsCmd = &cobra.Command{
	Use:   "comments <number>",
	Short: "List PR comments",
	Long: `List all comments for a pull request.

Shows review comments grouped by author. Use flags to filter and control output.`,
	Example: `  gh review comments 123
  gh review comments 123 --mine --states=pending --ids
  gh review comments 123 --states=changes_requested --tail=10
  gh review comments 123 --author=octocat`,
	Args: cobra.ExactArgs(1),
	RunE: runComments,
}

var (
	listStates     []string
	listAuthor     string
	listMine       bool
	listUnresolved bool
	listTail       int
	listIDs        bool
	listFlat       bool
	listLimit      int
)

func init() {
	rootCmd.AddCommand(commentsCmd)
	commentsCmd.Flags().StringSliceVar(&listStates, "states", nil, "Filter by review state: pending, approved, changes_requested, commented")
	commentsCmd.Flags().StringVarP(&listAuthor, "author", "a", "", "Filter by author username")
	commentsCmd.Flags().BoolVar(&listMine, "mine", false, "Show only my comments (current authenticated user)")
	commentsCmd.Flags().BoolVar(&listUnresolved, "unresolved", false, "Show only unresolved review threads")
	commentsCmd.Flags().IntVar(&listTail, "tail", 0, "Return last N comments (most recent first)")
	commentsCmd.Flags().BoolVar(&listIDs, "ids", false, "Include comment IDs in output")
	commentsCmd.Flags().BoolVar(&listFlat, "flat", false, "Disable author grouping (flat list)")
	commentsCmd.Flags().IntVar(&listLimit, "limit", 100, "Maximum comments to fetch")
}

func runComments(cmd *cobra.Command, args []string) error {
	pr, err := resolvePR(args[0])
	if err != nil {
		return err
	}

	client, err := api.NewClient()
	if err != nil {
		return err
	}

	// Resolve --mine to current user
	if listMine {
		login, err := client.ViewerLogin()
		if err != nil {
			return fmt.Errorf("resolve current user: %w", err)
		}
		listAuthor = login
	}

	var comments []*output.Comment
	var truncated bool

	if listUnresolved {
		// Use reviewThreads query for unresolved comments
		threads, err := client.ReviewThreads(pr, api.ReviewThreadsOptions{
			Limit:          listLimit,
			UnresolvedOnly: true,
		})
		if err != nil {
			return err
		}
		truncated = threads.Truncated

		for _, thread := range threads.Threads {
			for _, c := range thread.Comments {
				cmt := &output.Comment{
					ID:     c.ID,
					Path:   thread.Path,
					Line:   thread.Line,
					Body:   c.Body,
					State:  "unresolved",
					Author: c.Author,
				}
				if matchesFilters(cmt) {
					comments = append(comments, cmt)
				}
			}
		}
	} else {
		// Use standard reviews query
		allComments, err := client.AllPRComments(pr, api.AllCommentsOptions{
			Limit:  listLimit,
			States: listStates,
		})
		if err != nil {
			return err
		}
		truncated = allComments.Truncated

		for _, c := range allComments.ReviewComments {
			cmt := &output.Comment{
				ID:     c.ID,
				Path:   c.Path,
				Line:   c.Line,
				Body:   c.Body,
				State:  c.State,
				Author: c.Author,
			}
			if matchesFilters(cmt) {
				comments = append(comments, cmt)
			}
		}

		for _, c := range allComments.PRComments {
			cmt := &output.Comment{
				ID:     c.ID,
				Body:   c.Body,
				State:  "discussion",
				Author: c.Author,
			}
			if matchesFilters(cmt) {
				comments = append(comments, cmt)
			}
		}
	}

	// Apply --tail limit
	if listTail > 0 && len(comments) > listTail {
		comments = comments[len(comments)-listTail:]
	}

	// Build result
	var result output.CommentsResult
	result.PRRef = pr.String()
	result.IncludeIDs = listIDs

	if listFlat {
		// Flat mode: single group with all comments
		result.Groups = []output.CommentGroup{{Comments: comments}}
	} else {
		// Group by author
		byAuthor := make(map[string][]*output.Comment)
		for _, cmt := range comments {
			byAuthor[cmt.Author] = append(byAuthor[cmt.Author], cmt)
		}

		authors := make([]string, 0, len(byAuthor))
		for author := range byAuthor {
			authors = append(authors, author)
		}
		sort.Strings(authors)

		for _, author := range authors {
			result.Groups = append(result.Groups, output.CommentGroup{
				Author:   author,
				Comments: byAuthor[author],
			})
		}
	}

	formatter, err := output.NewFormatter(outputFormat(), os.Stdout)
	if err != nil {
		return err
	}

	if err := formatter.Format(result); err != nil {
		return err
	}

	if truncated {
		fmt.Fprintf(os.Stderr, "Warning: results may be truncated. Use --limit to fetch more.\n")
	}

	return nil
}

func matchesFilters(c *output.Comment) bool {
	if len(listStates) > 0 {
		matched := false
		for _, s := range listStates {
			if strings.EqualFold(c.State, s) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if listAuthor != "" && !strings.EqualFold(c.Author, listAuthor) {
		return false
	}
	return true
}
