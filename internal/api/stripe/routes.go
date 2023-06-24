package stripe

import (
	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/payment"
)

type stripeRouter struct {
	paymentProvider payment.PaymentProvider
	webhookSecret   string
}

func NewStripeRouter(
	paymentProvider payment.PaymentProvider,
	webhookSecret string,
) *stripeRouter {
	return &stripeRouter{paymentProvider, webhookSecret}
}

func (stp *stripeRouter) AddRoutes(router *gin.RouterGroup) {
	router.POST("/webhook", stp.webhook)
}
