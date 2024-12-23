package main

import (
	"cur/internal/config/loadEnv"
	"cur/internal/config/okxConfig"
	"fmt"
	"log"
)

func main() {
	initApp()
	conf, err := okxConfig.GetOkxApiConfig()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("config: %s %s %s", conf.ApiKey, conf.Secret, conf.PassPhrase)
}

func initApp() {
	loadEnv.LoadEnvs()
}
