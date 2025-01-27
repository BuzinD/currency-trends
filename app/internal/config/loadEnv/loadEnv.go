package loadEnv

import (
	"cur/internal/config/dbConfig"
	"cur/internal/config/okxConfig"
)

func LoadEnvs() {
	okxConfig.LoadEnv()
	dbConfig.LoadEnv()
}
