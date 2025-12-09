package api

import (
	"fmt"
	"strings"
	"time"
)

type PendingReview struct {
	ID         string
	State      string
	URL        string
	UpdatedAt  time.Time
	Author     string
	AuthorID   int64
	Comments   []*ReviewComment
	TotalCount int
}

type ReviewComment struct {
	ID        string
	Path      string
	Line      int
	StartLine *int
	Body      string
	Side      string
	StartSide *string
	Outdated  bool
	Author    string
}

type PRComment struct {
	ID        string
	Body      string
	Author    string
	CreatedAt time.Time
}

type AllCommentsResult struct {
	ReviewComments []*ReviewCommentWithState
	PRComments     []*PRComment
	Truncated      bool
}

type ReviewCommentWithState struct {
	ReviewComment
	State string
}

type PendingReviewsOptions struct {
	Reviewer string
	First    int
}

func (c *Client) PendingReviews(pr *PRRef, opts PendingReviewsOptions) ([]*PendingReview, error) {
	first := opts.First
	if first <= 0 {
		first = 20
	}

	reviewer := strings.TrimSpace(opts.Reviewer)
	if reviewer == "" {
		login, err := c.ViewerLogin()
		if err != nil {
			return nil, err
		}
		reviewer = login
	}

	const query = `query PendingReviews($owner: String!, $name: String!, $number: Int!, $first: Int!) {
  repository(owner: $owner, name: $name) {
    pullRequest(number: $number) {
      reviews(states: [PENDING], first: $first) {
        nodes {
          id
          state
          url
          updatedAt
          author {
            login
          }
          comments(first: 100) {
            totalCount
            nodes {
              id
              path
              line
              startLine
              body
              outdated
              originalLine
            }
          }
        }
      }
    }
  }
}`

	variables := map[string]interface{}{
		"owner":  pr.Owner,
		"name":   pr.Repo,
		"number": pr.Number,
		"first":  first,
	}

	var response struct {
		Repository struct {
			PullRequest struct {
				Reviews struct {
					Nodes []struct {
						ID        string `json:"id"`
						State     string `json:"state"`
						URL       string `json:"url"`
						UpdatedAt string `json:"updatedAt"`
						Author    struct {
							Login string `json:"login"`
						} `json:"author"`
						Comments struct {
							TotalCount int `json:"totalCount"`
							Nodes      []struct {
								ID           string `json:"id"`
								Path         string `json:"path"`
								Line         *int   `json:"line"`
								StartLine    *int   `json:"startLine"`
								Body         string `json:"body"`
								Outdated     bool   `json:"outdated"`
								OriginalLine *int   `json:"originalLine"`
							} `json:"nodes"`
						} `json:"comments"`
					} `json:"nodes"`
				} `json:"reviews"`
			} `json:"pullRequest"`
		} `json:"repository"`
	}

	if err := c.gql.Do(query, variables, &response); err != nil {
		return nil, fmt.Errorf("query pending reviews: %w", err)
	}

	var results []*PendingReview
	for _, node := range response.Repository.PullRequest.Reviews.Nodes {
		authorLogin := strings.TrimSpace(node.Author.Login)
		if !strings.EqualFold(authorLogin, reviewer) {
			continue
		}

		id := strings.TrimSpace(node.ID)
		if id == "" {
			continue
		}

		updatedAt, _ := time.Parse(time.RFC3339, node.UpdatedAt)

		comments := make([]*ReviewComment, 0, len(node.Comments.Nodes))
		for _, cmt := range node.Comments.Nodes {
			cmtID := strings.TrimSpace(cmt.ID)
			if cmtID == "" {
				continue
			}

			line := 0
			if cmt.Line != nil {
				line = *cmt.Line
			} else if cmt.OriginalLine != nil {
				line = *cmt.OriginalLine
			}

			comments = append(comments, &ReviewComment{
				ID:        cmtID,
				Path:      cmt.Path,
				Line:      line,
				StartLine: cmt.StartLine,
				Body:      cmt.Body,
				Outdated:  cmt.Outdated,
			})
		}

		results = append(results, &PendingReview{
			ID:         id,
			State:      strings.ToUpper(node.State),
			URL:        node.URL,
			UpdatedAt:  updatedAt,
			Author:     authorLogin,
			Comments:   comments,
			TotalCount: node.Comments.TotalCount,
		})
	}

	return results, nil
}

func (c *Client) LatestPendingReview(pr *PRRef, opts PendingReviewsOptions) (*PendingReview, error) {
	reviews, err := c.PendingReviews(pr, opts)
	if err != nil {
		return nil, err
	}

	if len(reviews) == 0 {
		return nil, fmt.Errorf("no pending review found")
	}

	latest := reviews[0]
	for _, r := range reviews[1:] {
		if r.UpdatedAt.After(latest.UpdatedAt) {
			latest = r
		}
	}

	return latest, nil
}

