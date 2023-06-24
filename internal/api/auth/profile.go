package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/github/ghoauth"
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/users"
)

func (ar *authRouter) profile(c *gin.Context) {
	authData, ok := GetAuthData(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	client := ghoauth.FromToken(authData.GithubToken)
	user, r, err := client.GetUser(c)
	if err != nil {
		if r.StatusCode == http.StatusUnauthorized {
			c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			return
		}

		logger.Ctx(c).Err(err).Msg("fail to get user from github")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	err = ar.usersService.Save(c, users.User{
		Email:    user.GetEmail(),
		Username: user.GetLogin(),
		Name:     user.GetName(),
		Provider: users.ProviderGithub,
	})
	if err != nil {
		logger.Ctx(c).Err(err).Msg("fail to save user")
	}

	c.JSON(http.StatusOK, gin.H{
		"avatar":   user.GetAvatarURL(),
		"username": user.GetLogin(),
	})
}
