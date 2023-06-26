package github

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/api/auth"
	"github.com/ergomake/ergomake/internal/github/ghoauth"
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/payment"
)

func (ghr *githubRouter) listUserOrganizations(c *gin.Context) {
	authData, ok := auth.GetAuthData(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	client := ghoauth.FromToken(authData.GithubToken)

	user, res, err := client.GetUser(c)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusUnauthorized {
			c.JSON(
				http.StatusUnauthorized,
				http.StatusText(http.StatusInternalServerError),
			)
			return
		}

		logger.Ctx(c).Err(err).
			Msg("fail to get authenticated user")
		c.JSON(
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
		)
		return
	}

	result := []gin.H{}
	isUserInstalled, err := ghr.ghApp.IsOwnerInstalled(c, user.GetLogin())
	if err != nil {
		logger.Ctx(c).Err(err).
			Msg("fail to get check if user is installed")
		c.JSON(
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
		)
		return
	}

	if isUserInstalled {
		paymentPlan, err := ghr.paymentProvider.GetOwnerPlan(c, user.GetLogin())
		if err != nil {
			logger.Ctx(c).Err(err).Str("user", user.GetLogin()).
				Msg("fail to get user payment plan")
			c.JSON(
				http.StatusInternalServerError,
				http.StatusText(http.StatusInternalServerError),
			)
			return
		}

		result = append(result, gin.H{
			"login":    user.GetLogin(),
			"avatar":   user.GetAvatarURL(),
			"isPaying": paymentPlan != payment.PaymentPlanFree,
		})
	}

	orgs, res, err := client.ListOrganizations(c)
	if err != nil {
		if res.StatusCode == http.StatusUnauthorized {
			c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			return
		}

		logger.Ctx(c).Err(err).
			Msg("fail to list organizations of authenticated user")
		c.JSON(
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
		)
		return
	}

	for _, org := range orgs {
		paymentPlan, err := ghr.paymentProvider.GetOwnerPlan(c, org.GetLogin())
		if err != nil {
			logger.Ctx(c).Err(err).Str("owner", org.GetLogin()).
				Msg("fail to get owner payment plan")
			c.JSON(
				http.StatusInternalServerError,
				http.StatusText(http.StatusInternalServerError),
			)
			return
		}

		owner := gin.H{
			"login":    org.GetLogin(),
			"avatar":   org.GetAvatarURL(),
			"isPaying": paymentPlan != payment.PaymentPlanFree,
		}
		result = append(result, owner)
	}

	c.JSON(http.StatusOK, result)
}
