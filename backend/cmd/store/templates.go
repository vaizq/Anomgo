package main

import (
	"LuomuTori/internal/model"
	"LuomuTori/internal/service/payment"
	"LuomuTori/internal/translate"
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// Provides anonymity for reviewing users by only returning first three letters and length of their username
func Obfuscate(username string) string {
	n := len(username)
	if n <= 3 {
		return "***"
	}

	return username[0:3] + func() string {
		b := new(strings.Builder)
		for i := 0; i < n-3; i++ {
			b.WriteRune('*')
		}
		return b.String()
	}()
}

// Retuns the head of text up to maxlen, cut from the last space
// eq. Head("hello wonderfull world", 18) => "hello wonderfull..."
func Head(text string, maxlen int) string {
	runes := []rune(text)
	if len(runes) <= maxlen {
		return text
	}

	result := string(runes[:maxlen])
	if end := strings.LastIndex(result, " "); end != -1 {
		return result[0:end] + "..."
	}

	return result + "..."
}

func Iterate(start, end int) []int {
	items := make([]int, 0)
	for i := range end - start {
		items = append(items, i+start)
	}
	return items
}

func FmtTime(t time.Time) string {
	return t.Format("15:04:05 02-01-2006")
}

func FmtDate(t time.Time) string {
	return t.Format(time.DateOnly)
}

func NewTemplateCache() (map[string]*template.Template, error) {
	tc := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		ts := template.New("base")

		ts = ts.Funcs(map[string]any{
			"Obfuscate":   Obfuscate,
			"XMR2Decimal": payment.XMR2Decimal,
			"Fiat2XMR":    payment.Fiat2XMR,
			"XMR2Fiat": func(xmr uint64) int {
				return int(payment.XMR2Fiat(xmr))
			},
			"T":       translate.T,
			"Head":    Head,
			"Iterate": Iterate,
			"FmtTime": FmtTime,
			"FmtDate": FmtDate,
		})

		ts, err := ts.ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		tc[filepath.Base(page)] = ts
	}

	return tc, nil
}

func (app *application) render(w http.ResponseWriter, req *http.Request, status int, page string, data interface{}) {
	tc, found := app.templateCache[page]
	if !found {
		err := fmt.Errorf("no page named %s", page)
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	if data == nil {
		data = app.newTemplateData(req, nil)
	}

	err := tc.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

type Note struct {
	Message string
	IsError bool
}

type templateData struct {
	Form  any
	Data  map[string]any
	User  *model.User
	Lang  string
	Notes []Note
}

func (app *application) newTemplateData(req *http.Request, data map[string]any) *templateData {
	if data == nil {
		data = make(map[string]any)
	}

	user := app.loggedInUser(req)
	if user != nil {
		wallet, _ := model.M.Wallet.GetForUser(app.db, user.ID)
		if wallet != nil {
			data["wallet"] = wallet
		}
		data["isVendor"] = model.M.User.IsVendor(app.db, user.ID)
	} else {
		data["isVendor"] = false
	}

	lang := func() string {
		res := app.sessionManager.GetString(req.Context(), "Lang")
		if res == "" {
			res = translate.En
			app.sessionManager.Put(req.Context(), "Lang", res)
			return res
		}
		return res
	}()

	// Notes are only viewed once
	notes, _ := app.sessionManager.Pop(req.Context(), "notes").([]Note)

	return &templateData{
		Data:  data,
		User:  user,
		Lang:  lang,
		Notes: notes,
	}
}
