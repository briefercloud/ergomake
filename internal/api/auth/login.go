package auth

import (
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func (ar *authRouter) login(c *gin.Context) {
	redirectURL, err := url.Parse(c.Query("redirectUrl"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid redirectUrl")
		return
	}

	state := base64.URLEncoding.EncodeToString([]byte(redirectURL.String()))
	authURL := ar.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}
