package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusInternalServerError, "the sever encountred a problem")
}

func (app *application) forbiddenResponseError(w http.ResponseWriter, r *http.Request) {
	app.logger.Warnw("forbidden error", "method", r.Method, "path", r.URL.Path, "error", "error")
	writeJSONError(w, http.StatusForbidden, "forbidden")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("bad request error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("conflict request error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusConflict, err.Error())
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusNotFound, "not found")
}

func (app *application) unathorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized", "method", r.Method, "path", r.URL.Path, "error", "unauthorized error")
	writeJSONError(w, http.StatusUnauthorized, err.Error())
}

func (app *application) unathorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized", "method", r.Method, "path", r.URL.Path, "error", "unauthorized error")

	w.Header().Set("WWW-Authenticate", `basic realm="restricted", charset="UTF-8"`)

	writeJSONError(w, http.StatusUnauthorized, err.Error())
}
