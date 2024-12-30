package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"github.com/umeh-promise/social/docs" // This is required to generate the swagger docs
	"github.com/umeh-promise/social/internal/auth"
	"github.com/umeh-promise/social/internal/mailer"
	"github.com/umeh-promise/social/internal/store"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         store.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	apiURL      string
	mail        mailConfig
	frontendURL string
	auth        authConfig
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	issuer string
}

type basicConfig struct {
	username string
	password string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type mailConfig struct {
	exp       time.Duration
	fromEmail string
	sendGrid  sendGridConfig
}

type sendGridConfig struct {
	apikey string
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
		router.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler)

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		router.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		router.Route("/posts", func(router chi.Router) {
			router.Use(app.AuthTokenMiddleware)
			router.Post("/", app.createPostHandler)

			router.Route("/{id}", func(router chi.Router) {
				router.Use(app.postMiddlewareHandler)
				router.Get("/", app.getPostHandler)
				router.Patch("/", app.updatePostHandler)
				router.Delete("/", app.deletePostHandler)
			})
		})

		router.Route("/users", func(router chi.Router) {
			router.Put("/activate/{token}", app.activateHandler)

			router.Route("/{id}", func(router chi.Router) {
				router.Use(app.AuthTokenMiddleware)
				router.Use(app.userMiddlewareHandler)
				router.Get("/", app.getUserHandler)
				router.Put("/follow", app.followUserHandler)
				router.Put("/unfollow", app.unfollowUserHandler)
			})

			router.Group(func(r chi.Router) {
				router.With(app.AuthTokenMiddleware).Get("/feed", app.getUserFeedHandler)
			})
		})

		// Public routes
		router.Route("/auth", func(router chi.Router) {
			router.Post("/user", app.registerUserHandler)
			router.Post("/token", app.createTokenHandler)
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

	app.logger.Infow("Server has started", "addr", app.config.addr, "env", app.config.env)

	return server.ListenAndServe()
}
