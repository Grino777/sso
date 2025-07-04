package admin

import (
	"fmt"
	"net/http"

	"github.com/Grino777/sso/internal/services/keys/manager"
	"github.com/gin-gonic/gin"
)

// keysStore интерфейс для работы с ключами
type keysStore interface {
	RotateKeys() (*manager.GenKeys, error)
}

// Route представляет маршрут для API
type Route struct {
	method  string
	path    string
	handler gin.HandlerFunc
}

// Routes представляет набор маршрутов для API
type Routes struct {
	keysStore keysStore
	routes    []*Route
}

// NewRoutes создает новый набор маршрутов
func NewRoutes(keysStore keysStore) *Routes {
	r := &Routes{
		keysStore: keysStore,
	}
	r.routes = r.initRoutes()
	return r
}

// RegisterRoutes регистрирует маршруты в GIN-сервере
func (r *Routes) RegisterRoutes(engine *gin.Engine) {
	for _, route := range r.routes {
		engine.Handle(route.method, route.path, route.handler)
	}
}

// initRoutes инициализирует набор маршрутов
func (r *Routes) initRoutes() []*Route {
	return []*Route{
		{
			method:  "GET",
			path:    "/ping",
			handler: r.ping,
		},
		{
			method:  "POST",
			path:    "/rotate-keys",
			handler: r.rotateKeys,
		},
	}
}

func (r *Routes) ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "PONG",
	})
}

func (r *Routes) rotateKeys(c *gin.Context) {
	if _, err := r.keysStore.RotateKeys(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to rotate keys: %v", err)})
	}
	c.JSON(200, gin.H{"message": "keys rotated"})
}
