package main

import (
	"log"

	"github.com/spf13/viper"
)

var (
	defaults = map[string]interface{}{
		"DEBUG":           true,
		"PORT":            80,
		"SSL_CERT":        "",
		"SSL_KEY":         "",
		"DB_HOST":         "host",
		"DB_PORT":         5432,
		"DB_NAME":         "database",
		"DB_USER":         "user",
		"DB_PASSWORD":     "password",
		"JWT_KEY":         "secret",
		"JWT_MAX_AGE":     1200,
		"REFRESH_MAX_AGE": 2592000,
		"HCAPTCHA_SECRET": "",
		"REGISTER":        true,
	}
	configPaths = []string{
		".",
	}
)

// Configuration struct
type Configuration struct {
	Debug          bool       `mapstructure:"DEBUG"`
	Port           int        `mapstructure:"PORT"`
	SSLCert        string     `mapstructure:"SSL_CERT"`
	SSLKey         string     `mapstructure:"SSL_KEY"`
	Db             DataSource `mapstructure:",squash"`
	JWTKey         string     `mapstructure:"JWT_KEY"`
	JWTMaxAge      int        `mapstructure:"JWT_MAX_AGE"`
	RefreshMaxAge  int        `mapstructure:"REFRESH_MAX_AGE"`
	HCaptchaSecret string     `mapstructure:"HCAPTCHA_SECRET"`
	Register       bool       `mapstructure:"REGISTER"`
}

// DataSource struct
type DataSource struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     int    `mapstructure:"DB_PORT"`
	Dbname   string `mapstructure:"DB_NAME"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
}

func ReadConfig(ENV string) (Configuration, error) {
	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
	viper.SetConfigName(ENV)
	for _, p := range configPaths {
		viper.AddConfigPath(p)
	}
	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err)
	}
	viper.AutomaticEnv()
	var config Configuration
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}
