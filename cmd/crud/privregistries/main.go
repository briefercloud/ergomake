package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/env"
	"github.com/ergomake/ergomake/internal/privregistry"
)

type config struct {
	DatabaseURL          string `split_words:"true"`
	PrivRegistriesSecret string `split_words:"true"`
}

func main() {
	var cfg config
	err := env.LoadEnv(&cfg)
	if err != nil {
		panic(errors.Wrap(err, "fail to load environment variables"))
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		panic(errors.Wrap(err, "fail to connect to the database"))
	}
	defer db.Close()

	privRegistryProvider := privregistry.NewDBPrivRegistryProvider(db, cfg.PrivRegistriesSecret)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "fetch-creds":
		if len(args) != 2 {
			fmt.Println("Invalid number of arguments for 'fetch-creds' command")
			printUsage()
			os.Exit(1)
		}
		owner := args[0]
		image := args[1]
		creds, err := privRegistryProvider.FetchCreds(context.Background(), owner, image)
		if err != nil {
			if errors.Is(err, privregistry.ErrRegistryNotFound) {
				fmt.Println("Private registry not found")
				os.Exit(1)
			}
			panic(errors.Wrap(err, "fail to fetch credentials from private registry"))
		}
		fmt.Printf("Registry URL: %s\nToken: %s\n", creds.URL, creds.Token)
	case "store-registry":
		if len(args) != 4 {
			fmt.Println("Invalid number of arguments for 'store-registry' command")
			printUsage()
			os.Exit(1)
		}
		owner := args[0]
		url := args[1]
		provider := args[2]
		credentials := args[3]
		err := privRegistryProvider.StoreRegistry(context.Background(), owner, url, provider, credentials)
		if err != nil {
			panic(errors.Wrap(err, "fail to store private registry"))
		}
		fmt.Println("Private registry stored successfully")
	default:
		fmt.Println("Invalid command")
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("Usage: %s <command> [arguments]\n", os.Args[0])
	fmt.Println("Commands:")
	fmt.Println("  fetch-creds <owner> <image>         Fetch credentials from a private registry")
	fmt.Println("  store-registry <owner> <url> <provider> <credentials>   Store a private registry")
}
