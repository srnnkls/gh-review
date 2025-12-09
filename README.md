# gh-review

A GitHub CLI extension for managing pull request review comments from the command line.

## Overview

`gh-review` streamlines the code review workflow by letting you:

- **Batch review comments** - Add multiple comments to a pending review before submitting
- **Filter and browse** - View comments by state, author, or resolution status
- **Multiple output formats** - Table, plain text, or JSON for scripting
- **Use templates** - Predefined comment templates for common feedback

## Installation

```bash
gh extension install srnnkls/gh-review
```

## Quick Start

```bash
# Add comments to a PR (auto-creates pending review)
gh review add 123 -p src/main.go -l 42 -b "Consider adding error handling here"
gh review add 123 -p src/utils.go -l 15 -b "This could be simplified"

# Review your pending comments
gh review comments 123 --mine

# Submit the review
gh review submit 123 -v approve -b "LGTM!"
```

## Commands

| Command | Description |
|---------|-------------|
| `add` | Add a draft comment to a pending review |
| `view` | View review threads hierarchically |
| `comments` | List PR comments with filtering |
| `edit` | Edit an existing draft comment |
| `delete` | Delete a draft comment |
| `submit` | Submit pending review with verdict |
| `discard` | Discard pending review entirely |

### Global Flags

```
-f, --format <format>    Output format: table, plain, json (default: table)
-R, --repo <owner/repo>  Repository in OWNER/REPO format
```

## Command Reference

### add

Add a draft comment to a pull request. Creates a pending review automatically if none exists.

```bash
gh review add <pr> -p <path> -l <line> -b <body>

# Required
-p, --path <path>     File path to comment on
-l, --line <line>     Line number

# Optional
-b, --body <text>     Comment body
-t, --template <name> Use a predefined template (naming, security, perf, style)
-s, --side <side>     Diff side: LEFT or RIGHT (default: RIGHT)
--start-line <line>   Start line for multi-line comments
--start-side <side>   Start side for multi-line comments
--review-id <id>      Explicit review ID (GraphQL node ID)
```

**Examples:**

```bash
# Simple comment
gh review add 123 -p src/main.go -l 42 -b "Add error handling"

# Using a template
gh review add 123 -p src/main.go -l 42 -t security

# Multi-line comment (lines 10-15)
gh review add 123 -p src/main.go -l 15 --start-line 10 -b "This block needs refactoring"

# Different repository
gh review add 123 -R owner/repo -p file.go -l 5 -b "Comment"
```

### view

View review threads with their comments in a hierarchical structure.

```bash
gh review view <pr> [flags]

--unresolved          Show only unresolved threads
--states <states>     Filter by state: pending, approved, changes_requested, commented
--ids                 Include thread/comment IDs in output
--limit <n>           Maximum threads to fetch (default: 100)
```

**Examples:**

```bash
gh review view 123
gh review view 123 --unresolved
gh review view 123 --states=pending,changes_requested
gh review view 123 --ids --format=json
```

### comments

List all comments for a pull request with flexible filtering.

```bash
gh review comments <pr> [flags]

--states <states>     Filter by state: pending, approved, changes_requested, commented
-a, --author <user>   Filter by author username
--mine                Show only your comments
--unresolved          Show only unresolved threads
--tail <n>            Return last N comments
--ids                 Include comment IDs in output
--flat                Disable author grouping
--limit <n>           Maximum comments to fetch (default: 100)
```

**Examples:**

```bash
gh review comments 123
gh review comments 123 --mine --states=pending
gh review comments 123 --author=octocat --ids
gh review comments 123 --states=changes_requested --tail=10
gh review comments 123 --flat --format=plain
```

### edit

Edit an existing draft comment in your pending review.

```bash
gh review edit <pr> -c <comment-id> -b <body>

-c, --comment <id>    Comment ID (GraphQL node ID, e.g., PRRC_xxx)
-b, --body <text>     New comment body
```

**Example:**

```bash
gh review edit 123 -c PRRC_kwDOABC123 -b "Updated: Please also add tests"
```

### delete

Delete a draft comment from your pending review.

```bash
gh review delete <pr> -c <comment-id>

-c, --comment <id>    Comment ID (GraphQL node ID)
```

**Example:**

```bash
gh review delete 123 -c PRRC_kwDOABC123
```

### submit

Submit your pending review with a verdict.

```bash
gh review submit <pr> -v <verdict> [-b <body>]

-v, --verdict <v>     Review verdict (required):
                        approve         - Approve the PR
                        comment         - General feedback only
                        request_changes - Request changes before merge
-b, --body <text>     Review summary (optional)
--review-id <id>      Explicit review ID
```

**Examples:**

```bash
gh review submit 123 -v approve
gh review submit 123 -v comment -b "Some suggestions for consideration"
gh review submit 123 -v request_changes -b "Please address the security concerns"
```

### discard

Discard your pending review and all its comments. This action cannot be undone.

```bash
gh review discard <pr>

--review-id <id>      Explicit review ID
```

**Example:**

```bash
gh review discard 123
```

## PR Reference Formats

All commands accept PR references in multiple formats:

```bash
gh review view 123                                      # PR number
gh review view #123                                     # With hash prefix
gh review view https://github.com/owner/repo/pull/123  # Full URL
gh review view 123 -R owner/repo                       # With repo flag
```

When using a URL, the repository is extracted automatically. Otherwise, the current repository is detected from your git context.

## Output Formats

### table (default)

Human-readable colored tables, ideal for terminal use.

### plain

Tab-separated values, useful for scripting and piping to other tools.

```bash
gh review comments 123 --format=plain | cut -f2
```

### json

Structured JSON output for programmatic consumption.

```bash
gh review comments 123 --format=json | jq '.groups[].comments[].body'
```

## Comment Templates

Use predefined templates with the `-t` flag when adding comments:

| Template | Description |
|----------|-------------|
| `naming` | Naming convention feedback |
| `security` | Security concern alert |
| `perf` | Performance consideration |
| `style` | Style guide violation |

```bash
gh review add 123 -p src/auth.go -l 42 -t security
```

## Development

### Building from Source

```bash
git clone https://github.com/srnnkls/gh-review
cd gh-review
go build -o gh-review .
```

### Running Tests

```bash
go test ./...
```

### Test Coverage

```bash
go test ./... -cover
```

## License

MIT
