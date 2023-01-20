package nethttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-monsters/monster/internals/logs/merror"
	"github.com/go-monsters/monster/pkg/httpserver"
)

const defaultAddress = ":9000"

type HttpServer struct {
	http    *http.Server
	address string
}

func NewNetHttpServer() httpserver.HttpServer {
	return &HttpServer{
		http: &http.Server{},
	}
}

func (h *HttpServer) GetClient() interface{} {
	return h.http
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

	if err := http.ListenAndServe(h.address, nil); err != nil {
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

	if err := http.ListenAndServeTLS(h.address, certFile, keyFile, nil); err != nil {
		return err
	}

	return nil
}

func init() {
	httpserver.RegisterNewHttpServerImpl("net/http", NewNetHttpServer)
}
