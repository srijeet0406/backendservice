package config

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/apache/trafficcontrol/lib/go-log"
	"io/ioutil"
)

type ConfigHypnotoad struct {
	Listen []string `json:"listen"`
}

// Config reflects the structure of the cdn.conf file
type Config struct {
	TOHost                     string            `json:"traffic_ops_host"`
	TOPort                     string            `json:"traffic_ops_port"`
	LogLocationError         string                     `json:"log_location_error"`
	LogLocationWarning       string                     `json:"log_location_warning"`
	LogLocationInfo          string                     `json:"log_location_info"`
	LogLocationDebug         string                     `json:"log_location_debug"`
	LogLocationEvent         string                     `json:"log_location_event"`
	TLSConfig         *tls.Config     `json:"tls_config"`
	ReadTimeout       int                        `json:"read_timeout"`
	RequestTimeout    int                        `json:"request_timeout"`
	ReadHeaderTimeout int                        `json:"read_header_timeout"`
	WriteTimeout      int                        `json:"write_timeout"`
	IdleTimeout       int                        `json:"idle_timeout"`
	Insecure          bool                       `json:"insecure"`
	Port              string                     `json:"port"`
	CertPath          string   `json:"-"`
	KeyPath           string   `json:"-"`
	ConfigHypnotoad   `json:"hypnotoad"`
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
	return cfg, nil
}