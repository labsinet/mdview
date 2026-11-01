// Обгортки над Go-методами та runtime, які Wails інжектить у window.
const api = () => window.go.main.App;
const rt = () => window.runtime;
const $ = (s) => document.querySelector(s);

let current = "";
let dict = {};
let langInfo = { lang: "en" };
let aboutInfo = null;
let themePref = "";
let toastTimer;

/* ====================== Переклади ====================== */

function t(key, vars) {
  let s = dict[key] ?? key;
  if (vars) for (const k in vars) s = s.split("{" + k + "}").join(vars[k]);
  return s;
}

function applyI18n() {
  document.documentElement.lang = langInfo.lang;
  // Не чіпаємо вміст документа, крім заглушки «порожній стан».
  const inDoc = (el) => el.closest("#doc") && !el.closest(".empty");
  for (const el of document.querySelectorAll("[data-i18n]")) if (!inDoc(el)) el.textContent = t(el.dataset.i18n);
  for (const el of document.querySelectorAll("[data-i18n-title]")) el.title = t(el.dataset.i18nTitle);
  for (const el of document.querySelectorAll("[data-i18n-placeholder]")) el.placeholder = t(el.dataset.i18nPlaceholder);
  for (const el of document.querySelectorAll("[data-i18n-aria]")) el.setAttribute("aria-label", t(el.dataset.i18nAria));
  document.body.dataset.drop = t("ui.dropHint");
}

function setI18n(info) {
  langInfo = info;
  dict = info.strings || {};
  applyI18n();
}
let headings = [];
let tocItems = [];
let hits = [];
let hitIdx = -1;
const MAX_HITS = 5000;

/* ====================== Тема та сповіщення ====================== */

const THEMES = ["", "light", "dark"];
const THEME_ICON = { "": "◐", light: "☀", dark: "☾" };
const THEME_KEY = { "": "menu.themeAuto", light: "menu.themeLight", dark: "menu.themeDark" };

function updateThemeButton() {
  const b = $("#theme");
  b.textContent = THEME_ICON[themePref];
  b.title = t("menu.theme") + ": " + t(THEME_KEY[themePref]);
}

function applyTheme(pref) {
  themePref = THEMES.includes(pref) ? pref : "";
  if (themePref) document.documentElement.dataset.theme = themePref;
  else delete document.documentElement.dataset.theme;
  updateThemeButton();
}

$("#theme").addEventListener("click", () => {
  api().SetTheme(THEMES[(THEMES.indexOf(themePref) + 1) % THEMES.length]);
});

function toast(msg) {
  const el = $("#toast");
  el.textContent = msg;
  el.hidden = false;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { el.hidden = true; }, 4000);
}

/* ====================== Документ ====================== */

function showDoc(res, keepScroll) {
  const content = $("#content");
  const y = keepScroll ? content.scrollTop : 0;
  const prevHit = hitIdx;

  $("#doc").innerHTML = res.html;
  $("#path").textContent = res.path;
  rt().WindowSetTitle(res.name + " — mdview");
  current = res.path;

  buildToc();
  if (!$("#find").hidden) runSearch(keepScroll ? prevHit : 0, false);
  content.scrollTop = y;
  updateSpy();
  refreshHistory();
}

function showError(msg) {
  const box = document.createElement("p");
  box.className = "error";
  box.textContent = t("ui.openError", { msg });
  $("#doc").replaceChildren(box);
  buildToc();
}

async function openPath(path) {
  try {
    showDoc(await api().OpenFile(path), false);
  } catch (e) {
    showError(String(e));
    refreshHistory();
  }
}

async function pickAndOpen() {
  const p = await api().PickFile();
  if (p) openPath(p);
}

/* ====================== Бічна панель ====================== */

function setTab(name) {
  const isToc = name === "toc";
  $("#history").hidden = isToc;
  $("#toc").hidden = !isToc;
  $("#clear").hidden = isToc;
  $("#tab-history").classList.toggle("active", !isToc);
  $("#tab-toc").classList.toggle("active", isToc);
}

function dirOf(p) {
  const i = Math.max(p.lastIndexOf("/"), p.lastIndexOf("\\"));
  return i > 0 ? p.slice(0, i) : p;
}

async function refreshHistory() {
  const list = (await api().GetHistory()) || [];
  const ul = $("#history");
  ul.replaceChildren();
  if (!list.length) {
    const li = document.createElement("li");
    li.className = "hist-empty";
    li.textContent = t("ui.historyEmpty");
    ul.append(li);
    return;
  }
  for (const e of list) {
    const li = document.createElement("li");
    if (e.path === current) li.classList.add("active");
    if (!e.exists) li.classList.add("missing");

    const item = document.createElement("div");
    item.className = "item";
    item.tabIndex = 0;
    item.title = e.path;
    const name = document.createElement("span");
    name.className = "name";
    name.textContent = e.name;
    const dir = document.createElement("span");
    dir.className = "dir";
    dir.textContent = dirOf(e.path);
    item.append(name, dir);
    item.addEventListener("click", () => openPath(e.path));
    item.addEventListener("keydown", (k) => { if (k.key === "Enter") openPath(e.path); });

    const x = document.createElement("button");
    x.className = "ghost x";
    x.textContent = "×";
    x.title = t("ui.removeFromHistory");
    x.addEventListener("click", async () => { await api().RemoveFromHistory(e.path); refreshHistory(); });

    li.append(item, x);
    ul.append(li);
  }
}

