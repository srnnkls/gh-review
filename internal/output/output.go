package output

import (
	"fmt"
	"io"
)

type Format string

const (
	FormatTable Format = "table"
	FormatPlain Format = "plain"
	FormatJSON  Format = "json"
)

type Formatter interface {
	Format(result Result) error
}

type Result interface {
	Type() string
}

type Comment struct {
	ID     string
	Path   string
	Line   int
	Body   string
	State  string
	Author string
}

type CommentGroup struct {
	Author   string
	Comments []*Comment
}

type CommentsResult struct {
	PRRef      string
	Groups     []CommentGroup
	IncludeIDs bool
}

func (r CommentsResult) Type() string { return "comments" }

type ViewThread struct {
	ID         string
	Path       string
	Line       int
	Resolved   bool
	Comments   []ViewThreadComment
}

type ViewThreadComment struct {
	ID     string
	Author string
	Body   string
}

type ViewResult struct {
	PRRef      string
	Threads    []ViewThread
	IncludeIDs bool
}

func (r ViewResult) Type() string { return "view" }

type AddResult struct {
	Path string
	Line int
}

func (r AddResult) Type() string { return "add" }

type EditResult struct {
	CommentID string
}

func (r EditResult) Type() string { return "edit" }

type DeleteResult struct {
	CommentID string
}

func (r DeleteResult) Type() string { return "delete" }

type SubmitResult struct {
	Verdict string
}

func (r SubmitResult) Type() string { return "submit" }

type DiscardResult struct {
	ReviewID string
}

func (r DiscardResult) Type() string { return "discard" }

type NoOpResult struct {
	Message string
}

func (r NoOpResult) Type() string { return "noop" }

func NewFormatter(format Format, w io.Writer) (Formatter, error) {
	switch format {
	case FormatTable:
		return newTableFormatter(w), nil
	case FormatPlain:
		return &plainFormatter{w: w}, nil
	case FormatJSON:
		return &jsonFormatter{w: w}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %q", format)
	}
}
