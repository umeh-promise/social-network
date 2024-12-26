package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/umeh-promise/social/internal/store"
)

type Userkey string

const userKey Userkey = "userKey"

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userIdParams := chi.URLParam(r, "id")
	userId, err := strconv.ParseInt(userIdParams, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	user, err := app.store.Users.GetByID(ctx, userId)
	if err != nil {
		switch err {
		case store.ErrorNotFound:
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {}

func (app *application) userMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIdParams := chi.URLParam(r, "id")
		userId, err := strconv.ParseInt(userIdParams, 10, 64)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}

		ctx := r.Context()
		user, err := app.store.Users.GetByID(ctx, userId)
		if err != nil {
			switch err {
			case store.ErrorNotFound:
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, userKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
