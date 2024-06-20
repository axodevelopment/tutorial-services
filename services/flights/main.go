package main

import (
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"os"

	sb "github.com/axodevelopment/servicebase"
	"github.com/gin-gonic/gin"

	"strings"
)

type Config struct {
	Port int
	UKey string
}

const (
	SERVICE_NAME = "FLIGHTS"
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

	APP_READY = make(chan struct{})

	CONFIG, err = loadConfig()

	if err != nil {
		log(err.Error())
		panic(err)
	}

	validateSvc(CONFIG)

	log("Initializing Service")
	svc, err = sb.New(SERVICE_NAME, sb.WithPort(CONFIG.Port), sb.WithHealthProbe(true), sb.WithCORS(true))

	if err != nil {
		log(err.Error())
		panic(err)
	}

	go func(svc *sb.Service) {
		log("Initializing service")
		initSvc(svc)

		log("Starting service core logic")
		serviceLogic(svc)

		log("Starting service")
		startSvc(svc)
	}(svc)

	<-APP_READY

	go func(s *sb.Service) {
		sb.Start(s)
	}(svc)

	<-svc.ExitAppChan
}

func initSvc(svc *sb.Service) {
	initUsrSvc()

}

func startSvc(svc *sb.Service) {
	defer fmt.Println("Start Svc... Done")
	fmt.Println("Start Svc...")

	svc.AppHealthz = true
	svc.AppReadyz = true
}

func serviceLogic(svc *sb.Service) {
	startUsrLogic(svc)

	close(APP_READY)
}

// User Service Init
func initUsrSvc() {
	defer log("Init Svc... Done")
	log("Init Svc...")

}

// User logic
func startUsrLogic(svc *sb.Service) {
	svc.GinEngine.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, os.Args)
	})
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
		log("OsEnvVar NotFound - [APP_PORT] => defaulted to 8080")
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