type AllCommentsOptions struct {
	Limit  int
	States []string
}

func (c *Client) AllPRComments(pr *PRRef, opts AllCommentsOptions) (*AllCommentsResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}

	// Build query dynamically based on whether states filter is provided
	var query string
	variables := map[string]interface{}{
		"owner":  pr.Owner,
		"name":   pr.Repo,
		"number": pr.Number,
		"limit":  limit,
	}

	if len(opts.States) > 0 {
		query = `query AllPRComments($owner: String!, $name: String!, $number: Int!, $limit: Int!, $states: [PullRequestReviewState!]) {
  repository(owner: $owner, name: $name) {
    pullRequest(number: $number) {
      reviews(states: $states, first: $limit) {
        totalCount
        nodes {
          id
          state
          author { login }
          comments(first: 100) {
            nodes {
              id
              path
              line
              startLine
              body
              outdated
              originalLine
              author { login }
            }
          }
        }
      }
      comments(first: $limit) {
        totalCount
        nodes {
          id
          body
          author { login }
          createdAt
        }
      }
    }
  }
}`
		// Convert states to uppercase for GraphQL enum
		gqlStates := make([]string, len(opts.States))
		for i, s := range opts.States {
			gqlStates[i] = strings.ToUpper(s)
		}
		variables["states"] = gqlStates
	} else {
		query = `query AllPRComments($owner: String!, $name: String!, $number: Int!, $limit: Int!) {
  repository(owner: $owner, name: $name) {
    pullRequest(number: $number) {
      reviews(first: $limit) {
        totalCount
        nodes {
          id
          state
          author { login }
          comments(first: 100) {
            nodes {
              id
              path
              line
              startLine
              body
              outdated
              originalLine
              author { login }
            }
          }
        }
      }
      comments(first: $limit) {
        totalCount
        nodes {
          id
          body
          author { login }
          createdAt
        }
      }
    }
  }
}`
	}

	var response struct {
		Repository struct {
			PullRequest struct {
				Reviews struct {
					TotalCount int `json:"totalCount"`
					Nodes      []struct {
						ID     string `json:"id"`
						State  string `json:"state"`
						Author struct {
							Login string `json:"login"`
						} `json:"author"`
						Comments struct {
							Nodes []struct {
								ID           string `json:"id"`
								Path         string `json:"path"`
								Line         *int   `json:"line"`
								StartLine    *int   `json:"startLine"`
								Body         string `json:"body"`
								Outdated     bool   `json:"outdated"`
								OriginalLine *int   `json:"originalLine"`
								Author       struct {
									Login string `json:"login"`
								} `json:"author"`
							} `json:"nodes"`
						} `json:"comments"`
					} `json:"nodes"`
				} `json:"reviews"`
				Comments struct {
					TotalCount int `json:"totalCount"`
					Nodes      []struct {
						ID     string `json:"id"`
						Body   string `json:"body"`
						Author struct {
							Login string `json:"login"`
						} `json:"author"`
						CreatedAt string `json:"createdAt"`
					} `json:"nodes"`
				} `json:"comments"`
			} `json:"pullRequest"`
		} `json:"repository"`
	}

	if err := c.gql.Do(query, variables, &response); err != nil {
		return nil, fmt.Errorf("query all PR comments: %w", err)
	}

	result := &AllCommentsResult{}

	for _, review := range response.Repository.PullRequest.Reviews.Nodes {
		reviewState := normalizeReviewState(review.State)
		reviewAuthor := strings.TrimSpace(review.Author.Login)

		for _, cmt := range review.Comments.Nodes {
			cmtID := strings.TrimSpace(cmt.ID)
			if cmtID == "" {
				continue
			}

			line := 0
			if cmt.Line != nil {
				line = *cmt.Line
			} else if cmt.OriginalLine != nil {
				line = *cmt.OriginalLine
			}

			author := strings.TrimSpace(cmt.Author.Login)
			if author == "" {
				author = reviewAuthor
			}

			result.ReviewComments = append(result.ReviewComments, &ReviewCommentWithState{
				ReviewComment: ReviewComment{
					ID:        cmtID,
					Path:      cmt.Path,
					Line:      line,
					StartLine: cmt.StartLine,
					Body:      cmt.Body,
					Outdated:  cmt.Outdated,
					Author:    author,
				},
				State: reviewState,
			})
		}
	}

	for _, cmt := range response.Repository.PullRequest.Comments.Nodes {
		cmtID := strings.TrimSpace(cmt.ID)
		if cmtID == "" {
			continue
		}

		createdAt, _ := time.Parse(time.RFC3339, cmt.CreatedAt)

		result.PRComments = append(result.PRComments, &PRComment{
			ID:        cmtID,
			Body:      cmt.Body,
			Author:    strings.TrimSpace(cmt.Author.Login),
			CreatedAt: createdAt,
		})
	}

	reviewCount := response.Repository.PullRequest.Reviews.TotalCount
	prCommentCount := response.Repository.PullRequest.Comments.TotalCount
	result.Truncated = reviewCount > limit || prCommentCount > limit

	return result, nil
}

