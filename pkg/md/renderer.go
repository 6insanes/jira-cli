package md

import (
	"bytes"
	"io"

	cf "github.com/kentaro-m/blackfriday-confluence"
	bf "github.com/russross/blackfriday/v2"
)

// renderer adapts the Confluence renderer to Jira wiki markup. The upstream
// renderer escapes punctuation that is only special in Confluence, which Jira
// stores verbatim and downstream integrations then read as literal backslashes.
// Plain URLs are kept as-is so Jira and other consumers can autolink them
// instead of receiving the redundant [url|url] form.
type renderer struct {
	*cf.Renderer

	buf bytes.Buffer
}

func newRenderer() *renderer {
	return &renderer{Renderer: &cf.Renderer{Flags: cf.IgnoreMacroEscaping}}
}

// Render prints out the whole document from the ast.
func (r *renderer) Render(ast *bf.Node) []byte {
	ast.Walk(func(node *bf.Node, entering bool) bf.WalkStatus {
		return r.RenderNode(&r.buf, node, entering)
	})

	return r.buf.Bytes()
}

// RenderNode renders a node, bypassing the Confluence escaping for plain text
// and treating autolinks as bare URLs.
func (r *renderer) RenderNode(w io.Writer, node *bf.Node, entering bool) bf.WalkStatus {
	switch node.Type {
	case bf.Text, bf.Code:
		if entering {
			_, _ = w.Write(node.Literal)
		}
		return bf.GoToNext
	case bf.Link:
		if entering && isAutolink(node) {
			_, _ = w.Write(node.Destination)
			return bf.SkipChildren
		}
	}

	return r.Renderer.RenderNode(w, node, entering)
}

func isAutolink(node *bf.Node) bool {
	child := node.FirstChild
	return child != nil &&
		child.Type == bf.Text &&
		child.Next == nil &&
		bytes.Equal(child.Literal, node.Destination)
}
