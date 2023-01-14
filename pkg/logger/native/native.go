package native

import (
	"log"

	"github.com/go-monsters/monster/pkg/logger"
)

type Native struct {
}

func New() logger.Logger {
	return &Native{}
}

func (nl *Native) Info(msg string, params ...interface{}) {
	log.Printf(msg, params...)
}

func (nl *Native) Warn(msg string, params ...interface{}) {
	log.Printf(msg, params...)
}

func (nl *Native) Error(msg string, params ...interface{}) {
	log.Printf(msg, params...)
}
