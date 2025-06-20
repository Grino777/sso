package admin

import (
	"github.com/gin-gonic/gin"
)

type keysStore interface {
	RotateKeys() error
}

type Route struct {
	method  string
	path    string
	handler gin.HandlerFunc
}

type Routes struct {
	keysStore keysStore
	routes    []*Route
}

func NewRoutes(keysStore keysStore) *Routes {
	r := &Routes{
		keysStore: keysStore,
	}
	r.routes = r.initRoutes()
	return r
}

func (r *Routes) RegisterRoutes(engine *gin.Engine) {
	for _, route := range r.routes {
		engine.Handle(route.method, route.path, route.handler)
	}
}

func (r *Routes) initRoutes() []*Route {
	return []*Route{
		{
			method:  "GET",
			path:    "/ping",
			handler: r.ping,
		},
	}
}

func (r *Routes) ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "PONG",
	})
}
