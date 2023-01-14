package fiber

import (
	sentryfiber "github.com/aldy505/sentry-fiber"
	"github.com/go-monsters/monster/pkg/httpServer"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"go.elastic.co/apm/module/apmfiber/v2"
)

type FiberHttpServer struct {
	app     *fiber.App
	address string
}

func New(address string) httpServer.HttpServer {
	if address == "" {
		address = "0.0.0.0:3000"
	}

	return &FiberHttpServer{
		app:     fiber.New(),
		address: address,
	}
}

func (fh *FiberHttpServer) ActiveApm() {
	fh.app.Use(apmfiber.Middleware())
}

func (fh *FiberHttpServer) ActiveSwagger() {
	fh.app.Get("/swagger/*", swagger.HandlerDefault)
}

func (fh *FiberHttpServer) ActiveSentry() {
	fh.app.Use(sentryfiber.New(sentryfiber.Options{}))
}

func (fh *FiberHttpServer) Listen() error {
	if err := fh.app.Listen(fh.address); err != nil {
		return err
	}
	return nil
}
