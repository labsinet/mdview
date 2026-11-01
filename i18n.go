package main

import (
	"embed"
	"encoding/json"
	"sort"
	"strings"

	locale "github.com/jeandeaual/go-locale"
)

// Кожен файл locales/<код>.json — окрема мова. Щоб додати мову, достатньо
// скопіювати en.json, перекласти значення та змінити поле "_name".
//
//go:embed locales/*.json
var localeFS embed.FS

const fallbackLang = "en"

type LangInfo struct {
	Code string `json:"code"`
	Name string `json:"name"` // рідна назва мови з поля "_name"
}

type I18n struct {
	bundles map[string]map[string]string
	langs   []LangInfo
}

func LoadI18n() *I18n {
	i := &I18n{bundles: map[string]map[string]string{}}
	entries, _ := localeFS.ReadDir("locales")
	for _, e := range entries {
		data, err := localeFS.ReadFile("locales/" + e.Name())
		if err != nil {
			continue
		}
		var m map[string]string
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		code := strings.TrimSuffix(e.Name(), ".json")
		i.bundles[code] = m
		name := m["_name"]
		if name == "" {
			name = code
		}
		i.langs = append(i.langs, LangInfo{Code: code, Name: name})
	}
	sort.Slice(i.langs, func(a, b int) bool { return i.langs[a].Name < i.langs[b].Name })
	return i
}

func (i *I18n) Has(code string) bool { _, ok := i.bundles[code]; return ok }

func (i *I18n) Languages() []LangInfo { return i.langs }

// Resolve: явний вибір користувача -> мова системи -> англійська.
func (i *I18n) Resolve(pref string) string {
	if pref != "" && i.Has(pref) {
		return pref
	}
	if tag, err := locale.GetLocale(); err == nil {
		tag = strings.ToLower(tag)
		if cut := strings.IndexAny(tag, "-_."); cut > 0 {
			tag = tag[:cut]
		}
		if i.Has(tag) {
			return tag
		}
	}
	return fallbackLang
}

// Strings повертає переклади мови; відсутні ключі добираються з англійської.
func (i *I18n) Strings(lang string) map[string]string {
	out := make(map[string]string, len(i.bundles[fallbackLang]))
	for k, v := range i.bundles[fallbackLang] {
		out[k] = v
	}
	for k, v := range i.bundles[lang] {
		out[k] = v
	}
	return out
}

// T підставляє пари {ключ}=значення: T("uk", "ui.openError", "msg", "...").
func (i *I18n) T(lang, key string, kv ...string) string {
	s, ok := i.bundles[lang][key]
	if !ok {
		s, ok = i.bundles[fallbackLang][key]
	}
	if !ok {
		return key
	}
	for n := 0; n+1 < len(kv); n += 2 {
		s = strings.ReplaceAll(s, "{"+kv[n]+"}", kv[n+1])
	}
	return s
}
