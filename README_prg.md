# mdview

Легкий Markdown-в'ювер: Go + Wails + goldmark.

## Перший запуск
```
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor          # перевірка залежностей
go mod tidy           # підтягне wails, goldmark, fsnotify
wails dev             # режим розробки
wails build           # готовий бінарник у build/bin/
```
Linux: потрібні `libgtk-3-dev` та `libwebkit2gtk-4.0-dev`
(на Ubuntu 24.04 — `libwebkit2gtk-4.1-dev` і `wails build -tags webkit2_41`).

## Запуск з файлом
`mdview.exe README.md`

## Асоціація .md
- Windows: `wails build -nsis` (потрібен NSIS) — інсталятор створить асоціацію;
  або вручну: «Відкрити за допомогою → Вибрати інший застосунок».
- macOS: береться з `fileAssociations` у wails.json (потрапляє в Info.plist).
- Linux: покласти `build/linux/mdview.desktop` у `~/.local/share/applications/`,
  потім `xdg-mime default mdview.desktop text/markdown`.

## Історія
`os.UserConfigDir()/mdview/history.json`

## Можливості
- Підсвічування коду (chroma, світла/темна тема)
- Зміст за заголовками (вкладка «Зміст», підсвічує поточний розділ)
- Пошук на сторінці: Ctrl+F, Enter / Shift+Enter (або F3) — наступний/попередній, Esc — закрити
- Drag-and-drop: перетягніть `.md` у вікно
- Ctrl+O — відкрити, Ctrl+B — бічна панель

## Мови
Інтерфейс і меню перекладені: українська, English, Polski, Deutsch.
За замовчуванням береться мова системи, вибір — **Вигляд → Мова** (зберігається в `settings.json`).

Додати мову: скопіюйте `locales/en.json` у `locales/<код>.json` (наприклад `fr.json`),
перекладіть значення, у полі `"_name"` напишіть рідну назву мови — і вона з'явиться в меню.
`go test ./...` перевіряє, що в усіх мовах однаковий набір ключів.

## Версія
`wails build -ldflags "-X main.Version=1.0.0"` — номер у вікні «Про програму».

## Тема
Вигляд → Тема (або кнопка ◐ у верхній панелі): системна / світла / темна. Вибір зберігається в `settings.json`.
Підсвічування коду й рамка вікна (Windows) перемикаються разом із темою.

## Зберегти як
Файл → Зберегти як… (Ctrl+Shift+S):
- `.html` — самодостатня сторінка (стилі, підсвічування й картинки вбудовані), з підтримкою світлої/темної теми та друку;
- `.md` / `.txt` — копія вихідного тексту.

## Автор і компанія
- `about.json` — те, що показує вікно «Про програму» (автор, компанія, сайт). Порожнє поле просто не показується.
- `wails.json` (`author`, `info.companyName`, `info.copyright`) — метадані самого `.exe` (Властивості → Подробиці) та інсталятора.
Після змін потрібна нова збірка (`wails build`).
