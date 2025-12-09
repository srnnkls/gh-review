package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/mattn/go-isatty"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12"))

	evenRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	oddRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	authorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14"))
)

type tableFormatter struct {
	w     io.Writer
	isTTY bool
}

func newTableFormatter(w io.Writer) *tableFormatter {
	isTTY := os.Getenv("FORCE_COLOR") != ""
	if !isTTY {
		if f, ok := w.(*os.File); ok {
			isTTY = isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
		}
	}
	return &tableFormatter{w: w, isTTY: isTTY}
}

func (f *tableFormatter) Format(result Result) error {
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

func (f *tableFormatter) formatComments(r CommentsResult) error {
	for i, group := range r.Groups {
		if i > 0 {
			fmt.Fprintln(f.w)
		}

		// Show author header only if author is set (not flat mode)
		if group.Author != "" {
			authorHeader := fmt.Sprintf("@%s (%d comments)", group.Author, len(group.Comments))
			if f.isTTY {
				authorHeader = authorStyle.Render(authorHeader)
			}
			fmt.Fprintln(f.w, authorHeader)
		}

		rows := make([][]string, len(group.Comments))
		for j, c := range group.Comments {
			bodyPreview := c.Body
			if len(bodyPreview) > 40 {
				bodyPreview = bodyPreview[:40] + "..."
			}
			bodyPreview = strings.ReplaceAll(bodyPreview, "\n", " ")

			location := "(global)"
			if c.Path != "" {
				pathShort := c.Path
				if len(pathShort) > 20 {
					pathShort = "..." + pathShort[len(pathShort)-17:]
				}
				if c.Line > 0 {
					location = fmt.Sprintf("%s:%d", pathShort, c.Line)
				} else {
					location = pathShort
				}
			}

			row := []string{c.State, location, bodyPreview}
			if r.IncludeIDs {
				row = append([]string{c.ID}, row...)
			}
			// In flat mode, include author per row
			if group.Author == "" && c.Author != "" {
				row = append(row, c.Author)
			}
			rows[j] = row
		}

		headers := []string{"State", "Location", "Body"}
		if r.IncludeIDs {
			headers = append([]string{"ID"}, headers...)
		}
		if group.Author == "" {
			headers = append(headers, "Author")
		}

		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
			Headers(headers...).
			Rows(rows...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return headerStyle
				}
				if row%2 == 0 {
					return evenRowStyle
				}
				return oddRowStyle
			})

		fmt.Fprintln(f.w, t)
	}
	return nil
}

func (f *tableFormatter) formatView(r ViewResult) error {
	for i, thread := range r.Threads {
		if i > 0 {
			fmt.Fprintln(f.w)
		}

		// Thread header
		status := "resolved"
		if !thread.Resolved {
			status = "unresolved"
		}

		location := thread.Path
		if thread.Line > 0 {
			location = fmt.Sprintf("%s:%d", thread.Path, thread.Line)
		}

		header := fmt.Sprintf("[%s] %s", status, location)
		if r.IncludeIDs {
			header = fmt.Sprintf("[%s] %s (%s)", status, location, thread.ID)
		}
		if f.isTTY {
			if thread.Resolved {
				header = dimStyle.Render(header)
			} else {
				header = authorStyle.Render(header)
			}
		}
		fmt.Fprintln(f.w, header)

		// Comments
		for _, c := range thread.Comments {
			prefix := "  "
			line := fmt.Sprintf("%s@%s: %s", prefix, c.Author, truncateBody(c.Body, 60))
			if r.IncludeIDs {
				line = fmt.Sprintf("%s[%s] @%s: %s", prefix, c.ID, c.Author, truncateBody(c.Body, 50))
			}
			fmt.Fprintln(f.w, line)
		}
	}
	return nil
}

func truncateBody(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func (f *tableFormatter) formatAdd(r AddResult) error {
	msg := fmt.Sprintf("✓ Added comment at %s:%d", r.Path, r.Line)
	if f.isTTY {
		msg = successStyle.Render(msg)
	}
	fmt.Fprintln(f.w, msg)
	return nil
}

func (f *tableFormatter) formatEdit(r EditResult) error {
	msg := fmt.Sprintf("✓ Updated comment %s", r.CommentID)
	if f.isTTY {
		msg = successStyle.Render(msg)
	}
	fmt.Fprintln(f.w, msg)
	return nil
}

func (f *tableFormatter) formatDelete(r DeleteResult) error {
	msg := fmt.Sprintf("✓ Deleted comment %s", r.CommentID)
	if f.isTTY {
		msg = successStyle.Render(msg)
	}
	fmt.Fprintln(f.w, msg)
	return nil
}

func (f *tableFormatter) formatSubmit(r SubmitResult) error {
	msg := fmt.Sprintf("✓ Submitted review (%s)", r.Verdict)
	if f.isTTY {
		msg = successStyle.Render(msg)
	}
	fmt.Fprintln(f.w, msg)
	return nil
}

func (f *tableFormatter) formatDiscard(r DiscardResult) error {
	msg := fmt.Sprintf("✓ Discarded pending review %s", r.ReviewID)
	if f.isTTY {
		msg = successStyle.Render(msg)
	}
	fmt.Fprintln(f.w, msg)
	return nil
}

func (f *tableFormatter) formatNoOp(r NoOpResult) error {
	msg := r.Message
	if f.isTTY {
		msg = dimStyle.Render(msg)
	}
	fmt.Fprintln(f.w, msg)
	return nil
}
