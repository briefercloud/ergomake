package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/ergomake/ergomake/internal/api"
	"github.com/ergomake/ergomake/internal/env"
	"github.com/ergomake/ergomake/internal/github/ghapp"
)

func main() {
	var cfg api.Config
	err := env.LoadEnv(&cfg)
	if err != nil {
		panic(errors.Wrap(err, "fail to load environment variables"))
	}

	ghApp, err := ghapp.NewGithubClient(cfg.GithubPrivateKey, cfg.GithubAppID)
	if err != nil {
		panic(errors.Wrap(err, "fail to create GitHub client"))
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "installations":
		owners, err := ghApp.ListInstalledOwners(context.Background())
		if err != nil {
			panic(errors.Wrap(err, "fail to list installed owners"))
		}

		for _, owner := range owners {
			fmt.Println(owner)
		}
	}
}

func printUsage() {
	fmt.Printf("Usage: %s <command> [arguments]\n", os.Args[0])
	fmt.Println("Commands:")
	fmt.Println("  installations   List current ergomake installations")
}
