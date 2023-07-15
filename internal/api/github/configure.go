package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/api/auth"
	"github.com/ergomake/ergomake/internal/logger"
)

func (ghr *githubRouter) configureRepo(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	authData, ok := auth.GetAuthData(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	isAuthorized, err := auth.IsAuthorized(c, owner, authData)
	if err != nil {
		logger.Ctx(c).Err(err).
			Str("owner", owner).
			Str("repo", repo).
			Msg("fail to check if user is authorized to configure repo")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if !isAuthorized {
		c.JSON(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	ergopack := `apps:
  app:
    path: ../
    publicPort: 3000
`
	changes := map[string]string{
		".ergomake/ergopack.yaml": ergopack,
	}
	title := "Introduce pull request previews"
	description := `# Summary

These changes include a template for setting up pull request previews.

After you adjust this configuration file, [Ergomake](https://ergomake.dev) will create a preview environment whenever developers create a pull-request. Once the preview environment is up, Ergomake will post a link to access it.


# How it works

The ` + "`docker-compose.yaml`" + ` file within ` + "`.ergomake`" + ` contains the configurations necessary to spin up an environment. Whenever this file exists in a pull-request, we'll use it to spin up a preview.

Please update this `+ "`docker-compose.yaml`" + `file by pushing more code to this branch (` + "`ergomake`" + `). Once it works fine, you should have a working preview link.

Here are the most common actions you may need to take:

1. Create a ` + "`Dockerfile`" + ` to build your application and add it to ` + "`docker-compose.yaml`" + `.
2. Add any databases or other services your application depends on to ` + "`docker-compose.yaml`" + `.
3. Add environment variables by logging into [the dashboard](https://app.ergomake.dev) and selecting this repository.

For more information, please see our [documentation](https://docs.ergomake.dev/).


## Tips for writing your compose file

- You can see the build logs for your services in the [dashboard](https://app.ergomake.dev).
- Make the first service your front-end application. This will be the service whose link comes first in our comment.
- Expose your applications by binding their desired ports to ` + "`localhost`" + `. To expose port 3000, for example, you can use ` + "`3000:3000`" + `.
- Avoid unnecessary complications, like using ` + "`depends_on`," + "`volumes`" + `, and ` + "`networks`" + `.
- To seed your database, you can use a ` + "`command`" + ` to run a script after the database is up. For example, you can use ` + "`command: bash -c \"sleep 5 && npm run seed\"" + ` to seed a database after 5 seconds. Make sure that your seed command doesn't cause the container to crash if it fails.


# Where to go from here

In our platform, you can configure branches to be permanently deployed. That way, you can access that branch at any time, regardless of whether there's a PR with its contents. **Permanent branches are useful for permanent staging, QA, or development environments.**

---

💻 [GitHub](https://github.com/ergomake/ergomake) | 🌐 [Discord](https://discord.gg/daGzchUGDt) | 🐦 [Twitter](https://twitter.com/GetErgomake)`

	pr, err := ghr.ghApp.CreatePullRequest(
		context.Background(),
		owner, repo, "ergomake",
		changes, title, description,
	)
	if err != nil {
		logger.Ctx(c).Err(err).
			Str("owner", owner).
			Str("repo", repo).
			Msg("fail to create pull request")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	url := fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repo, pr.GetNumber())
	c.JSON(http.StatusOK, gin.H{"pullRequestURL": url})
}
