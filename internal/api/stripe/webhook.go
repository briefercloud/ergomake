package stripe

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"

	"github.com/ergomake/ergomake/internal/logger"
)

func (stp *stripeRouter) webhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logger.Ctx(c).Err(err).Msg("error reading request body")
		c.JSON(http.StatusServiceUnavailable, http.StatusText(http.StatusServiceUnavailable))
		return
	}

	event, err := webhook.ConstructEvent(body, c.GetHeader("Stripe-Signature"), stp.webhookSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	log := logger.With(logger.Ctx(c)).Str("type", event.Type).Logger()

	if event.Type != "checkout.session.completed" {
		log.Info().Msg("stripe webhook ignored")
		c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
		return
	}

	var checkoutSession stripe.CheckoutSession
	err = json.Unmarshal(event.Data.Raw, &checkoutSession)
	if err != nil {
		log.Err(err).Msg("fail to unmarshall checkout session")
		c.JSON(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	if checkoutSession.Mode != stripe.CheckoutSessionModeSubscription {
		log.Info().Str("mode", string(checkoutSession.Mode)).Msg("got a unexpected checkout session mode")
		c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
		return
	}

	err = stp.paymentProvider.SaveSubscription(
		c,
		checkoutSession.ClientReferenceID,
		checkoutSession.Subscription.ID,
	)

	if err != nil {
		log.Err(err).Msg("fail to save subscription to database")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
}
