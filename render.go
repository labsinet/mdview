package main

import (
	"bytes"
	"net/url"
	"path/filepath"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// RenderMarkdown перетворює markdown у HTML.
// baseDir потрібен, щоб відносні картинки та посилання вказували на правильні файли.
func RenderMarkdown(src []byte, baseDir string) (string, error) {
	return render(src, baseDir, false)
}

// RenderMarkdownExport — для «Зберегти як»: картинки вбудовуються як data-URI,
// посилання лишаються як є, тож HTML-файл самодостатній.
func RenderMarkdownExport(src []byte, baseDir string) (string, error) {
	return render(src, baseDir, true)
}

func render(src []byte, baseDir string, export bool) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // таблиці, ~~закреслення~~, task-списки, автопосилання
			extension.Footnote,
			// Підсвічування коду CSS-класами (тема світла/темна — через CSS, див. HighlightCSS).
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithASTTransformers(
				util.Prioritized(&pathRewriter{baseDir: baseDir, export: export}, 100),
			),
		),
		goldmark.WithRendererOptions(
			// Дозволяємо сирий HTML (<img>, <details> тощо) — це ваші локальні файли.
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// pathRewriter переписує відносні шляхи:
//   - картинки -> /_local?p=<абсолютний шлях> (віддає localFileHandler)
//   - посилання на .md -> mdview:open?p=<абсолютний шлях> (обробляє JS)
type pathRewriter struct {
	baseDir string
	export  bool
}

func (t *pathRewriter) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.Image:
			v.Destination = []byte(t.image(string(v.Destination)))
		case *ast.Link:
			v.Destination = []byte(t.link(string(v.Destination)))
		}
		return ast.WalkContinue, nil
	})
}

func isExternal(dest string) bool {
	if dest == "" || strings.HasPrefix(dest, "#") || strings.HasPrefix(dest, "//") {
		return true
	}
	u, err := url.Parse(dest)
	if err != nil {
		return true
	}
	return u.Scheme != "" || u.Host != ""
}

func (t *pathRewriter) abs(rel string) string {
	p, err := url.PathUnescape(rel)
	if err != nil {
		p = rel
	}
	p = filepath.FromSlash(p)
	if !filepath.IsAbs(p) {
		p = filepath.Join(t.baseDir, p)
	}
	return filepath.Clean(p)
}

func (t *pathRewriter) image(dest string) string {
	if isExternal(dest) {
		return dest
	}
	if t.export {
		if uri, ok := dataURI(t.abs(dest)); ok {
			return uri
		}
		return dest
	}
	return "/_local?p=" + url.QueryEscape(t.abs(dest))
}

func (t *pathRewriter) link(dest string) string {
	if t.export || isExternal(dest) {
		return dest
	}
	// відкидаємо #якір
	pathPart := dest
	if i := strings.Index(pathPart, "#"); i >= 0 {
		pathPart = pathPart[:i]
	}
	switch strings.ToLower(filepath.Ext(pathPart)) {
	case ".md", ".markdown":
		return "mdview:open?p=" + url.QueryEscape(t.abs(pathPart))
	}
	return dest
}

// HighlightCSS повертає CSS підсвічування коду. Світла тема — за замовчуванням;
// темна вмикається, якщо система темна (і не вибрано «Світла»)
// або якщо на <html> стоїть data-theme="dark".
func HighlightCSS() string {
	f := chromahtml.New(chromahtml.WithClasses(true))
	var light, dark bytes.Buffer
	_ = f.WriteCSS(&light, styles.Get("github"))
	_ = f.WriteCSS(&dark, styles.Get("github-dark"))
	d := dark.String()
	return light.String() +
		"\n@media (prefers-color-scheme: dark) {\n" + scopeCSS(d, `:root:not([data-theme="light"]) `) + "}\n" +
		scopeCSS(d, `:root[data-theme="dark"] `)
}

// scopeCSS додає префікс до селекторів. Chroma пише кожне правило в один рядок.
func scopeCSS(css, prefix string) string {
	var out strings.Builder
	for _, line := range strings.Split(css, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "/*") {
			if i := strings.Index(line, "*/"); i >= 0 {
				line = strings.TrimSpace(line[i+2:])
			}
		}
		open := strings.Index(line, "{")
		if open < 0 {
			continue
		}
		sels := strings.Split(line[:open], ",")
		for i, sel := range sels {
			sels[i] = prefix + strings.TrimSpace(sel)
		}
		out.WriteString(strings.Join(sels, ", ") + " " + line[open:] + "\n")
	}
	return out.String()
}
