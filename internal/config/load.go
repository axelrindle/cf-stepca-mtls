package config

import (
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
)

func Load() *Config {
	c := &Config{}

	err := env.ParseWithOptions(c, env.Options{
		DefaultValueTagName: "default",
		PrefixTagName:       "prefix",
		RequiredIfNoDef:     true,
	})
	if err != nil {
		if aggErr, ok := err.(env.AggregateError); ok {
			for _, err := range aggErr.Errors {
				log.Println(err)
			}
			os.Exit(1)
		} else {
			log.Fatal(err)
		}
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(c); err != nil {
		log.Fatal(err)
	}

	return c
}
