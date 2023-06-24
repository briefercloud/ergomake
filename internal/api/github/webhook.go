package github

import (
	"bytes"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v52/github"

	"github.com/ergomake/ergomake/internal/ginutils"
	"github.com/ergomake/ergomake/internal/logger"
)

var ownersBlockList = map[string]struct{}{"RahmaNiftaliyev": {}}

func (r *githubRouter) webhook(c *gin.Context) {
	log := logger.Ctx(c.Request.Context())

	var bodyBytes []byte
	err := c.ShouldBindBodyWith(&bodyBytes, ginutils.BYTES)
	if err != nil {
		log.Err(err).Msg("fail to read body")
		c.JSON(
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
		)
		return
	}

	payload, err := github.ValidatePayloadFromBody(
		c.ContentType(),
		bytes.NewReader(bodyBytes),
		c.GetHeader("X-Hub-Signature-256"),
		[]byte(r.webhookSecret),
	)
	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			http.StatusText(http.StatusUnauthorized),
		)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(c.Request), payload)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			http.StatusText(http.StatusBadRequest),
		)
		return
	}

	c.Status(http.StatusNoContent)

	githubDelivery := c.GetHeader("X-GitHub-Delivery")

	go func() {
		switch event := event.(type) {
		case *github.PullRequestEvent:
			action := event.GetAction()

			owner := event.GetRepo().GetOwner().GetLogin()
			repo := event.GetRepo().GetName()
			branch := event.GetPullRequest().GetHead().GetRef()
			sha := event.GetPullRequest().GetHead().GetSHA()
			prNumber := event.GetPullRequest().GetNumber()

			logCtx := logger.With(log).
				Str("githubDelivery", githubDelivery).
				Str("action", action).
				Str("owner", owner).
				Str("repo", repo).
				Int("prNumber", prNumber).
				Str("author", event.GetSender().GetLogin()).
				Str("branch", branch).
				Str("SHA", sha).
				Logger()
			log = &logCtx
			ctx := log.WithContext(context.Background())

			if _, blocked := ownersBlockList[owner]; blocked {
				log.Warn().Msg("event ignored because owner is in block list")
				return
			}

			log.Info().Msg("got a pull request event from github")
			switch action {
			case "opened", "reopened", "synchronize":
				err := r.terminateEnvironment(ctx, event)
				if err != nil {
					log.Err(err).Msg("fail to terminate environment")
				}
				err = r.launchEnvironment(ctx, event)
				if err != nil {
					log.Err(err).Msg("fail to launch environment")
				}
			case "closed":
				err := r.terminateEnvironment(ctx, event)
				if err != nil {
					log.Err(err).Msg("fail to terminate environment")
				}
			}
		}
	}()
}
