package env

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

func LoadEnv(e interface{}) error {
	err := godotenv.Load(".env.local")
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "fail to load .env.local file")
		}
	}

	return errors.Wrap(envconfig.Process("", e), "fail to extract env variables")
}
