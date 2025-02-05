package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/umeh-promise/social/internal/store"
)

type Userkey string

const userCtx Userkey = "user"

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || userID < 1 {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	// user, err := app.store.Users.GetByID(ctx, userID)
	user, err := app.getUser(ctx, userID)
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

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	followedUserID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	ctx := r.Context()

	if err := app.store.Followers.Follow(ctx, user.ID, followedUserID); err != nil {
		switch err {
		case store.ErrorConflict:
			app.conflictResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	unfollowedUserID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	ctx := r.Context()
	if err := app.store.Followers.Unfollow(ctx, user.ID, unfollowedUserID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) activateHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrorNotFound:
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, ""); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// func (app *application) userMiddlewareHandler(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		userIdParams := chi.URLParam(r, "id")
// 		userId, err := strconv.ParseInt(userIdParams, 10, 64)
// 		if err != nil {
// 			app.badRequestResponse(w, r, err)
// 			return
// 		}

// 		ctx := r.Context()
// 		user, err := app.store.Users.GetByID(ctx, userId)
// 		if err != nil {
// 			switch err {
// 			case store.ErrorNotFound:
// 				app.notFoundResponse(w, r, err)
// 			default:
// 				app.internalServerError(w, r, err)
// 			}
// 			return
// 		}
// 		ctx = context.WithValue(ctx, userCtx, user)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

func getUserFromContext(r *http.Request) *store.User {
	user := r.Context().Value(userCtx).(*store.User)

	return user
}
