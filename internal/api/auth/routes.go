package auth

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"github.com/ergomake/ergomake/internal/github/ghapp"
	"github.com/ergomake/ergomake/internal/users"
)

type authRouter struct {
	oauthConfig  *oauth2.Config
	jwtSecret    string
	secure       bool
	usersService users.Service
	frontendURL  string
	ghApp        ghapp.GHAppClient
}

func NewAuthRouter(
	clientID string,
	clientSecret string,
	redirectURL string,
	jwtSecret string,
	secure bool,
	usersService users.Service,
	frontendURL string,
	ghapp ghapp.GHAppClient,
) *authRouter {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}

	return &authRouter{oauthConfig, jwtSecret, secure, usersService, frontendURL, ghapp}
}

func (ar *authRouter) AddRoutes(router *gin.RouterGroup) {
	router.GET("/login", ar.login)
	router.GET("/logout", ar.logout)
	router.GET("/callback", ar.callback)
	router.GET("/profile", ar.profile)
}
