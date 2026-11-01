package main

import (
	"encoding/base64"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const exportMaxImage = 8 << 20 // більші картинки не вбудовуємо

var imageMIME = map[string]string{
	".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".gif": "image/gif",
	".webp": "image/webp", ".svg": "image/svg+xml", ".bmp": "image/bmp",
	".ico": "image/x-icon", ".avif": "image/avif",
}

func dataURI(path string) (string, bool) {
	mt, ok := imageMIME[strings.ToLower(filepath.Ext(path))]
	if !ok {
		return "", false
	}
	st, err := os.Stat(path)
	if err != nil || st.Size() > exportMaxImage {
		return "", false
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return "data:" + mt + ";base64," + base64.StdEncoding.EncodeToString(b), true
}

// SaveAs: діалог «Зберегти як». HTML — самодостатня сторінка (за замовчуванням),
// .md/.markdown/.txt — копія вихідного тексту.
func (a *App) SaveAs() {
	a.mu.Lock()
	cur := a.current
	a.mu.Unlock()
	if cur == "" {
		a.toast(a.tr("ui.nothingToSave"))
		return
	}
	base := strings.TrimSuffix(filepath.Base(cur), filepath.Ext(cur))
	dest, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           a.tr("dialog.save"),
		DefaultFilename: base + ".html",
		Filters: []runtime.FileFilter{
			{DisplayName: a.tr("dialog.filterHtml"), Pattern: "*.html;*.htm"},
			{DisplayName: a.tr("dialog.filterMdCopy"), Pattern: "*.md"},
		},
	})
	if err != nil || dest == "" {
		return
	}
	if dest, err = a.exportTo(cur, dest); err != nil {
		a.toast(a.tr("ui.saveError", "msg", err.Error()))
		return
	}
	a.toast(a.tr("ui.saved", "path", dest))
}

func (a *App) exportTo(src, dest string) (string, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return dest, err
	}
	switch strings.ToLower(filepath.Ext(dest)) {
	case ".md", ".markdown", ".txt":
		return dest, os.WriteFile(dest, data, 0o644)
	case "":
		dest += ".html"
	}
	body, err := RenderMarkdownExport(data, filepath.Dir(src))
	if err != nil {
		return dest, err
	}
	lang, _ := a.langState()
	title := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))
	return dest, os.WriteFile(dest, []byte(buildExportHTML(lang, title, body)), 0o644)
}

func buildExportHTML(lang, title, body string) string {
	// Replacer не перескановує вставлений текст, тож «{{» у документі безпечні.
	return strings.NewReplacer(
		"{{LANG}}", html.EscapeString(lang),
		"{{TITLE}}", html.EscapeString(title),
		"{{CSS}}", exportCSS+"\n"+HighlightCSS(),
		"{{BODY}}", body,
	).Replace(exportTemplate)
}

const exportTemplate = `<!doctype html>
<html lang="{{LANG}}">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{TITLE}}</title>
<style>
{{CSS}}
</style>
</head>
<body>
<article class="md">
{{BODY}}
</article>
</body>
</html>
`

const exportCSS = `
:root { --bg:#f3f5f8; --page:#fff; --ink:#1c2230; --muted:#6b7587; --line:#dde2ea; --accent:#2b5f95; --code-bg:#eef1f6;
  --ui:system-ui,-apple-system,"Segoe UI",Roboto,sans-serif; --read:Charter,"Iowan Old Style","Palatino Linotype",Georgia,serif;
  --mono:ui-monospace,"Cascadia Code",Menlo,Consolas,monospace; color-scheme: light dark; }
@media (prefers-color-scheme: dark) {
  :root { --bg:#12151a; --page:#1a1e25; --ink:#dfe4ec; --muted:#8b95a7; --line:#2b313b; --accent:#82b2e6; --code-bg:#232933; }
}
body { margin:0; padding:24px 16px; background:var(--bg); color:var(--ink); }
.md { max-width:760px; margin:0 auto; background:var(--page); padding:36px 48px; border:1px solid var(--line); border-radius:8px;
  font:17px/1.7 var(--read); overflow-wrap:break-word; }
.md h1,.md h2,.md h3,.md h4 { font-family:var(--ui); line-height:1.25; margin:1.6em 0 .5em; }
.md h1 { font-size:1.9em; margin-top:0; padding-bottom:.3em; border-bottom:1px solid var(--line); }
.md h2 { font-size:1.5em; padding-bottom:.25em; border-bottom:1px solid var(--line); }
.md h3 { font-size:1.2em; }
.md p,.md ul,.md ol,.md pre,.md table,.md blockquote { margin:0 0 1em; }
.md a { color:var(--accent); }
.md img { max-width:100%; height:auto; }
.md code { font:.85em var(--mono); background:var(--code-bg); padding:.15em .35em; border-radius:4px; }
.md pre { background:var(--code-bg); padding:14px 16px; border-radius:6px; overflow:auto; line-height:1.5; }
.md pre code { background:none; padding:0; font-size:.85em; }
.md pre.chroma { background:var(--code-bg); }
.md blockquote { margin-left:0; padding:0 1em; color:var(--muted); border-left:3px solid var(--line); }
.md table { border-collapse:collapse; display:block; overflow:auto; font:15px/1.5 var(--ui); }
.md th,.md td { border:1px solid var(--line); padding:6px 12px; }
.md th { background:var(--code-bg); }
.md hr { border:0; border-top:1px solid var(--line); margin:2em 0; }
@media (max-width:700px) { .md { padding:20px; } }
@media print { body { background:#fff; padding:0; } .md { border:0; padding:0; max-width:none; } }
`
