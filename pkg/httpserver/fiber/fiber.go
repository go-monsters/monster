package fiber

import (
	"encoding/json"

	"github.com/go-monsters/monster/internals/logs/merror"
	"github.com/go-monsters/monster/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

const defaultAddress = ":9000"

type HttpServer struct {
	app     *fiber.App
	address string
}

func NewFiberHttpServer() httpserver.HttpServer {
	return &HttpServer{
		app: fiber.New(),
	}
}

func (h *HttpServer) GetClient() interface{} {
	return h.app
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

	if err := h.app.Listen(h.address); err != nil {
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

	if err := h.app.ListenTLS(h.address, certFile, keyFile); err != nil {
		return err
	}

	return nil
}

func init() {
	httpserver.RegisterNewHttpServerImpl("fiber", NewFiberHttpServer)
}
