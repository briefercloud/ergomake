package github

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v52/github"

	"github.com/ergomake/ergomake/internal/ginutils"
	"github.com/ergomake/ergomake/internal/logger"
)

func (r *githubRouter) marketplaceWebhook(c *gin.Context) {
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

	switch event := event.(type) {
	case *github.MarketplacePurchaseEvent:
		action := event.GetAction()
		installationTarget := event.MarketplacePurchase.Account.GetLogin()

		logCtx := logger.With(log).
			Str("action", action).
			Str("installation_target", installationTarget).
			Str("installation_originator", event.Sender.GetLogin()).
			Logger()
		log = &logCtx

		log.Info().Msg("got a marketplace event")

		err := r.db.SaveEvent(installationTarget, action)
		if err != nil {
			log.Err(err).Msg("fail to save marketplace event in database")
		}
	}

	c.Status(http.StatusNoContent)
}
