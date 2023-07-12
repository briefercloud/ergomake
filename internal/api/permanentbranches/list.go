package permanentbranches

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/api/auth"
	"github.com/ergomake/ergomake/internal/logger"
)

func (pbr *permanentBranchesRouter) list(c *gin.Context) {
	authData, ok := auth.GetAuthData(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	isAuthorized, err := auth.IsAuthorized(c, owner, authData)
	if err != nil {
		logger.Ctx(c).Err(err).Msg("fail to list variables")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if !isAuthorized {
		c.JSON(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	branches, err := pbr.permanentbranchesProvider.List(c, owner, repo)
	if err != nil {
		logger.Ctx(c).Err(err).Msgf("fail to list permanent branches for repo %s/%s", owner, repo)
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	output := make([]gin.H, 0)
	for _, b := range branches {
		output = append(output, gin.H{"name": b})
	}
	c.JSON(http.StatusOK, output)
}
