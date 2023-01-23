package mux

import (
	"encoding/json"
	"net/http"

	"github.com/go-monsters/monster/internals/logs/merror"
	"github.com/go-monsters/monster/pkg/httpserver"

	"github.com/gorilla/mux"
)

const defaultAddress = ":9000"

type HttpServer struct {
	router  *mux.Router
	address string
}

func NewMuxHttpServer() httpserver.HttpServer {
	return &HttpServer{
		router: mux.NewRouter(),
	}
}

func (h *HttpServer) GetClient() interface{} {
	return h.router
}

func (h *HttpServer) ListenAndServe(config string) error {
	var cf map[string]string
	err := json.Unmarshal([]byte(config), &cf)
	if err != nil {
		return merror.Wrapf(err, "could not unmarshal the config: %s", config)
	}
	if _, ok := cf["address"]; !ok {
		cf["address"] = defaultAddress
	}

	h.address = cf["address"]

	srv := &http.Server{
		Handler: h.router,
		Addr:    defaultAddress,
	}

	if err := srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func (h *HttpServer) ListenAndServeTLS(config, certFile, keyFile string) error {
	var cf map[string]string
	err := json.Unmarshal([]byte(config), &cf)
	if err != nil {
		return merror.Wrapf(err, "could not unmarshal the config: %s", config)
	}
	if _, ok := cf["address"]; !ok {
		cf["address"] = defaultAddress
	}

	h.address = cf["address"]

	srv := &http.Server{
		Handler: h.router,
		Addr:    defaultAddress,
	}

	if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
		return err
	}

	return nil
}

func init() {
	httpserver.RegisterNewHttpServerImpl("mux", NewMuxHttpServer)
}
