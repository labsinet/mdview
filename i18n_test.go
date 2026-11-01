package main

import "testing"

// Усі мови мають містити ті самі ключі, що й англійська.
func TestLocalesHaveSameKeys(t *testing.T) {
	i := LoadI18n()
	en := i.bundles[fallbackLang]
	if len(en) == 0 {
		t.Fatal("english locale missing")
	}
	for code, b := range i.bundles {
		for k := range en {
			if _, ok := b[k]; !ok {
				t.Errorf("%s: missing key %q", code, k)
			}
		}
		for k := range b {
			if _, ok := en[k]; !ok {
				t.Errorf("%s: unknown key %q", code, k)
			}
		}
	}
}
