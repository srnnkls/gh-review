package output

import (
	"encoding/json"
	"fmt"
	"io"
)

type jsonFormatter struct {
	w io.Writer
}

func (f *jsonFormatter) Format(result Result) error {
	var v interface{}

	switch r := result.(type) {
	case CommentsResult:
		v = f.formatComments(r)
	case ViewResult:
		v = f.formatView(r)
	case AddResult:
		v = f.formatAdd(r)
	case EditResult:
		v = f.formatEdit(r)
	case DeleteResult:
		v = f.formatDelete(r)
	case SubmitResult:
		v = f.formatSubmit(r)
	case DiscardResult:
		v = f.formatDiscard(r)
	case NoOpResult:
		v = f.formatNoOp(r)
	default:
		return fmt.Errorf("unknown result type: %T", result)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	if _, err := f.w.Write(data); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}

	if _, err := f.w.Write([]byte("\n")); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	return nil
}

type jsonCommentsResult struct {
	PR     string             `json:"pr"`
	Groups []jsonCommentGroup `json:"groups"`
}

type jsonCommentGroup struct {
	Author   string        `json:"author,omitempty"`
	Comments []jsonComment `json:"comments"`
}

type jsonComment struct {
	ID    string `json:"id,omitempty"`
	State string `json:"state"`
	Path  string `json:"path,omitempty"`
	Line  int    `json:"line,omitempty"`
	Body  string `json:"body"`
}

func (f *jsonFormatter) formatComments(r CommentsResult) jsonCommentsResult {
	groups := make([]jsonCommentGroup, len(r.Groups))
	for i, g := range r.Groups {
		comments := make([]jsonComment, len(g.Comments))
		for j, c := range g.Comments {
			cmt := jsonComment{
				State: c.State,
				Path:  c.Path,
				Line:  c.Line,
				Body:  c.Body,
			}
			if r.IncludeIDs {
				cmt.ID = c.ID
			}
			comments[j] = cmt
		}
		groups[i] = jsonCommentGroup{
			Author:   g.Author,
			Comments: comments,
		}
	}

	return jsonCommentsResult{
		PR:     r.PRRef,
		Groups: groups,
	}
}

type jsonViewResult struct {
	PR      string           `json:"pr"`
	Threads []jsonViewThread `json:"threads"`
}

type jsonViewThread struct {
	ID       string               `json:"id,omitempty"`
	Path     string               `json:"path"`
	Line     int                  `json:"line,omitempty"`
	Resolved bool                 `json:"resolved"`
	Comments []jsonViewComment    `json:"comments"`
}

type jsonViewComment struct {
	ID     string `json:"id,omitempty"`
	Author string `json:"author"`
	Body   string `json:"body"`
}

func (f *jsonFormatter) formatView(r ViewResult) jsonViewResult {
	threads := make([]jsonViewThread, len(r.Threads))
	for i, t := range r.Threads {
		comments := make([]jsonViewComment, len(t.Comments))
		for j, c := range t.Comments {
			cmt := jsonViewComment{
				Author: c.Author,
				Body:   c.Body,
			}
			if r.IncludeIDs {
				cmt.ID = c.ID
			}
			comments[j] = cmt
		}
		thread := jsonViewThread{
			Path:     t.Path,
			Line:     t.Line,
			Resolved: t.Resolved,
			Comments: comments,
		}
		if r.IncludeIDs {
			thread.ID = t.ID
		}
		threads[i] = thread
	}

	return jsonViewResult{
		PR:      r.PRRef,
		Threads: threads,
	}
}

type jsonAddResult struct {
	Action string `json:"action"`
	Path   string `json:"path"`
	Line   int    `json:"line"`
}

func (f *jsonFormatter) formatAdd(r AddResult) jsonAddResult {
	return jsonAddResult{
		Action: "added",
		Path:   r.Path,
		Line:   r.Line,
	}
}

type jsonEditResult struct {
	Action    string `json:"action"`
	CommentID string `json:"comment_id"`
}

func (f *jsonFormatter) formatEdit(r EditResult) jsonEditResult {
	return jsonEditResult{
		Action:    "edited",
		CommentID: r.CommentID,
	}
}

type jsonDeleteResult struct {
	Action    string `json:"action"`
	CommentID string `json:"comment_id"`
}

func (f *jsonFormatter) formatDelete(r DeleteResult) jsonDeleteResult {
	return jsonDeleteResult{
		Action:    "deleted",
		CommentID: r.CommentID,
	}
}

type jsonSubmitResult struct {
	Action  string `json:"action"`
	Verdict string `json:"verdict"`
}

func (f *jsonFormatter) formatSubmit(r SubmitResult) jsonSubmitResult {
	return jsonSubmitResult{
		Action:  "submitted",
		Verdict: r.Verdict,
	}
}

type jsonDiscardResult struct {
	Action   string `json:"action"`
	ReviewID string `json:"review_id"`
}

func (f *jsonFormatter) formatDiscard(r DiscardResult) jsonDiscardResult {
	return jsonDiscardResult{
		Action:   "discarded",
		ReviewID: r.ReviewID,
	}
}

type jsonNoOpResult struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

func (f *jsonFormatter) formatNoOp(r NoOpResult) jsonNoOpResult {
	return jsonNoOpResult{
		Action:  "noop",
		Message: r.Message,
	}
}
