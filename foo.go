package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/apache/trafficcontrol/lib/go-log"
	//"github.com/apache/trafficcontrol/traffic_ops/traffic_ops_golang/api"
	//"github.com/apache/trafficcontrol/traffic_ops/traffic_ops_golang/auth"
	//"github.com/apache/trafficcontrol/traffic_ops/traffic_ops_golang/routing"
	"github.com/srijeet0406/backendservice/config"
	"net/http"
	"os"
	"time"
)

func foo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ohohohoooo")
	return
}
func main() {
	configFileName := flag.String("cfg", "", "The config file path")
	if configFileName == nil {
		fmt.Errorf("no config file found")
		os.Exit(1)
	}
	cfg, err := config.LoadConfig(*configFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Loading Config: %v\n", err)
		os.Exit(1)
	}

	if e := log.InitCfg(cfg); e != nil {
		fmt.Printf("Error initializing loggers: %v\n", e)
		fmt.Println(e)
		os.Exit(1)
	}
	if err != nil {
		log.Warnln(err)
	}
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		TLSConfig:         cfg.TLSConfig,
		ReadTimeout:       time.Duration(cfg.ReadTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(cfg.IdleTimeout) * time.Second,
		ErrorLog:          log.Error,
	}
	if server.TLSConfig == nil {
		server.TLSConfig = &tls.Config{}
	}
	server.TLSConfig.InsecureSkipVerify = cfg.Insecure


	//routing.Srijeet(0,
	//	routing.Route{
	//	Version:             api.Version{
	//		Major: 4,
	//		Minor: 0,
	//	},
	//	Method:              http.MethodGet,
	//	Path:                "foo",
	//	Handler:             foo,
	//	RequiredPrivLevel:   auth.PrivLevelReadOnly,
	//	RequiredPermissions: nil,
	//	Authenticated:       true,
	//	Middlewares:         nil,
	//	ID:                  123456789,
	//}, cfg)

	go func() {
		if cfg.KeyPath == "" {
			log.Errorf("key cannot be blank in %s", cfg.ConfigHypnotoad.Listen)
			os.Exit(1)
		}

		if cfg.CertPath == "" {
			log.Errorf("cert cannot be blank in %s", cfg.ConfigHypnotoad.Listen)
			os.Exit(1)
		}

		if file, err := os.Open(cfg.CertPath); err != nil {
			log.Errorf("cannot open %s for read: %s", cfg.CertPath, err.Error())
			os.Exit(1)
		} else {
			file.Close()
		}

		if file, err := os.Open(cfg.KeyPath); err != nil {
			log.Errorf("cannot open %s for read: %s", cfg.KeyPath, err.Error())
			os.Exit(1)
		} else {
			file.Close()
		}

		if err := server.ListenAndServeTLS(cfg.CertPath, cfg.KeyPath); err != nil {
			log.Errorf("stopping server: %v\n", err)
			os.Exit(1)
		}
	}()
}
