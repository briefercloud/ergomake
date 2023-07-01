package github

import (
	"bytes"
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
		case *github.PushEvent:
			r.handlePushEvent(githubDelivery, event)
		case *github.PullRequestEvent:
			r.handlePullRequestEvent(githubDelivery, event)
		}
	}()
}
