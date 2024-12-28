package translate

import (
	"LuomuTori/internal/log"
	"LuomuTori/internal/model"
	"encoding/json"
	"os"
	"strings"
)

const (
	En string = "en"
	Fi string = "fi"
	Se string = "se"
)

// translation for each language
type Translation map[string]string

var translations = map[string]Translation{}

func LoadTranslations() error {
	data, err := os.ReadFile("translations.json")
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &translations); err != nil {
		return err
	}

	return nil
}

// automatic capitalization
func T(text any, lang string) string {
	var word string
	switch text.(type) {
	case string:
		word = text.(string)
	case model.OrderStatus:
		word = string(text.(model.OrderStatus))
	}

	if lang == En {
		return word
	}
	if len(word) == 0 {
		log.Error.Printf("ERROR: Translation: T with %s\n", word)
		return word
	}

	tmap, ok := translations[strings.ToLower(word[0:1])+word[1:]]
	if !ok {
		log.Error.Printf("ERROR: Translation: translation not found for %s\n", word)
		return word
	}

	translation, ok := tmap[lang]
	if !ok {
		log.Error.Printf("ERROR: Translation: translation not found for %s\n", word)
		return word
	}

	isCapitalized := word[0:1] == strings.ToUpper(word[0:1])
	if isCapitalized {
		return strings.ToUpper(translation[0:1]) + translation[1:]
	} else {
		return translation
	}
}
