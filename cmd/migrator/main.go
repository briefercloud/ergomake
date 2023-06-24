package main

import (
	"log"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/env"
)

type cfg struct {
	DatabaseUrl string `split_words:"true"`
}

func main() {
	var cfg cfg
	err := env.LoadEnv(&cfg)
	if err != nil {
		log.Panic(errors.Wrap(err, "fail to load environment variables"))
	}

	db, err := database.Connect(cfg.DatabaseUrl)
	if err != nil {
		log.Panic(errors.Wrap(err, "fail to connect to database"))
	}

	n, err := db.Migrate()
	if err != nil {
		log.Panic(errors.Wrap(err, "fail to execute db migrations"))
	}

	log.Println("Applied", n, "migrations")
}
