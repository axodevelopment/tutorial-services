package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io/fs"
	"net/http"
	"os"
	"strconv"

	sb "github.com/axodevelopment/servicebase"
	u "github.com/axodevelopment/servicebase/pkg/utils"
	"github.com/gin-gonic/gin"

	"strings"
)

type Config struct {
	Port int
	UKey string
}

type Airport struct {
	Code          string  `json:"code"`
	Lat           string  `json:"lat"`
	Lon           string  `json:"lon"`
	Name          string  `json:"name"`
	City          string  `json:"city"`
	State         string  `json:"state"`
	Country       string  `json:"country"`
	WoeId         string  `json:"woeid"`
	Tz            string  `json:"tz"`
	Phone         string  `json:"phone"`
	Type          string  `json:"type"`
	Email         string  `json:"email"`
	Url           string  `json:"url"`
	RunwayLength  *string `json:"runway_length"`
	Elev          *string `json:"elev"`
	Icao          string  `json:"icao"`
	DirectFlights string  `json:"direct_flights"`
	Carriers      string  `json:"carriers"`
}

type DataRetriever[T any] func(k string, v string) []T

const (
	SERVICE_NAME = "AIRPORT"
	VERSION      = "v1"
	PROJECT      = "TUTORIAL-SERVICES"
)

var (
	CONFIG *Config

	APP_READY chan struct{}

	Airports []Airport

	dictionary    = make(map[string]map[string][]Airport)
	allowedFields = make(map[string]bool)
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

	initDataAndDynRoutes(svc)
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

	file, err := os.OpenFile("airports.json", 0, fs.FileMode(os.O_RDONLY))

	if err != nil {
		log("[Init Svc].initSvc Error opening airports.json: ", err.Error())
		panic(err)
	}

	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&Airports)

	if err != nil {
		log("[Init Svc].initSvc Error decoding airports.json ", err.Error())
		panic(err)
	}
}

func initDataAndDynRoutes(svc *sb.Service) {
	defer log("Init Data... Done")
	log("Init Data...")

	allowedFields["State"] = true
	allowedFields["City"] = true
	allowedFields["Country"] = true

	//pivot data into new for from allowed fields
	//buildAirportIndex(allowedFields, dictionary)
	dictionary = u.BuildIndexedDataFromStructByFilter[Airport](allowedFields, &Airports)

	for f := range dictionary {
		log(" - Field Len: ", strconv.Itoa(len(dictionary[f])))
	}

	//In the future i could pull reflection into here but i think the data is static and small so just return Airports for now
	u.DynamicRESTFromTypeStruct[Airport](allowedFields, svc.GinEngine, func(fieldname string, value string) []Airport {
		return dictionary[fieldname][value]
	})
}

// User logic
func startUsrLogic(svc *sb.Service) {
	svc.GinEngine.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, os.Args)
	})

	svc.GinEngine.GET("/Airports", func(ctx *gin.Context) {

		if len(Airports) > 0 {
			ctx.JSON(http.StatusOK, Airports)
		} else {
			ctx.JSON(http.StatusNotFound, Airports)
		}

	})

	svc.GinEngine.GET("/Airports/:id", func(ctx *gin.Context) {

		id := ctx.Param("id")
		var result *Airport

		for i := range Airports {
			if Airports[i].Code == id {
				result = &Airports[i]
			}
		}

		if result != nil {
			ctx.JSON(http.StatusOK, result)
		} else {
			ctx.JSON(http.StatusNotFound, result)
		}

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
