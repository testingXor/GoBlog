package main

import (
	"bytes"
	"strings"

	marktag "git.jlel.se/jlelse/goldmark-mark"
	"github.com/PuerkitoBio/goquery"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

func (a *goBlog) initMarkdown() {
	defaultGoldmarkOptions := []goldmark.Option{
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.Footnote,
			extension.Typographer,
			extension.Linkify,
			marktag.Mark,
			emoji.Emoji,
		),
	}
	a.md = goldmark.New(append(defaultGoldmarkOptions, goldmark.WithExtensions(&customExtension{
		absoluteLinks: false,
		publicAddress: a.cfg.Server.PublicAddress,
	}))...)
	a.absoluteMd = goldmark.New(append(defaultGoldmarkOptions, goldmark.WithExtensions(&customExtension{
		absoluteLinks: true,
		publicAddress: a.cfg.Server.PublicAddress,
	}))...)
}

func (a *goBlog) renderMarkdown(source string, absoluteLinks bool) (rendered []byte, err error) {
	var buffer bytes.Buffer
	if absoluteLinks {
		err = a.absoluteMd.Convert([]byte(source), &buffer)
	} else {
		err = a.md.Convert([]byte(source), &buffer)
	}
	return buffer.Bytes(), err
}

func (a *goBlog) renderText(s string) string {
	h, err := a.renderMarkdown(s, false)
	if err != nil {
		return ""
	}
	d, err := goquery.NewDocumentFromReader(bytes.NewReader(h))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(d.Text())
}

// Extensions etc...

// Links
type customExtension struct {
	absoluteLinks bool
	publicAddress string
}

func (l *customExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&customRenderer{
			absoluteLinks: l.absoluteLinks,
			publicAddress: l.publicAddress,
		}, 500),
	))
}

type customRenderer struct {
	absoluteLinks bool
	publicAddress string
}

func (c *customRenderer) RegisterFuncs(r renderer.NodeRendererFuncRegisterer) {
	r.Register(ast.KindLink, c.renderLink)
	r.Register(ast.KindImage, c.renderImage)
}

func (c *customRenderer) renderLink(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.Link)
		_, _ = w.WriteString("<a href=\"")
		// Make URL absolute if it's relative
		newDestination := util.URLEscape(n.Destination, true)
		if c.absoluteLinks && c.publicAddress != "" && bytes.HasPrefix(newDestination, []byte("/")) {
			_, _ = w.Write(util.EscapeHTML([]byte(c.publicAddress)))
		}
		_, _ = w.Write(util.EscapeHTML(newDestination))
		_, _ = w.WriteRune('"')
		// Open external links (links that start with "http") in new tab
		if isAbsoluteURL(string(n.Destination)) {
			_, _ = w.WriteString(` target="_blank" rel="noopener"`)
		}
		// Title
		if n.Title != nil {
			_, _ = w.WriteString(" title=\"")
			_, _ = w.Write(n.Title)
			_, _ = w.WriteRune('"')
		}
		_, _ = w.WriteRune('>')
	} else {
		_, _ = w.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}

func (c *customRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	// Make URL absolute if it's relative
	destination := util.URLEscape(n.Destination, true)
	if c.publicAddress != "" && bytes.HasPrefix(destination, []byte("/")) {
		destination = util.EscapeHTML(append([]byte(c.publicAddress), destination...))
	} else {
		destination = util.EscapeHTML(destination)
	}
	_, _ = w.WriteString("<a href=\"")
	_, _ = w.Write(destination)
	_, _ = w.WriteString("\">")
	_, _ = w.WriteString("<img src=\"")
	_, _ = w.Write(destination)
	_, _ = w.WriteString("\" alt=\"")
	_, _ = w.Write(util.EscapeHTML(n.Text(source)))
	_ = w.WriteByte('"')
	_, _ = w.WriteString(" loading=\"lazy\"")
	if n.Title != nil {
		_, _ = w.WriteString(" title=\"")
		_, _ = w.Write(n.Title)
		_ = w.WriteByte('"')
	}
	_, _ = w.WriteString("></a>")
	return ast.WalkSkipChildren, nil
}
