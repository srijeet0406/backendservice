package config

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/apache/trafficcontrol/lib/go-log"
	"io/ioutil"
	"net/url"
)

type ConfigHypnotoad struct {
	Listen []string `json:"listen"`
}

// Config reflects the structure of the cdn.conf file
type Config struct {
	TOHost             string      `json:"traffic_ops_host"`
	TOPort             string      `json:"traffic_ops_port"`
	LogLocationError   string      `json:"log_location_error"`
	LogLocationWarning string      `json:"log_location_warning"`
	LogLocationInfo    string      `json:"log_location_info"`
	LogLocationDebug   string      `json:"log_location_debug"`
	LogLocationEvent   string      `json:"log_location_event"`
	TLSConfig          *tls.Config `json:"tls_config"`
	ReadTimeout        int         `json:"read_timeout"`
	RequestTimeout     int         `json:"request_timeout"`
	ReadHeaderTimeout  int         `json:"read_header_timeout"`
	WriteTimeout       int         `json:"write_timeout"`
	IdleTimeout        int         `json:"idle_timeout"`
	Insecure           bool        `json:"insecure"`
	Port               string      `json:"port"`
	CertPath           string      `json:"-"`
	KeyPath            string      `json:"-"`
	ConfigHypnotoad    `json:"hypnotoad"`
	URL                *url.URL `json:"-"`
}

type DBConfig struct {
	Description string `json:"description"`
	DBName      string `json:"dbname"`
	Hostname    string `json:"hostname"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Port        string `json:"port"`
	Type        string `json:"type"`
	SSL         bool   `json:"ssl"`
}
// ErrorLog - critical messages
func (c Config) ErrorLog() log.LogLocation {
	return log.LogLocation(c.LogLocationError)
}

// WarningLog - warning messages
func (c Config) WarningLog() log.LogLocation {
	return log.LogLocation(c.LogLocationWarning)
}

// InfoLog - information messages
func (c Config) InfoLog() log.LogLocation { return log.LogLocation(c.LogLocationInfo) }

// DebugLog - troubleshooting messages
func (c Config) DebugLog() log.LogLocation {
	return log.LogLocation(c.LogLocationDebug)
}

// EventLog - access.log high level transactions
func (c Config) EventLog() log.LogLocation {
	return log.LogLocation(c.LogLocationEvent)
}

// GetCertPath - extracts path to cert .cert file
func (c Config) GetCertPath() string {
	v, ok := c.URL.Query()["cert"]
	if ok {
		return v[0]
	}
	return ""
}

// GetKeyPath - extracts path to cert .key file
func (c Config) GetKeyPath() string {
	v, ok := c.URL.Query()["key"]
	if ok {
		return v[0]
	}
	return ""
}

func LoadConfig(confPath string) (Config, error) {
	// load json from conf file
	confBytes, err := ioutil.ReadFile(confPath)
	if err != nil {
		return Config{}, fmt.Errorf("reading CDN conf '%s': %v", confPath, err)
	}

	cfg := Config{}
	err = json.Unmarshal(confBytes, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("unmarshalling '%s': %v", confPath, err)
	}
	if len(cfg.Listen) < 1 {
		return Config{}, fmt.Errorf("no listen in config")
	} else {
		listen := cfg.Listen[0]
		if cfg.URL, err = url.Parse(listen); err != nil {
			return Config{}, fmt.Errorf("invalid Traffic Ops URL '%s': %v", listen, err)
		}
		cfg.KeyPath = cfg.GetKeyPath()
		cfg.CertPath = cfg.GetCertPath()

		newURL := url.URL{Scheme: cfg.URL.Scheme, Host: cfg.URL.Host, Path: cfg.URL.Path}
		cfg.URL = &newURL
	}
	return cfg, nil
}