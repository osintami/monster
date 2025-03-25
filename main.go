// © 2022 Sloan Childers
package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/osintami/monster/server"
	"github.com/osintami/monster/utils"
	"github.com/osintami/plumbr/sink"
)

func main() {
	svrConfig := utils.ServerConfig{}
	sink.LoadEnv(&svrConfig)
	sink.InitLogger(svrConfig.LogLevel)

	// TODO:  make this configuration driven between DynamoDB, Redis, etc.
	cache := sink.NewFastCache(svrConfig.FSPath + "cache.db")
	cache.LoadFile()

	shutdown := sink.NewShutdownHandler()
	shutdown.AddListener(cache.SaveFile)
	shutdown.Listen()

	core := utils.ServerCore{
		Config:   svrConfig,
		Cache:    cache,
		Secrets:  LoadSecrets(),
		Shutdown: shutdown,
		//Rules:    engine.NewRulesEngine(svrConfig.FSPath),
	}

	in := server.NewServer(core)

	router := chi.NewMux()
	router.Route(svrConfig.PathPrefix, func(r chi.Router) {
		r.Get("/csr", in.CookieSync)
	})

	http.ListenAndServe(svrConfig.ListenAddr, router)
}

func LoadSecrets() *sink.SecretsManager {
	keys := []string{
		"DGRAPH_USER",
		"DGRAPH_PASS"}
	return sink.NewSecretsManager(keys)
}
