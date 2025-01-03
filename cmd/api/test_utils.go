package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/umeh-promise/social/internal/auth"
	"github.com/umeh-promise/social/internal/store"
	"github.com/umeh-promise/social/internal/store/cache"
	"go.uber.org/zap"
)

func newTestApplication(t testing.TB) *application {
	t.Helper()
	logger := zap.NewNop().Sugar()
	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewMockStore()
	testAuth := &auth.TestAuthenticator{}

	return &application{
		logger:        logger,
		store:         mockStore,
		cacheStorage:  mockCacheStore,
		authenticator: testAuth,
	}
}

func executeRequest(req *http.Request, mux *chi.Mux) *httptest.ResponseRecorder {
	reqRecorder := httptest.NewRecorder()
	mux.ServeHTTP(reqRecorder, req)

	return reqRecorder
}

func checkResponseCode(t testing.TB, expected, actual int) {
	t.Helper()

	if expected != actual {
		t.Errorf("expcted error code %d. got %d", expected, actual)
	}
}
