package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"k8s.io/utils/pointer"

	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/env"
	"github.com/ergomake/ergomake/internal/envvars"
)

type config struct {
	DatabaseURL   string `split_words:"true"`
	EnvVarsSecret string `split_words:"true"`
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

	envVarProvider := envvars.NewDBEnvVarProvider(db, cfg.EnvVarsSecret)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "upsert":
		if len(args) != 4 {
			fmt.Println("Invalid number of arguments for 'upsert' command")
			printUsage()
			os.Exit(1)
		}
		owner := args[0]
		repo := args[1]
		name := args[2]
		value := args[3]
		branch := new(string)
		if len(args) >= 4 {
			branch = pointer.String(args[4])
		}
		err := envVarProvider.Upsert(context.Background(), owner, repo, name, value, branch)
		if err != nil {
			panic(errors.Wrap(err, "fail to upsert environment variable"))
		}
		fmt.Println("Environment variable upserted successfully")
	case "delete":
		if len(args) != 3 {
			fmt.Println("Invalid number of arguments for 'delete' command")
			printUsage()
			os.Exit(1)
		}
		owner := args[0]
		repo := args[1]
		name := args[2]
		err := envVarProvider.Delete(context.Background(), owner, repo, name)
		if err != nil {
			panic(errors.Wrap(err, "fail to delete environment variable"))
		}
		fmt.Println("Environment variable deleted successfully")
	case "list":
		if len(args) != 2 {
			fmt.Println("Invalid number of arguments for 'list' command")
			printUsage()
			os.Exit(1)
		}
		owner := args[0]
		repo := args[1]
		vars, err := envVarProvider.ListByRepo(context.Background(), owner, repo)
		if err != nil {
			panic(errors.Wrap(err, "fail to list environment variables"))
		}
		fmt.Println("Environment variables:")
		for _, v := range vars {
			fmt.Printf("%s: %s\n", v.Name, v.Value)
		}
	default:
		fmt.Println("Invalid command")
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("Usage: %s <command> [arguments]\n", os.Args[0])
	fmt.Println("Commands:")
	fmt.Println("  upsert <owner> <repo> <name> <value>   Upsert an environment variable")
	fmt.Println("  delete <owner> <repo> <name>           Delete an environment variable")
	fmt.Println("  list <owner> <repo>                    List environment variables by repository")
}
