package output

import (
	"fmt"
	"io"
	"strings"
)

type plainFormatter struct {
	w io.Writer
}

func (f *plainFormatter) Format(result Result) error {
	switch r := result.(type) {
	case CommentsResult:
		return f.formatComments(r)
	case ViewResult:
		return f.formatView(r)
	case AddResult:
		return f.formatAdd(r)
	case EditResult:
		return f.formatEdit(r)
	case DeleteResult:
		return f.formatDelete(r)
	case SubmitResult:
		return f.formatSubmit(r)
	case DiscardResult:
		return f.formatDiscard(r)
	case NoOpResult:
		return f.formatNoOp(r)
	default:
		return fmt.Errorf("unknown result type: %T", result)
	}
}

func (f *plainFormatter) formatComments(r CommentsResult) error {
	for _, group := range r.Groups {
		if group.Author != "" {
			fmt.Fprintf(f.w, "@%s\n", group.Author)
		}
		for _, c := range group.Comments {
			parts := []string{c.State}
			if r.IncludeIDs {
				parts = append(parts, c.ID)
			}
			if c.Path != "" {
				loc := c.Path
				if c.Line > 0 {
					loc = fmt.Sprintf("%s:%d", c.Path, c.Line)
				}
				parts = append(parts, loc)
			}
			parts = append(parts, c.Body)
			if group.Author == "" && c.Author != "" {
				parts = append(parts, c.Author)
			}
			prefix := "\t"
			if group.Author == "" {
				prefix = ""
			}
			fmt.Fprintf(f.w, "%s%s\n", prefix, joinTSV(parts))
		}
	}
	return nil
}

func joinTSV(parts []string) string {
	return fmt.Sprintf("%s", strings.Join(parts, "\t"))
}

func (f *plainFormatter) formatView(r ViewResult) error {
	for _, t := range r.Threads {
		status := "resolved"
		if !t.Resolved {
			status = "unresolved"
		}
		loc := t.Path
		if t.Line > 0 {
			loc = fmt.Sprintf("%s:%d", t.Path, t.Line)
		}
		if r.IncludeIDs {
			fmt.Fprintf(f.w, "%s\t%s\t%s\n", t.ID, status, loc)
		} else {
			fmt.Fprintf(f.w, "%s\t%s\n", status, loc)
		}
		for _, c := range t.Comments {
			body := strings.ReplaceAll(c.Body, "\n", " ")
			if r.IncludeIDs {
				fmt.Fprintf(f.w, "\t%s\t%s\t%s\n", c.ID, c.Author, body)
			} else {
				fmt.Fprintf(f.w, "\t%s\t%s\n", c.Author, body)
			}
		}
	}
	return nil
}

func (f *plainFormatter) formatAdd(r AddResult) error {
	fmt.Fprintf(f.w, "added\t%s\t%d\n", r.Path, r.Line)
	return nil
}

func (f *plainFormatter) formatEdit(r EditResult) error {
	fmt.Fprintf(f.w, "edited\t%s\n", r.CommentID)
	return nil
}

func (f *plainFormatter) formatDelete(r DeleteResult) error {
	fmt.Fprintf(f.w, "deleted\t%s\n", r.CommentID)
	return nil
}

func (f *plainFormatter) formatSubmit(r SubmitResult) error {
	fmt.Fprintf(f.w, "submitted\t%s\n", r.Verdict)
	return nil
}

func (f *plainFormatter) formatDiscard(r DiscardResult) error {
	fmt.Fprintf(f.w, "discarded\t%s\n", r.ReviewID)
	return nil
}

func (f *plainFormatter) formatNoOp(r NoOpResult) error {
	fmt.Fprintf(f.w, "noop\t%s\n", r.Message)
	return nil
}
