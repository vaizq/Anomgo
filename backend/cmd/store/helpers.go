package main

import (
	"LuomuTori/internal/log"
	"LuomuTori/internal/model"
	"LuomuTori/internal/service/captcha"
	"context"
	"fmt"
	"github.com/google/uuid"
	"math"
	"net/http"
	"net/url"
	"runtime/debug"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	log.Info.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) redirectBack(w http.ResponseWriter, r *http.Request) {
	if res, err := url.Parse(r.Referer()); err == nil {
		http.Redirect(w, r, res.RequestURI(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// NOTE: Check that each handle where this is called is wrapped in requireAuth middleware
func (app *application) loggedInUser(req *http.Request) *model.User {
	ctx := req.Context()
	key := "userID"
	if app.sessionManager.Exists(ctx, key) {
		uid, ok := app.sessionManager.Get(ctx, key).(uuid.UUID)
		if !ok {
			return nil
		}
		user, err := model.M.User.Get(app.db, uid)
		if err != nil {
			return nil
		}
		return user
	}

	return nil
}

func (app *application) renderInvalidForm(w http.ResponseWriter, r *http.Request, page string, form any) {
	data := app.newTemplateData(r, nil)
	data.Form = form
	app.render(w, r, http.StatusBadRequest, page, data)
}

func (app *application) renderInvalidFormWithData(w http.ResponseWriter, r *http.Request, page string, form any, data *templateData) {
	if data == nil {
		data = app.newTemplateData(r, nil)
	}
	data.Form = form
	app.render(w, r, http.StatusBadRequest, page, data)
}

func (app *application) decodeForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.schemaDecoder.Decode(dst, r.PostForm)
	if err != nil {
		return err
	}
	return nil
}

func (app *application) validateCaptcha(ctx context.Context, answer CaptchaAnswer) bool {
	solution, ok := app.sessionManager.Get(ctx, "captchaSolution").(captcha.Solution)
	if !ok {
		log.Info.Println("Unable to read captcha solution from session store")
		return false
	}

	return math.Pow(float64(solution.X-answer.X), 2)+math.Pow(float64(solution.Y-answer.Y), 2) < math.Pow(float64(solution.Radius), 2)
}

func (app *application) addNotes(ctx context.Context, notes ...string) {
	const key = "notes"
	tmp, ok := app.sessionManager.Get(ctx, key).([]string)
	if !ok {
		tmp = notes
	} else {
		tmp = append(tmp, notes...)
	}

	app.sessionManager.Put(ctx, key, tmp)
}
