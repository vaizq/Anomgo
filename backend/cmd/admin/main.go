package main

import (
	"LuomuTori/internal/config"
	"LuomuTori/internal/log"
	"LuomuTori/internal/service/captcha"
	"LuomuTori/internal/service/payment"
	"LuomuTori/internal/translate"
	"context"
	"errors"
	"os/signal"
	"sync"
	"syscall"

	"database/sql"
	"encoding/gob"
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/schema"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type application struct {
	db             *sql.DB
	templateCache  map[string]*template.Template
	schemaDecoder  *schema.Decoder
	sessionManager *scs.SessionManager
}

type service struct {
	name     string
	interval time.Duration
	job      func()
}

func main() {
	log.Init()

	gob.Register(captcha.Solution{})
	gob.Register(uuid.UUID{})

	if err := translate.LoadTranslations(); err != nil {
		log.Error.Fatalf("Failed to load translations: %s\n", err.Error())
	}

	config.Parse()

	db, err := openDB(config.DSN)
	if err != nil {
		log.Error.Fatal(err)
	}
	defer db.Close()

	tc, err := NewTemplateCache()
	if err != nil {
		log.Error.Fatal(err)
	}

	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)

	app := application{
		db:             db,
		templateCache:  tc,
		schemaDecoder:  schema.NewDecoder(),
		sessionManager: sessionManager,
	}

	srv := http.Server{
		Addr:     config.Addr,
		ErrorLog: log.Error,
		Handler:  app.route(),
	}

	if err := payment.UpdateXMRPrice(); err != nil {
		log.Error.Fatalf("Failed to update XMR price: %s\n", err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	services := []service{
		{
			name:     "XMR price",
			interval: time.Hour,
			job: func() {
				if err := payment.UpdateXMRPrice(); err != nil {
					log.Error.Fatalf("Failed to update XMR price: %s\n", err.Error())
				}
			},
		},
	}

	var wg sync.WaitGroup
	for _, service := range services {
		wg.Add(1)
		go func() {
			quit := false
			for !quit {
				select {
				case <-ctx.Done():
					quit = true
					break
				case <-time.After(service.interval):
					log.Debug.Printf("Service %s job running...\n", service.name)
					service.job()
					log.Debug.Printf("Service %s job done.\n", service.name)
				}
			}
			log.Debug.Printf("Service %s shutdown.\n", service.name)
			wg.Done()
		}()
	}

	wg.Add(1)
	go func() {
		log.Info.Printf("server listening on address %s\n", config.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error.Fatalf("Server failed: %v\n", err)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		<-ctx.Done()
		log.Debug.Println("Shutting down servers...")
		if err := srv.Shutdown(context.TODO()); err != nil {
			log.Error.Fatalf("server failed to shutdown: %v\n", err)
		}
		log.Debug.Println("Servers shutdown.")
		wg.Done()
	}()

	wg.Wait()
	log.Info.Println("Bye!")
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