/* ====================== Зміст ====================== */

function buildToc() {
  const ul = $("#toc");
  ul.replaceChildren();
  headings = [...$("#doc").querySelectorAll("h1, h2, h3, h4")];
  tocItems = [];

  if (!headings.length) {
    const li = document.createElement("li");
    li.className = "hist-empty";
    li.textContent = t("ui.noHeadings");
    ul.append(li);
    return;
  }
  const min = Math.min(...headings.map((h) => Number(h.tagName[1])));
  for (const h of headings) {
    const li = document.createElement("li");
    const item = document.createElement("div");
    item.className = "toc-item";
    item.tabIndex = 0;
    item.style.paddingLeft = 8 + (Number(h.tagName[1]) - min) * 14 + "px";
    item.textContent = h.textContent;
    item.title = h.textContent;
    const go = () => h.scrollIntoView({ behavior: "smooth", block: "start" });
    item.addEventListener("click", go);
    item.addEventListener("keydown", (k) => { if (k.key === "Enter") go(); });
    li.append(item);
    ul.append(li);
    tocItems.push(li);
  }
}

// Підсвічує в змісті поточний розділ під час прокрутки.
let spyQueued = false;
function updateSpy() {
  if (!headings.length) return;
  const top = $("#content").getBoundingClientRect().top + 24;
  let idx = 0;
  for (let i = 0; i < headings.length; i++) {
    if (headings[i].getBoundingClientRect().top <= top) idx = i;
    else break;
  }
  tocItems.forEach((li, i) => li.classList.toggle("active", i === idx));
  if (!$("#toc").hidden) tocItems[idx].scrollIntoView({ block: "nearest" });
}
$("#content").addEventListener("scroll", () => {
  if (spyQueued) return;
  spyQueued = true;
  requestAnimationFrame(() => { spyQueued = false; updateSpy(); });
});

/* ====================== Пошук ====================== */

function clearMarks() {
  for (const m of document.querySelectorAll("#doc mark.hit")) {
    const p = m.parentNode;
    p.replaceChild(document.createTextNode(m.textContent), m);
    p.normalize();
  }
  hits = [];
  hitIdx = -1;
}

