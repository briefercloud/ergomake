package auth

import (
	"encoding/base64"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"

	"github.com/ergomake/ergomake/internal/logger"
)

const AuthTokenCookieName = "auth_token"

type AuthData struct {
	jwt.StandardClaims
	GithubToken *oauth2.Token `json:"githubToken"`
}

func (ar *authRouter) callback(c *gin.Context) {
	log := logger.Ctx(c)

	rawRedirectURL, err := base64.URLEncoding.DecodeString(c.Query("state"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid state")
		return
	}

	redirectURL, err := url.Parse(string(rawRedirectURL))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid redirectUrl")
		return
	}

	queryParams := url.Values{}

	code := c.Query("code")
	githubToken, err := ar.oauthConfig.Exchange(c, code)
	if err != nil {
		queryParams.Set("error", "exchange")
		redirectURL.RawQuery = queryParams.Encode()
		c.Redirect(http.StatusTemporaryRedirect, redirectURL.String())
		return
	}

	expTime := time.Now().Add(time.Hour * 24)
	data := AuthData{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expTime.Unix(),
		},
		GithubToken: githubToken,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	tokenStr, err := token.SignedString([]byte(ar.jwtSecret))
	if err != nil {
		log.Err(err).Msg("fail to sign jwt token")
		queryParams.Set("error", "exchange")
		redirectURL.RawQuery = queryParams.Encode()
		c.Redirect(http.StatusTemporaryRedirect, redirectURL.String())
		return
	}

	httpsOnly := ar.secure
	maxAge := int(math.Floor(time.Until(expTime).Seconds()))
	c.SetCookie(AuthTokenCookieName, tokenStr, maxAge, "/", "", httpsOnly, true)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL.String())
}
