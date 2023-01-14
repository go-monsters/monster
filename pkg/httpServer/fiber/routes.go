package fiber

import (
	"github.com/go-monsters/monster/pkg/httpServer"
)

func (fh *FiberHttpServer) SetRouteGroups(groupName string, routes []httpServer.Route) {
	g := fh.app.Group("/" + groupName)
	for _, route := range routes {
		g.Add(string(route.Method), route.Path, route.Handler)
	}
}
