package auth

import (
	"encoding/base64"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ergomake/ergomake/internal/github/ghapp"
	"github.com/ergomake/ergomake/internal/logger"
)

const AuthTokenCookieName = "auth_token"

type AuthData struct {
	jwt.StandardClaims
	GithubToken *oauth2.Token `json:"githubToken"`
}

func (ar *authRouter) callback(c *gin.Context) {
	log := logger.Ctx(c)

	bytesRedirectURL, err := base64.URLEncoding.DecodeString(c.Query("state"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid state")
		return
	}

	strRedirectURL := string(bytesRedirectURL)
	if strRedirectURL == "" {
		strRedirectURL = ar.frontendURL
	}

	installationID, err := strconv.ParseInt(c.Query("installation_id"), 10, 0)
	if err == nil {
		inst, err := ar.ghApp.GetInstallation(c, installationID)
		if err != nil && !errors.Is(err, ghapp.InstallationNotFoundError) {
			logger.Ctx(c).Err(err).Msg("fail to get installation")
			c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}

		strRedirectURL = fmt.Sprintf("%s/gh/%s", ar.frontendURL, inst.GetAccount().GetLogin())
	}

	redirectURL, err := url.Parse(strRedirectURL)
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
