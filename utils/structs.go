// © 2022 Sloan Childers
package utils

import (
	"time"

	"github.com/osintami/plumbr/sink"
)

type ICache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, duration time.Duration)
}

type ServerCore struct {
	Config   ServerConfig
	Cache    ICache
	Secrets  *sink.SecretsManager
	Shutdown *sink.ShutdownHandler
	//Rules    *engine.RulesEngine
}

type ServerConfig struct {
	CookieDomain string `env:"COOKIE_DOMAIN" envDefault:"a.osintami.com"`
	FSPath       string `env:"LOCAL_FILE_PATH" envDefault:"./external/"`
	PathPrefix   string `env:"PATH_PREFIX" envDefault:""`
	ListenAddr   string `env:"LISTEN_ADDR,required" envDefault:"127.0.0.1:8080"`
	LogLevel     string `env:"LOG_LEVEL" envDefault:"TRACE"`
	GinMode      string `env:"GIN_MODE" envDefault:"RELEASE"`
	DgraphSvr    string `env:"DGRAPH_SVR" envDefault:"localhost:9080"`
	DgraphUser   string `env:"DGRAPH_USER" envDefault:"groot"`
	DgraphPass   string `env:"DGRAPH_PASS" envDefault:"password"`
}
