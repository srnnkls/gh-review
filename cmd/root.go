package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/srnnkls/gh-review/internal/api"
	"github.com/srnnkls/gh-review/internal/output"
)

var (
	formatFlag string
	repoFlag   string
)

var rootCmd = &cobra.Command{
	Use:   "gh-review",
	Short: "Manage PR review comments",
	Long: `gh-review manages pull request review comments.

Start a review, add inline comments, edit or delete them, then submit
or discard the entire review.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "table", "Output format: table, plain, json")
	rootCmd.PersistentFlags().StringVarP(&repoFlag, "repo", "R", "", "Select repository using OWNER/REPO format")
}

func outputFormat() output.Format {
	switch formatFlag {
	case "plain":
		return output.FormatPlain
	case "json":
		return output.FormatJSON
	default:
		return output.FormatTable
	}
}

// resolvePR creates a PRRef from the PR argument and -R flag.
// Accepts: number, #number, or full GitHub URL.
func resolvePR(arg string) (*api.PRRef, error) {
	number, repoFromURL, err := api.ParsePRArg(arg)
	if err != nil {
		return nil, err
	}

	// URL-extracted repo takes precedence, then -R flag, then current repo
	repo := repoFromURL
	if repo == "" {
		repo = repoFlag
	}

	return api.NewPRRef(number, repo)
}
