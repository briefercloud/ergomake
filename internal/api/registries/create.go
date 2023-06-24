package registries

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/api/auth"
	"github.com/ergomake/ergomake/internal/logger"
)

type createRegistry struct {
	URL         string `json:"url"`
	Provider    string `json:"provider"`
	Credentials string `json:"credentials"`
}

func (rr *registriesRouter) create(c *gin.Context) {
	authData, ok := auth.GetAuthData(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	owner := c.Param("owner")
	if owner == "" {
		c.JSON(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	var body createRegistry
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "malformed-payload"})
		return
	}

	isAuthorized, err := auth.IsAuthorized(c, owner, authData)
	if err != nil {
		logger.Ctx(c).Err(err).Msg("fail to create registry")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if !isAuthorized {
		c.JSON(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	err = rr.privRegistryProvider.StoreRegistry(c, owner, body.URL, body.Provider, body.Credentials)
	if err != nil {
		logger.Ctx(c).Err(err).Msg("fail to create registry")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusCreated, nil)
}
