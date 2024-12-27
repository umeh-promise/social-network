package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"github.com/umeh-promise/social/docs" // This is required to generate the swagger docs
	"github.com/umeh-promise/social/internal/store"
)

type application struct {
	config config
	store  store.Storage
}

type config struct {
	addr   string
	db     dbConfig
	env    string
	apiURL string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *application) mount() *chi.Mux {
	router := chi.NewRouter()

	// A good base middleware stack
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	router.Use(middleware.Timeout(60 * time.Second))

	router.Route("/v1", func(router chi.Router) {
		router.Get("/health", app.healthCheckHandler)

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		router.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		router.Route("/posts", func(router chi.Router) {
			router.Post("/", app.createPostHandler)

			router.Route("/{id}", func(router chi.Router) {
				router.Use(app.postMiddlewareHandler)
				router.Get("/", app.getPostHandler)
				router.Patch("/", app.updatePostHandler)
				router.Delete("/", app.deletePostHandler)
			})
		})

		router.Route("/users", func(router chi.Router) {
			router.Route("/{id}", func(router chi.Router) {
				router.Use(app.userMiddlewareHandler)
				router.Get("/", app.getUserHandler)
				router.Put("/follow", app.followUserHandler)
				router.Put("/unfollow", app.unfollowUserHandler)
			})

			router.Group(func(r chi.Router) {
				router.Get("/feed", app.getUserFeedHandler)
			})
		})
	})

	return router
}

func (app *application) run(mux *chi.Mux) error {
	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	server := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server is running at port %s", app.config.addr)

	return server.ListenAndServe()
}
