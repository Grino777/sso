package apiserver

import (
	"github.com/gin-gonic/gin"
)

var routes = []struct {
	method  string
	path    string
	handler gin.HandlerFunc
}{
	{"GET", "/ping", ping},
}

var ping = func(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "PONG",
	})
}