function runSearch(startIdx, scroll) {
  clearMarks();
  const q = $("#q").value;
  if (!q) { updateCount(); return; }

  const re = new RegExp(q.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"), "giu");
  const walker = document.createTreeWalker($("#doc"), NodeFilter.SHOW_TEXT);
  const nodes = [];
  while (walker.nextNode()) nodes.push(walker.currentNode);

  outer: for (const node of nodes) {
    const text = node.nodeValue;
    re.lastIndex = 0;
    let m, last = 0, frag = null;
    while ((m = re.exec(text))) {
      if (!m[0].length) { re.lastIndex++; continue; }
      if (hits.length >= MAX_HITS) break;
      if (!frag) frag = document.createDocumentFragment();
      if (m.index > last) frag.append(text.slice(last, m.index));
      const mark = document.createElement("mark");
      mark.className = "hit";
      mark.textContent = m[0];
      frag.append(mark);
      hits.push(mark);
      last = m.index + m[0].length;
    }
    if (frag) {
      if (last < text.length) frag.append(text.slice(last));
      node.replaceWith(frag);
    }
    if (hits.length >= MAX_HITS) break outer;
  }
  goTo(Math.min(Math.max(startIdx, 0), hits.length - 1), scroll);
}

function goTo(i, scroll = true) {
  if (!hits.length) { hitIdx = -1; updateCount(); return; }
  if (hitIdx >= 0 && hits[hitIdx]) hits[hitIdx].classList.remove("cur");
  hitIdx = (i + hits.length) % hits.length;
  const el = hits[hitIdx];
  el.classList.add("cur");
  if (scroll) el.scrollIntoView({ block: "center" });
  updateCount();
}

function updateCount() {
  const c = $("#count");
  if (!$("#q").value) c.textContent = "";
  else if (!hits.length) c.textContent = t("ui.noMatches");
  else c.textContent = `${hitIdx + 1} / ${hits.length}${hits.length >= MAX_HITS ? "+" : ""}`;
}

function openFind() {
  const sel = String(window.getSelection() || "");
  $("#find").hidden = false;
  if (sel && sel.length < 100 && !sel.includes("\n")) $("#q").value = sel;
  $("#q").focus();
  $("#q").select();
  runSearch(0, true);
}

function closeFind() {
  $("#find").hidden = true;
  clearMarks();
  updateCount();
}

let findTimer;
$("#q").addEventListener("input", () => {
  clearTimeout(findTimer);
  findTimer = setTimeout(() => runSearch(0, true), 150);
});
$("#q").addEventListener("keydown", (ev) => {
  if (ev.key === "Enter") { ev.preventDefault(); goTo(hitIdx + (ev.shiftKey ? -1 : 1)); }
  if (ev.key === "Escape") closeFind();
});
$("#next").addEventListener("click", () => goTo(hitIdx + 1));
$("#prev").addEventListener("click", () => goTo(hitIdx - 1));
$("#fclose").addEventListener("click", closeFind);

/* ====================== Загальні обробники ====================== */

// Посилання: зовнішні -> браузер, на інші .md -> відкрити тут, #якорі -> як є.
document.addEventListener("click", (ev) => {
  const a = ev.target.closest("a");
  if (!a) return;
  const href = a.getAttribute("href") || "";
  if (href.startsWith("#")) return;
  ev.preventDefault();
  if (href.startsWith("mdview:open?")) {
    const p = new URLSearchParams(href.slice("mdview:open?".length)).get("p");
    if (p) openPath(p);
  } else if (/^(https?:|mailto:)/i.test(href)) {
    rt().BrowserOpenURL(href);
  }
});

document.addEventListener("keydown", (ev) => {
  if (ev.key === "F3") { ev.preventDefault(); if (hits.length) goTo(hitIdx + (ev.shiftKey ? -1 : 1)); return; }
  if (!(ev.ctrlKey || ev.metaKey)) return;
  const k = ev.key.toLowerCase();
  if (k === "s" && ev.shiftKey) { ev.preventDefault(); api().SaveAs(); return; }
  if (k === "o") { ev.preventDefault(); pickAndOpen(); }
  if (k === "b") { ev.preventDefault(); $("#sidebar").classList.toggle("hidden"); }
  if (k === "f") { ev.preventDefault(); openFind(); }
  if (k === "g" && hits.length) { ev.preventDefault(); goTo(hitIdx + (ev.shiftKey ? -1 : 1)); }
});

$("#open").addEventListener("click", pickAndOpen);
$("#toggle").addEventListener("click", () => $("#sidebar").classList.toggle("hidden"));
$("#clear").addEventListener("click", async () => { await api().ClearHistory(); refreshHistory(); });
$("#tab-history").addEventListener("click", () => setTab("history"));
$("#tab-toc").addEventListener("click", () => { setTab("toc"); updateSpy(); });

/* ====================== Про програму ====================== */

function renderAbout() {
  if (!aboutInfo) return;
  $("#about-version").textContent = t("about.version", { v: aboutInfo.version });
  const author = $("#about-author");
  author.hidden = !aboutInfo.author;
  author.textContent = aboutInfo.author ? t("about.author", { v: aboutInfo.author }) : "";
  const company = $("#about-company");
  company.hidden = !aboutInfo.company;
  company.textContent = aboutInfo.company ? t("about.company", { v: aboutInfo.company }) : "";
  const site = $("#about-site");
  site.hidden = !aboutInfo.website;
  site.href = aboutInfo.website || "#";
  site.textContent = (aboutInfo.website || "").replace(/^https?:\/\//, "");
  $("#about-platform").textContent = t("about.platform", { os: aboutInfo.os, arch: aboutInfo.arch });
  $("#about-dir").textContent = aboutInfo.configDir;
}

async function showAbout() {
  aboutInfo = await api().GetAboutInfo();
  renderAbout();
  const d = $("#about");
  if (!d.open) d.showModal();
}

$("#about-close").addEventListener("click", () => $("#about").close());
$("#about").addEventListener("click", (ev) => { if (ev.target === $("#about")) $("#about").close(); });

window.addEventListener("DOMContentLoaded", async () => {
  setI18n(await api().GetI18n());
  applyTheme(await api().GetTheme());

  // CSS підсвічування коду генерується в Go (світла + темна тема).
  try {
    const style = document.createElement("style");
    style.textContent = await api().GetHighlightCSS();
    document.head.append(style);
  } catch (e) { console.warn("highlight css:", e); }

  rt().EventsOn("file-opened", (res) => showDoc(res, false));   // 2-й запуск / Finder / drag-and-drop
  rt().EventsOn("file-changed", (res) => {                      // файл змінено на диску
    if (res.path === current) showDoc(res, true);
  });
  rt().EventsOn("open-error", showError);
  rt().EventsOn("lang-changed", (info) => {                     // зміна мови з меню
    setI18n(info);
    refreshHistory();
    buildToc();
    updateCount();
    renderAbout();
    updateThemeButton();
  });
  rt().EventsOn("theme-changed", applyTheme);
  rt().EventsOn("toast", toast);
  rt().EventsOn("show-about", showAbout);                       // Довідка → Про програму
  rt().EventsOn("menu-sidebar", () => $("#sidebar").classList.toggle("hidden"));
  rt().EventsOn("menu-find", openFind);

  setTab("history");
  await refreshHistory();
  const initial = await api().GetInitialFile();
  if (initial) openPath(initial);
});