func normalizeReviewState(state string) string {
	switch strings.ToLower(state) {
	case "pending":
		return "pending"
	case "approved":
		return "approved"
	case "changes_requested":
		return "changes_requested"
	case "commented":
		return "commented"
	default:
		return "submitted"
	}
}

type ThreadComment struct {
	ID     string
	Body   string
	Author string
}

type Thread struct {
	ID         string
	Path       string
	Line       int
	IsResolved bool
	State      string
	Comments   []*ThreadComment
}

type ThreadsResult struct {
	Threads   []*Thread
	Truncated bool
}

type ReviewThreadsOptions struct {
	Limit          int
	UnresolvedOnly bool
	States         []string
}

func (c *Client) ReviewThreads(pr *PRRef, opts ReviewThreadsOptions) (*ThreadsResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}

	const query = `query ReviewThreads($owner: String!, $name: String!, $number: Int!, $limit: Int!) {
  repository(owner: $owner, name: $name) {
    pullRequest(number: $number) {
      reviewThreads(first: $limit) {
        totalCount
        nodes {
          id
          isResolved
          path
          line
          originalLine
          comments(first: 50) {
            nodes {
              id
              body
              author { login }
              pullRequestReview { state }
            }
          }
        }
      }
    }
  }
}`

	variables := map[string]interface{}{
		"owner":  pr.Owner,
		"name":   pr.Repo,
		"number": pr.Number,
		"limit":  limit,
	}

	var response struct {
		Repository struct {
			PullRequest struct {
				ReviewThreads struct {
					TotalCount int `json:"totalCount"`
					Nodes      []struct {
						ID           string `json:"id"`
						IsResolved   bool   `json:"isResolved"`
						Path         string `json:"path"`
						Line         *int   `json:"line"`
						OriginalLine *int   `json:"originalLine"`
						Comments     struct {
							Nodes []struct {
								ID     string `json:"id"`
								Body   string `json:"body"`
								Author struct {
									Login string `json:"login"`
								} `json:"author"`
								PullRequestReview struct {
									State string `json:"state"`
								} `json:"pullRequestReview"`
							} `json:"nodes"`
						} `json:"comments"`
					} `json:"nodes"`
				} `json:"reviewThreads"`
			} `json:"pullRequest"`
		} `json:"repository"`
	}

	if err := c.gql.Do(query, variables, &response); err != nil {
		return nil, fmt.Errorf("query review threads: %w", err)
	}

	result := &ThreadsResult{}

	for _, thread := range response.Repository.PullRequest.ReviewThreads.Nodes {
		if opts.UnresolvedOnly && thread.IsResolved {
			continue
		}

		threadID := strings.TrimSpace(thread.ID)
		if threadID == "" {
			continue
		}

		// Get state from first comment's review
		var threadState string
		if len(thread.Comments.Nodes) > 0 {
			threadState = normalizeReviewState(thread.Comments.Nodes[0].PullRequestReview.State)
		}

		// Filter by states if specified
		if len(opts.States) > 0 {
			matched := false
			for _, s := range opts.States {
				if strings.EqualFold(threadState, s) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		line := 0
		if thread.Line != nil {
			line = *thread.Line
		} else if thread.OriginalLine != nil {
			line = *thread.OriginalLine
		}

		comments := make([]*ThreadComment, 0, len(thread.Comments.Nodes))
		for _, cmt := range thread.Comments.Nodes {
			cmtID := strings.TrimSpace(cmt.ID)
			if cmtID == "" {
				continue
			}

			comments = append(comments, &ThreadComment{
				ID:     cmtID,
				Body:   cmt.Body,
				Author: strings.TrimSpace(cmt.Author.Login),
			})
		}

		result.Threads = append(result.Threads, &Thread{
			ID:         threadID,
			Path:       thread.Path,
			Line:       line,
			IsResolved: thread.IsResolved,
			State:      threadState,
			Comments:   comments,
		})
	}

	result.Truncated = response.Repository.PullRequest.ReviewThreads.TotalCount > limit

	return result, nil
}
