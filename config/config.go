package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/sebastianrakel/openvoxview/model"
	"github.com/spf13/viper"
)

var configPath = flag.String("config", "config.yml", "path to the config file ")
var printVersion = flag.Bool("version", false, "prints version")

func init() {
	flag.Parse()
}

type ConfigPqlQuery struct {
	Description string `mapstructure:"description"`
	Query       string `mapstructure:"query"`
}

type Config struct {
	Listen         string   `mapstructure:"listen"`
	Port           uint64   `mapstructure:"port"`
	TrustedProxies []string `mapstructure:"trusted_proxies"`
	PuppetDB       struct {
		Host      string `mapstructure:"host"`
		Port      uint64 `mapstructure:"port"`
		TLS       bool   `mapstructure:"tls"`
		TLSIgnore bool   `mapstructure:"tls_ignore"`
		TLS_CA    string `mapstructure:"tls_ca"`
		TLS_KEY   string `mapstructure:"tls_key"`
		TLS_CERT  string `mapstructure:"tls_cert"`
	} `mapstructure:"puppetdb"`
	PqlQueries []ConfigPqlQuery `mapstructure:"queries"`
	Views      []model.View     `mapstructure:"views"`
}

func PrintVersion(version string) bool {
	if *printVersion {
		fmt.Println(version)
		return true
	}
	return false
}

func GetConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if configPath != nil {
		log.Printf("Using config: %s", *configPath)
		viper.SetConfigFile(*configPath)
	}

	viper.SetDefault("port", 5000)
	viper.SetDefault("puppetdb.host", "localhost")
	viper.SetDefault("puppetdb.port", 8080)
	viper.SetDefault("puppetdb.tls_ignore", false)

	viper.AutomaticEnv()

	viper.BindEnv("port", "PORT")
	viper.BindEnv("listen", "LISTEN")
	viper.BindEnv("trusted_proxies", "TRUSTED_PROXIES")
	viper.BindEnv("puppetdb.port", "PUPPETDB_PORT")
	viper.BindEnv("puppetdb.host", "PUPPETDB_HOST")
	viper.BindEnv("puppetdb.tls", "PUPPETDB_TLS")
	viper.BindEnv("puppetdb.tls_ignore", "PUPPETDB_TLS_IGNORE")
	viper.BindEnv("puppetdb.tls_ca", "PUPPETDB_TLS_CA")
	viper.BindEnv("puppetdb.tls_key", "PUPPETDB_TLS_KEY")
	viper.BindEnv("puppetdb.tls_cert", "PUPPETDB_TLS_CERT")

	viper.ReadInConfig()

	var cfg Config

	err := viper.Unmarshal(&cfg)

	cfg.TrustedProxies = viper.GetStringSlice("trusted_proxies")

	return &cfg, err
}

func (c *Config) GetPuppetDbAddress() string {
	scheme := "http"
	if c.PuppetDB.TLS {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s:%d", scheme, c.PuppetDB.Host, c.PuppetDB.Port)
}
