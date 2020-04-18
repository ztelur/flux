package server

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"os"
	"os/signal"
	"time"
)

func InitDefaultLogger() {
	l, err := InitLogger()
	if err != nil && l != nil {
		l.Panic("FluxServer logger init:", err)
	} else {
		ext.SetLogger(l)
	}
	if nil == l {
		panic("logger is nil")
	}
}

func Run(ver flux.BuildInfo) {
	fx := NewFluxServer()
	globals := LoadConfig()
	if err := fx.Prepare(globals); nil != err {
		logger.Panic("FluxServer prepare:", err)
	}
	if err := fx.Init(globals); nil != err {
		logger.Panic("FluxServer init:", err)
	}
	go func() {
		if err := fx.Startup(ver); nil != err {
			logger.Error(err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := fx.Shutdown(ctx); nil != err {
		logger.Error(err)
	}
}
