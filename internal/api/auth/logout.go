package auth

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func (ar *authRouter) logout(c *gin.Context) {
	redirectURL, err := url.Parse(c.Query("redirectUrl"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid redirectUrl")
		return
	}

	c.SetCookie(AuthTokenCookieName, "", -1, "/", "", false, true)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL.String())
}
