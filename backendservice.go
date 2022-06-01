package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"github.com/apache/trafficcontrol/lib/go-log"
	"github.com/apache/trafficcontrol/lib/go-tc"
	"github.com/apache/trafficcontrol/traffic_ops/traffic_ops_golang/api"
	"github.com/jmoiron/sqlx"
	"github.com/srijeet0406/backendservice/config"

	"net/http"
	"os"
	"time"
)

const readQuery = `SELECT * FROM foo`

type Foo struct {
	Name    string    `json:"name" db:"name"`
	ID         int    `json:"id" db:"id"`
	LastUpdated time.Time `json:"lastUpdated" db:"last_updated"`
}
func main() {
	dbConfigFileName := flag.String("dbcfg", "", "The database config path")
	configFileName := flag.String("cfg", "", "The config file path")
	flag.Parse()
	if configFileName == nil {
		fmt.Errorf("no config file found")
		os.Exit(1)
	}
	fmt.Println(*configFileName)
	fmt.Println(*dbConfigFileName)
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

	http.HandleFunc( "/api/4.0/foo", Getfoo)
	http.HandleFunc( "/api/4.0/foos", func( res http.ResponseWriter, req *http.Request ) {
		fmt.Println("in here ", req.URL.Path)
		res.Write([]byte("returning from the foos endpoint here\n"))
		return
	} )

	http.HandleFunc( "/", func( res http.ResponseWriter, req *http.Request ) {
		alerts := tc.CreateAlerts(tc.ErrorLevel, fmt.Sprintf("The requested path '%s' does not exist.", req.URL.Path))
		api.WriteAlerts(res, req, http.StatusNotFound, alerts)

	} )
	if err := server.ListenAndServeTLS(cfg.CertPath, cfg.KeyPath); err != nil {
		log.Errorf("stopping server: %v\n", err)
		os.Exit(1)
	}
}

func Getfoo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside foo here")
	db, err := sqlx.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&fallback_application_name=trafficops", "traffic_ops", "twelve", "localhost", "5432", "to_soa_development", "disable"))
	if err != nil {
		log.Errorf("opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	fooList := []Foo{}
	query := readQuery
	rows, err := db.NamedQuery(query, map[string]interface{}{})
	if err != nil {
		api.HandleErr(w, r, nil, http.StatusInternalServerError, nil, errors.New("querying foo: "+err.Error()))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var foo Foo
		if err = rows.Scan(&foo.ID, &foo.Name, &foo.LastUpdated); err != nil {
			api.HandleErr(w, r, nil, http.StatusInternalServerError, nil, errors.New("scanning cdn locks: "+err.Error()))
			return
		}
		fooList = append(fooList, foo)
	}
	api.WriteResp(w, r, fooList)
}