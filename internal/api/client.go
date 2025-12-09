package api

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
)

// GraphQLClient is the interface for GraphQL operations
type GraphQLClient interface {
	Do(query string, variables map[string]interface{}, response interface{}) error
}

type Client struct {
	gql GraphQLClient
}

func NewClient() (*Client, error) {
	gql, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("create GraphQL client: %w", err)
	}
	return &Client{gql: gql}, nil
}

type PRRef struct {
	Owner  string
	Repo   string
	Number int
}

func (pr *PRRef) String() string {
	return fmt.Sprintf("%s/%s#%d", pr.Owner, pr.Repo, pr.Number)
}

// NewPRRef creates a PRRef from number and optional repo string.
// If repo is empty, uses current repository from git context.
func NewPRRef(number int, repo string) (*PRRef, error) {
	if number <= 0 {
		return nil, fmt.Errorf("PR number must be positive")
	}

	var owner, name string

	if repo == "" {
		current, err := repository.Current()
		if err != nil {
			return nil, fmt.Errorf("could not determine repository: %w (use -R owner/repo)", err)
		}
		owner = current.Owner
		name = current.Name
	} else {
		parsed, err := repository.Parse(repo)
		if err != nil {
			return nil, fmt.Errorf("invalid repository %q: %w", repo, err)
		}
		owner = parsed.Owner
		name = parsed.Name
	}

	return &PRRef{
		Owner:  owner,
		Repo:   name,
		Number: number,
	}, nil
}

// prURLPattern matches GitHub PR URLs: github.com/owner/repo/pull/123
var prURLPattern = regexp.MustCompile(`(?:https?://)?(?:www\.)?github\.com/([^/]+)/([^/]+)/pull/(\d+)`)

// ParsePRArg parses a PR reference from string argument.
// Accepts: number, #number, or full GitHub URL.
// Returns the PR number and optionally extracted repo info.
func ParsePRArg(arg string) (number int, repoOverride string, err error) {
	arg = strings.TrimSpace(arg)

	// Check if it's a URL with /pull/N
	if matches := prURLPattern.FindStringSubmatch(arg); matches != nil {
		number, _ = strconv.Atoi(matches[3])
		return number, fmt.Sprintf("%s/%s", matches[1], matches[2]), nil
	}

	// Otherwise treat as a number
	arg = strings.TrimPrefix(arg, "#")
	number, err = strconv.Atoi(arg)
	if err != nil {
		return 0, "", fmt.Errorf("invalid PR reference %q: expected number or URL", arg)
	}
	if number <= 0 {
		return 0, "", fmt.Errorf("PR number must be positive")
	}
	return number, "", nil
}

type PRIdentity struct {
	NodeID     string
	HeadRefOID string
}

func (c *Client) ResolvePR(pr *PRRef) (*PRIdentity, error) {
	const query = `query ResolvePR($owner: String!, $name: String!, $number: Int!) {
  repository(owner: $owner, name: $name) {
    pullRequest(number: $number) {
      id
      headRefOid
    }
  }
}`

	variables := map[string]interface{}{
		"owner":  pr.Owner,
		"name":   pr.Repo,
		"number": pr.Number,
	}

	var response struct {
		Repository struct {
			PullRequest struct {
				ID         string `json:"id"`
				HeadRefOID string `json:"headRefOid"`
			} `json:"pullRequest"`
		} `json:"repository"`
	}

	if err := c.gql.Do(query, variables, &response); err != nil {
		return nil, fmt.Errorf("resolve PR: %w", err)
	}

	nodeID := strings.TrimSpace(response.Repository.PullRequest.ID)
	headOID := strings.TrimSpace(response.Repository.PullRequest.HeadRefOID)
	if nodeID == "" || headOID == "" {
		return nil, fmt.Errorf("PR %s not found or missing metadata", pr)
	}

	return &PRIdentity{NodeID: nodeID, HeadRefOID: headOID}, nil
}

func (c *Client) ViewerLogin() (string, error) {
	const query = `query ViewerLogin { viewer { login } }`

	var response struct {
		Viewer struct {
			Login string `json:"login"`
		} `json:"viewer"`
	}

	if err := c.gql.Do(query, nil, &response); err != nil {
		return "", fmt.Errorf("get viewer login: %w", err)
	}

	login := strings.TrimSpace(response.Viewer.Login)
	if login == "" {
		return "", fmt.Errorf("viewer login unavailable")
	}

	return login, nil
}
