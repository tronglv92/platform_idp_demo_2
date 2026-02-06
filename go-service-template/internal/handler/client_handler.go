package handler

import (
	"go-service-template/helper/server/http/response"
	"go-service-template/internal/registry"
	"net/http"
)

type ClientHandler interface {
	Detail() http.HandlerFunc
	Health() http.HandlerFunc
}

type clientHandler struct {
	reg *registry.ServiceContext
}

func NewClientHandler(reg *registry.ServiceContext) ClientHandler {
	return &clientHandler{
		reg: reg,
	}
}

func (p *clientHandler) Detail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"message": "hello",
		}
		response.OkJson(r.Context(), w, resp)
	}
}
func (p *clientHandler) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"status": "up",
		}
		response.OkJson(r.Context(), w, resp)
	}
}
