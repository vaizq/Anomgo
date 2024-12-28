package main

import (
	"LuomuTori/internal/log"
	"LuomuTori/internal/model"
	"LuomuTori/internal/model/view"
	"LuomuTori/internal/service/dispute"
	"github.com/google/uuid"
	"net/http"
)

const (
	maxMemory = 10 << 20
)

func (app *application) servePage(page string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.render(w, r, http.StatusOK, page, nil)
	}
}

func (app *application) admin(w http.ResponseWriter, r *http.Request) {

	disputes, err := model.M.Dispute.GetAll(app.db)
	if err != nil {
		app.serverError(w, err)
		return
	}

	tickets, err := model.M.Ticket.GetAll(app.db)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r, map[string]any{
		"disputes": disputes,
		"tickets":  tickets,
	})
	app.render(w, r, http.StatusOK, "admin.html", data)
}

func (app *application) handleOperation(w http.ResponseWriter, r *http.Request) {
	form := deleteForm{}
	if err := app.decodeForm(&form, r); err != nil {
		app.serverError(w, err)
		return
	}

	switch form.Operation {
	case "banUser":
		if _, err := model.M.Ban.Create(app.db, form.ID); err != nil {
			app.serverError(w, err)
			return
		}
	case "deleteBan":
		if err := model.M.Ban.Delete(app.db, form.ID); err != nil {
			app.serverError(w, err)
			return
		}
	case "deleteListing":
		if err := model.M.Product.Delete(app.db, form.ID); err != nil {
			app.serverError(w, err)
			return
		}
	case "deleteReview":
		if err := model.M.Review.Delete(app.db, form.ID); err != nil {
			app.serverError(w, err)
			return
		}
	default:
		log.Error.Printf("Unknown operation: %s\n", form.Operation)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) dispute(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Info.Println("unable to parse id")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dispute, err := model.M.Dispute.Get(app.db, id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	counter, err := model.M.CounterDispute.GetForDispute(app.db, id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	order, err := view.V.Order.Get(app.db, dispute.OrderID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r, map[string]any{
		"dispute": dispute,
		"counter": counter,
		"order":   order,
	})
	app.render(w, r, http.StatusOK, "handle-dispute.html", data)
}

func (app *application) handleDispute(w http.ResponseWriter, r *http.Request) {
	form := disputeForm{}
	if err := app.decodeForm(&form, r); err != nil {
		app.serverError(w, err)
		return
	}

	if _, err := dispute.CreateDisputeDecision(app.db, form.DisputeID, form.Outcome, form.Reason); err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) ticket(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Info.Printf("unable to parse id from query url %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ticket, err := view.V.Ticket.Get(app.db, id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r, map[string]any{
		"ticket": ticket,
	})
	app.render(w, r, http.StatusOK, "handle-ticket.html", data)
}

func (app *application) handleTicket(w http.ResponseWriter, r *http.Request) {
	form := ticketResponseForm{}
	if err := app.decodeForm(&form, r); err != nil {
		app.serverError(w, err)
		return
	}

	log.Info.Printf("received ticket response: %v\n", form)

	tx, err := app.db.Begin()
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer tx.Rollback()

	if _, err := model.M.TicketResponse.Create(tx, form.Message, form.TicketID, "admin"); err != nil {
		app.serverError(w, err)
		return
	}

	if form.CloseTicket {
		if _, err := model.M.Ticket.Close(tx, form.TicketID); err != nil {
			app.serverError(w, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		app.serverError(w, err)
		return
	}

	app.redirectBack(w, r)
}
