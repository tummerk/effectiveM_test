package server

import (
	"effectiveMobile_test/internal/server/api"
	"effectiveMobile_test/pkg/middlewarex"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

func NewRouter(svc SubscriptionsService) http.Handler {
	h := NewHandler(svc)
	strictHandler := api.NewStrictHandler(h, nil)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middlewarex.Logger)
	r.Use(middleware.Timeout(60 * time.Second))

	api.HandlerFromMux(strictHandler, r)

	return r
}
