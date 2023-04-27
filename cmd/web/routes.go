package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Config) routes() http.Handler {
	// create router
	r := chi.NewRouter()

	// set up middleware
	r.Use(middleware.Recoverer)
	r.Use(app.SessionLoad)

	// define application routes
	r.Get("/", app.HomePage)
	r.Get("/login", app.LoginPage)
	r.Post("/login", app.PostLoginPage)
	r.Get("/logout", app.Logout)
	r.Get("/register", app.RegisterPage)
	r.Post("/register", app.PostRegisterPage)
	r.Get("/activate", app.ActivateAccount)

	r.Mount("/members", app.authRouter())

	return r
}

func (app *Config) authRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(app.Auth)

	r.Get("/plans", app.ChooseSubscription)
	r.Get("/subscribe", app.SubscribeToPlan)

	return r
}
