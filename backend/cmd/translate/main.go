package main

import (
	"LuomuTori/internal/translate"
	"encoding/json"
	"os"
	"sort"
	"strings"
)

func getHtmlPages() []string {
	de, err := os.ReadDir("ui/html/pages")
	if err != nil {
		panic(err)
	}

	pages := []string{}
	for _, entry := range de {
		page, err := os.ReadFile("ui/html/pages/" + entry.Name())
		if err != nil {
			panic(err)
		}
		pages = append(pages, string(page))
	}

	de, err = os.ReadDir("ui/html/partials")
	if err != nil {
		panic(err)
	}

	for _, entry := range de {
		page, err := os.ReadFile("ui/html/partials/" + entry.Name())
		if err != nil {
			panic(err)
		}
		pages = append(pages, string(page))

	}

	return pages
}

func parsePhrases(page string) []string {
	const suba = `{{T "`
	const subb = `" $.Lang}}`

	phrases := make([]string, 0)
	for {
		a := strings.Index(page, suba)
		if a == -1 {
			break
		}
		b := strings.Index(page, subb)
		if b == -1 {
			break
		}
		phrases = append(phrases, page[a+len(suba):b])
		page = page[b+len(subb):]
	}

	return phrases
}

// Returns all phrases from html files in lower case and without duplicates
func getPhraseList() []string {
	pages := getHtmlPages()
	phrases := make([]string, 0)
	for _, page := range pages {
		phrases = append(phrases, parsePhrases(page)...)
	}

	for i, phrase := range phrases {
		phrases[i] = strings.ToLower(phrase)
	}

	sort.StringSlice(phrases).Sort()

	pmap := map[string]bool{}
	for _, phrase := range phrases {
		if _, ok := pmap[phrase]; !ok {
			pmap[phrase] = true
		}
	}

	phrases = []string{}
	for key := range pmap {
		phrases = append(phrases, key)
	}

	return phrases
}

func main() {
	phrases := getPhraseList()

	translations := map[string]translate.Translation{}
	for _, phrase := range phrases {
		translations[phrase] = translate.Translation{
			translate.Fi: "",
			translate.Se: "",
		}
	}

	j, err := json.MarshalIndent(translations, "", " ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("translations.json", j, 0666); err != nil {
		panic(err)
	}
}
