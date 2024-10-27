package main

import (
	"github.com/gin-gonic/gin"
	"github.com/misakimei123/redbook/interactive/events"
	"github.com/misakimei123/redbook/pkg/grpcx"
)

type App struct {
	consumers []events.Consumer
	server    *grpcx.Server
	web       *gin.Engine
}
