# mdview

**A lightweight, cross-platform Markdown viewer.** Double-click a `.md` file, read it, get out of the way.

English | [Українська](README.uk.md)

![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-v2-red)
![Platforms](https://img.shields.io/badge/platforms-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)
![License](https://img.shields.io/badge/license-MIT-green)

<!-- Add a screenshot: docs/screenshot.png
![mdview screenshot](docs/screenshot.png)
-->

mdview is a small desktop app built with Go and [Wails](https://wails.io). It renders Markdown with
[goldmark](https://github.com/yuin/goldmark) in the system WebView, so there is no bundled browser:
the binary is a few megabytes and starts instantly. It is a *viewer*, not an editor, so the interface
is deliberately minimal.

## Contents

- [Features](#features)
- [Getting started](#getting-started)
- [Usage](#usage)
- [Opening `.md` files with a double-click](#opening-md-files-with-a-double-click)
- [Settings and data](#settings-and-data)
- [Languages](#languages)
- [Customizing: author, company, version](#customizing-author-company-version)
- [Project structure](#project-structure)
- [How it works](#how-it-works)
- [Known limitations](#known-limitations)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Fast and small**: Go backend + the system WebView (WebView2 / WKWebView / WebKitGTK).
- **GitHub-flavoured Markdown**: tables, task lists, strikethrough, autolinks, footnotes, raw HTML.
- **Syntax highlighting** for fenced code blocks (chroma), with matching light and dark palettes.
- **Table of contents** generated from headings, with the current section highlighted while you scroll.
- **Find on page** with match counter and next/previous navigation.
- **Open history** (last 50 files) in the sidebar; missing files are shown struck through.
- **Live reload**: the page refreshes when the file changes on disk and keeps the scroll position.
- **Drag and drop** a Markdown file into the window.
- **Single instance**: opening another `.md` reuses the running window.
- **Relative images and links work**: local images are served from disk, links to other `.md` files open inside mdview, external links open in your browser.
- **Light / dark / system theme**, switchable from the menu or the button in the top bar.
- **Save as…**: export a self-contained HTML page (styles, highlighting and images embedded; print-friendly) or copy the source text.
- **Localized UI**: Ukrainian, English, Polish, German; new languages are a single JSON file.
- **Cross-platform**: Windows, macOS, Linux from one codebase.

## Getting started

### Requirements

| | |
|---|---|
| Go | 1.21 or newer |
| Wails CLI | v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`) |
| Windows | WebView2 runtime (preinstalled on Windows 10/11) |
| macOS | Xcode command line tools (`xcode-select --install`) |
| Linux | `libgtk-3-dev` and `libwebkit2gtk-4.0-dev` (Ubuntu 24.04: `libwebkit2gtk-4.1-dev`, build with `-tags webkit2_41`) |

Run `wails doctor` to check that everything is in place.

### Build from source

```bash
git clone https://github.com/<your-user>/mdview.git
cd mdview
go mod tidy
wails build
```

The binary appears in `build/bin/` (`mdview.exe` on Windows, `mdview.app` on macOS, `mdview` on Linux).

Development mode with hot reload for the frontend:

```bash
wails dev
```

> Wails uses native libraries, so build each platform **on that platform** (or use the included CI, see below).

### Useful build variants

| Command | Result |
|---|---|
| `wails build -clean` | clean `build/bin` first |
| `wails build -ldflags "-s -w"` | smaller binary (strips debug info) |
| `wails build -ldflags "-X main.Version=1.0.0"` | sets the version shown in *About* |
| `wails build -nsis` | Windows installer with file association (requires [NSIS](https://nsis.sourceforge.io/)) |
| `wails build -tags webkit2_41` | Linux with WebKitGTK 4.1 (Ubuntu 24.04+) |

### Continuous integration

`.github/workflows/build.yml` builds Windows, macOS and Linux binaries with a GitHub Actions matrix.
Push a tag like `v0.3.0` (or run the workflow manually) and download the artifacts from the **Actions** tab.

## Usage

```bash
mdview README.md        # open a file directly
mdview                  # start empty, then use File → Open, Ctrl+O or drag a file in
```

### Keyboard shortcuts

| Shortcut | Action |
|---|---|
| `Ctrl/Cmd + O` | Open file |
| `Ctrl/Cmd + Shift + S` | Save as… |
| `Ctrl/Cmd + F` | Find on page |
| `Enter` / `Shift + Enter` | Next / previous match |
| `F3` / `Shift + F3`, `Ctrl/Cmd + G` | Next / previous match |
| `Esc` | Close the search bar |
| `Ctrl/Cmd + B` | Show or hide the sidebar |

### Menu

- **File**: Open…, Save as…, Quit
- **View**: Toggle sidebar, Find on page, Theme, Language
- **Help**: About mdview (on macOS: in the application menu)

## Opening `.md` files with a double-click

mdview receives the file path as a command-line argument (Windows, Linux) or an OS event (macOS Finder).

**Windows**
- Easiest: right-click any `.md` → *Open with* → *Choose another app* → select `mdview.exe` → *Always*.
- Or build an installer with `wails build -nsis`; it registers `.md` and `.markdown` automatically.

**macOS**
The association is declared in `wails.json` (`fileAssociations`) and written to `Info.plist`.
After copying `mdview.app` to *Applications*, use *Get Info → Open with → Change All…* on any `.md` file.

**Linux**
```bash
sudo install -m 755 build/bin/mdview /usr/local/bin/mdview
install -Dm 644 build/linux/mdview.desktop ~/.local/share/applications/mdview.desktop
xdg-mime default mdview.desktop text/markdown
```

## Settings and data

Stored in the standard per-user config directory (`os.UserConfigDir()`):

| OS | Location |
|---|---|
| Windows | `%APPDATA%\mdview\` |
| macOS | `~/Library/Application Support/mdview/` |
| Linux | `~/.config/mdview/` |

| File | Contents |
|---|---|
| `history.json` | recently opened files (max 50) |
| `settings.json` | `language` and `theme` (empty = follow the system) |

The path is also shown in the *About* window. Delete the files to reset the app.

## Languages

The interface follows the system language and falls back to English. You can override it in **View → Language**.

Translations live in `locales/<code>.json` and are embedded into the binary. To add a language:

1. Copy `locales/en.json` to `locales/<code>.json` (for example `fr.json`).
2. Translate the values; set `"_name"` to the language's own name (it appears in the menu).
3. Rebuild. Missing keys fall back to English.

`go test ./...` checks that every locale has exactly the same set of keys.

## Customizing: author, company, version

| What | Where |
|---|---|
| Author, company, website shown in *About* | `about.json` |
| `.exe` properties (Details tab) and installer metadata | `wails.json` → `author`, `info.companyName`, `info.copyright`, `info.productVersion` |
| Version shown in *About* | `-ldflags "-X main.Version=x.y.z"`, default in `lang.go` |

Empty fields in `about.json` are simply not displayed.

## Project structure

```
mdview/
├── main.go            # Wails setup, single instance, drag-and-drop, macOS file-open event
├── app.go             # methods exposed to JS, file watching, local image handler
├── render.go          # goldmark pipeline, link/image rewriting, highlight CSS
├── export.go          # "Save as": standalone HTML export
├── history.go         # history.json (atomic writes)
├── settings.go        # settings.json, config directory
├── i18n.go            # embedded locales, language detection
├── lang.go            # language & About bindings, version
├── theme.go           # theme setting, native window theme, toasts
├── menu.go            # native menu, rebuilt when language/theme changes
├── about.json         # author / company / website
├── locales/           # uk.json, en.json, pl.json, de.json
├── frontend/          # plain HTML/CSS/JS, no build step
│   ├── index.html
│   ├── style.css
│   └── app.js
├── build/linux/       # .desktop file
├── .github/workflows/ # CI builds for 3 platforms
└── wails.json
```

## How it works

- **Rendering**: goldmark (GFM, footnotes, auto heading IDs) + goldmark-highlighting with CSS classes. A custom AST transformer rewrites relative image URLs to `/_local?p=…` and links to other Markdown files to `mdview:open?p=…`.
- **Local images**: a small handler registered in the Wails asset server serves only image file types from disk.
- **Live reload**: `fsnotify` watches the file's *directory* (editors often save via temp file + rename) and debounces events.
- **Themes**: CSS variables with `prefers-color-scheme`, overridden by `data-theme="light|dark"` on `<html>`; the highlight palette is generated per theme from chroma styles (`github` / `github-dark`).
- **Menu and i18n**: strings come from the same embedded JSON as the frontend; changing language or theme rebuilds the native menu and notifies the UI through Wails events.
- **Export**: the same renderer in "export" mode embeds images as data URIs (up to 8 MB each) and wraps the result in a standalone page.

## Known limitations

- Raw HTML inside Markdown is rendered as-is (`html.WithUnsafe`). That is fine for your own files; consider disabling it in `render.go` before opening untrusted documents.
- In-document anchor links to headings written in non-Latin scripts (e.g. `[see](#заголовок)`) do not resolve, because goldmark generates ASCII-only heading IDs. The table of contents is not affected.
- The window frame follows the chosen theme on Windows only; on macOS and Linux it follows the system.
- Rendering differs slightly between WebView2, WebKit (macOS) and WebKitGTK; check your CSS on all three.
- Only local, relative resources are resolved; no Mermaid, math or remote embeds yet.

## Troubleshooting

| Message / symptom | Explanation |
|---|---|
| `Wails is now using the new Go WebView2Loader` | Informational. Nothing to do. |
| `Warning: Cannot create installer: makensis not found` | You used `-nsis` without NSIS installed. Install it (`winget install NSIS.NSIS`) or build without `-nsis`. |
| Linux: build fails on `webkit2gtk-4.0` | Install `libwebkit2gtk-4.1-dev` and build with `-tags webkit2_41`. |
| File dialog or drop does nothing on a very old Wails | Update the CLI and run `go get github.com/wailsapp/wails/v2@latest && go mod tidy`. |

## Contributing

Issues and pull requests are welcome. Please run `gofmt -w .` and `go test ./...` before submitting.
Ideas on the roadmap: Mermaid diagrams, math (KaTeX), reading-width and font-size settings, PDF export.

## License

[MIT](LICENSE)

## Acknowledgements

[Wails](https://wails.io) · [goldmark](https://github.com/yuin/goldmark) ·
[goldmark-highlighting](https://github.com/yuin/goldmark-highlighting) · [chroma](https://github.com/alecthomas/chroma) ·
[fsnotify](https://github.com/fsnotify/fsnotify) · [go-locale](https://github.com/jeandeaual/go-locale)
