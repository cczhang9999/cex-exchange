//go:build wireinject
// +build wireinject

package main

import (
	"backend-v2/internal/biz"
	"backend-v2/internal/conf"
	"backend-v2/internal/data"
	"backend-v2/internal/server"
	"backend-v2/internal/service"

	"github.com/google/wire"
)

func initApp(bc *conf.Bootstrap) (*server.Server, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		data.UserProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		server.NewServer,
	))
}
