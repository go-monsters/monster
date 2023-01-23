package httpserver

import (
	"github.com/go-monsters/monster/internals/logs/merror"
)

type HttpServer interface {
	GetClient() interface{}
	ListenAndServe(config string) error
	ListenAndServeTLS(config, certFile, keyFile string) error
}

type Instance func() HttpServer

var impls = make(map[string]Instance)

func RegisterNewHttpServerImpl(name string, impl Instance) {
	if impl == nil {
		panic(merror.Error("httpServer: Register adapter is nil").Error())
	}
	if _, ok := impls[name]; ok {
		panic("httpServer: Register called twice for adapter " + name)
	}
	impls[name] = impl
}

func NewHttpServer(implName string) (adapter HttpServer, err error) {
	instanceFunc, ok := impls[implName]
	if !ok {
		err = merror.Errorf("cache: unknown impl name %s (forgot to import?)", implName)
		return
	}
	return instanceFunc(), nil
}
