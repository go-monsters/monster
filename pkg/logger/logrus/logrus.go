package logrus

import (
	"github.com/go-monsters/monster/pkg/logger"
	"github.com/sirupsen/logrus"
)

type Logrus struct {
}

func New() logger.Logger {
	return &Logrus{}
}

func (nl *Logrus) Info(msg string, params ...interface{}) {
	logrus.Infof(msg, params...)
}

func (nl *Logrus) Warn(msg string, params ...interface{}) {
	logrus.Warnf(msg, params...)
}

func (nl *Logrus) Error(msg string, params ...interface{}) {
	logrus.Errorf(msg, params...)
}
