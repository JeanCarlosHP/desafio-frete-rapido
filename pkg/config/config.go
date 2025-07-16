package config

import (
	"reflect"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort           string `mapstructure:"APP_PORT"`
	LoggingJSONFormat bool   `mapstructure:"LOGGING_JSON_FORMAT"`
	LoggingLevel      string `mapstructure:"LOGGING_LEVEL"`

	DatabaseHost     string `mapstructure:"DATABASE_HOST"`
	DatabasePort     string `mapstructure:"DATABASE_PORT"`
	DatabaseUser     string `mapstructure:"DATABASE_USER"`
	DatabasePassword string `mapstructure:"DATABASE_PASSWORD"`
	DatabaseName     string `mapstructure:"DATABASE_NAME"`

	HTTPCorsAllowedHeaders string `mapstructure:"HTTP_CORS_ALLOWED_HEADERS"`
	HTTPCorsAllowedMethods string `mapstructure:"HTTP_CORS_ALLOWED_METHODS"`
	HTTPCorsAllowedOrigins string `mapstructure:"HTTP_CORS_ALLOWED_ORIGINS"`
}

func New() *Config {
	config := &Config{}
	err := load(config)
	if err != nil {
		panic(err)
	}

	return config
}

func load(config *Config) error {
	viper.AddConfigPath(".")
	viper.SetConfigType("env")
	viper.SetConfigName(".env.local")

	viper.AutomaticEnv()

	variableNames := getTags("mapstructure", Config{})

	for _, v := range variableNames {
		if err := viper.BindEnv(v); err != nil {
			return err
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		return err
	}

	return nil
}

func getTags(tagName string, obj any) []string {
	var tags []string
	envVarType := reflect.TypeOf(obj)

	for i := 0; i < envVarType.NumField(); i++ {
		field := envVarType.Field(i)
		tag := field.Tag.Get(tagName)
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags
}
