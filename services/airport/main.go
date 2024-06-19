package main

import (
	//"context"
	//"encoding/json"
	"fmt"
	//"path/filepath"

	"net/http"
	"os"

	"github.com/spf13/viper"

	sb "github.com/axodevelopment/servicebase"
	//u "github.com/axodevelopment/servicebase/pkg/utils"
	"github.com/gin-gonic/gin"

	"strings"
)

type Config struct {
	Port int
	UKey string
}

const (
	SERVICE_NAME = "AIRPORT"
	VERSION      = "v1"
	PROJECT      = "TUTORIAL-SERVICES"
)

var (
	CONFIG *Config

	APP_READY chan struct{}
)

func log(args ...string) {
	var builder strings.Builder

	for i := range args {
		builder.WriteString(args[i])

		if i < (len(args) - 1) {
			builder.WriteString(" / ")
		}
	}

	fmt.Println(SERVICE_NAME, ": ", builder.String())
}

func main() {
	defer log("Application Exiting...")
	log("Application Starting...")

	var svc *sb.Service
	var err error

	CONFIG, err = loadConfig()

	initSvc()

	validateSvc(CONFIG)

	log("Initializing Service")
	svc, err = sb.New(SERVICE_NAME, sb.WithPort(CONFIG.Port), sb.WithHealthProbe(true), sb.WithCORS(true))

	if err != nil {
		log(err.Error())
		panic(err)
	}

	go func(svc *sb.Service) {
		log("Starting service core logic")
		serviceLogic(svc)
	}(svc)

	<-APP_READY

	go func(s *sb.Service) {
		sb.Start(s)
	}(svc)

}

func initSvc() {
	APP_READY = make(chan struct{})
}

func startSvc(svc *sb.Service) {
	defer fmt.Println("Start Svc... Done")
	fmt.Println("Start Svc...")

	svc.AppHealthz = true
	svc.AppReadyz = true
}

func serviceLogic(svc *sb.Service) {

	svc.GinEngine.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, os.Args)
	})

	close(APP_READY)
}

func loadConfig() (*Config, error) {
	viper.SetEnvPrefix("APP")

	viper.BindEnv("PORT")
	viper.BindEnv("UKEY")

	viper.AutomaticEnv()

	config := &Config{
		Port: viper.GetInt("PORT"),
		UKey: viper.GetString("UKEY"),
	}

	if config.Port <= 0 {
		fmt.Println("OsEnvVar NotFound - [APP_PORT] => defaulted to 8080")
		config.Port = 8080
	}

	fmt.Println(config)

	//for now nil error in the future validation would could prevent panic and work in a limited state ie a db connection or something
	return config, nil
}

func validateSvc(cfg *Config) {
	if cfg.Port <= 0 {
		panic("Port should be greater then 0.")
	}
}
