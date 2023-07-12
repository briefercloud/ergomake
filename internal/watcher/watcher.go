package watcher

import (
	"context"
	"errors"
	"time"

	"k8s.io/utils/pointer"

	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/environments"
	"github.com/ergomake/ergomake/internal/github/ghapp"
	"github.com/ergomake/ergomake/internal/github/ghlauncher"
	"github.com/ergomake/ergomake/internal/logger"
)

func WatchEnvironments(
	ctx context.Context,
	db *database.DB,
	environmentsProvider environments.EnvironmentsProvider,
	ghApp ghapp.GHAppClient,
	ghLauncher ghlauncher.GHLauncher,
) func() {
	stopCh := make(chan struct{})
	go func() {
		timer := time.NewTimer(5 * time.Second)
		for {
			timer.Reset(time.Second * 5)

			logger.Ctx(ctx).Info().Msg("checking limited environments")
			var envs []database.Environment
			err := db.Table("environments").Find(&envs, map[string]interface{}{"status": database.EnvLimited}).Error
			if err != nil {
				logger.Get().Err(err).Msg("fail to list limited environments")
			}

			visitedOwners := make(map[string]bool)

			logger.Ctx(ctx).Info().Msgf("got %d limited environments", len(envs))
			for _, env := range envs {
				log := logger.With(logger.Ctx(ctx)).Interface("environment", env).Logger()
				if visitedOwners[env.Owner] {
					continue
				}

				isLimited, err := environmentsProvider.IsOwnerLimited(ctx, env.Owner)
				if err != nil {
					log.Err(err).Msg("fail to check if owner is limited")
					continue
				}

				visitedOwners[env.Owner] = true
				if isLimited {
					continue
				}

				log.Info().Msg("owner is not limited relaunching environment")

				var pr *int
				if env.PullRequest.Valid {
					pr = pointer.Int(int(env.PullRequest.Int32))
				}

				terminateReq := environments.TerminateEnvironmentRequest{
					Owner:    env.Owner,
					Repo:     env.Repo,
					Branch:   env.Branch.String,
					PrNumber: pr,
				}

				sha := ""
				if env.Branch.Valid {
					s, err := ghApp.GetBranchSHA(ctx, env.BranchOwner, env.Repo, env.Branch.String)
					if err != nil {
						if errors.Is(err, ghapp.BranchNotFoundError) {
							log.Warn().Msg("got BranchNotFoundError when trying to relaunch limited env, terminating env")
							err := environmentsProvider.TerminateEnvironment(ctx, terminateReq)
							if err != nil {
								log.Err(err).Msg("fail to terminate limited environment after branch not found")
							}
							continue
						}

						log.Err(err).Msg("fail to get branch sha")
						continue
					}
					sha = s
				}

				isPrivate, err := ghApp.IsRepoPrivate(ctx, env.BranchOwner, env.Repo)
				if err != nil {
					if errors.Is(err, ghapp.RepoNotFoundError) {
						log.Warn().Msg("got RepoNotFoundError when trying to relaunch limited env")
						err := environmentsProvider.TerminateEnvironment(ctx, terminateReq)
						if err != nil {
							log.Err(err).Msg("fail to terminate limited environment after repo not found")
						}
						continue
					}
				}

				err = environmentsProvider.TerminateEnvironment(ctx, terminateReq)
				if err != nil {
					log.Err(err).Msg("fail to terminate limited environment for relaunch")
				}

				launchReq := ghlauncher.LaunchEnvironmentRequest{
					Owner:       env.Owner,
					BranchOwner: env.BranchOwner,
					Repo:        env.Repo,
					Branch:      env.Branch.String,
					SHA:         sha,
					PrNumber:    pr,
					Author:      env.Author,
					IsPrivate:   isPrivate,
				}
				go func() {
					err := ghLauncher.LaunchEnvironment(context.Background(), launchReq)
					if err != nil {
						log.Err(err).Interface("launch", launchReq).Msg("fail to launch environment")
					}
				}()
			}

			select {
			case <-timer.C:
				continue
			case <-stopCh:
				return
			}
		}
	}()

	return func() {
		close(stopCh)
	}
}
