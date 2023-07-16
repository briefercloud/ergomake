package deploy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ergomake/ergomake/internal/aws"
	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/environments"
	"github.com/ergomake/ergomake/internal/envvars"
	"github.com/ergomake/ergomake/internal/github/ghapp"
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/privregistry"
	"github.com/ergomake/ergomake/internal/transformer/builder"
	"github.com/ergomake/ergomake/internal/transformer/tarfile"
)

type deployRouter struct {
	db                      *database.DB
	ghApp                   ghapp.GHAppClient
	clusterClient           cluster.Client
	envVarsProvider         envvars.EnvVarsProvider
	privRegistryProvider    privregistry.PrivRegistryProvider
	environmentsProvider    environments.EnvironmentsProvider
	s3Bucket                string
	awsAccessKey            string
	awsSecretAccessKey      string
	dockerhubPullSecretName string
}

func NewDeployRouter(
	db *database.DB,
	ghApp ghapp.GHAppClient,
	clusterClient cluster.Client,
	envVarsProvider envvars.EnvVarsProvider,
	privRegistryProvider privregistry.PrivRegistryProvider,
	environmentsProvider environments.EnvironmentsProvider,
	s3Bucket string,
	awsAccessKey string,
	awsSecretAccessKey string,
	dockerhubPullSecretName string,
) *deployRouter {
	return &deployRouter{
		db,
		ghApp,
		clusterClient,
		envVarsProvider,
		privRegistryProvider,
		environmentsProvider,
		s3Bucket,
		awsAccessKey,
		awsSecretAccessKey,
		dockerhubPullSecretName,
	}
}

func (dr *deployRouter) AddRoutes(router *gin.RouterGroup) {
	router.POST("/deploy", dr.handleUpload)
}

func (dr *deployRouter) handleUpload(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	defer file.Close()

	id := uuid.New()

	tempDir, err := os.MkdirTemp("", fmt.Sprintf("upload-%s", id.String()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	tempFilePath := filepath.Join(tempDir, "archive.tar.gz")
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	defer tempFile.Close()

	// Copy the file content to the destination file
	if _, err := io.Copy(tempFile, file); err != nil {
		logger.Ctx(c).Err(err).Stack().Str("id", id.String()).Msg("fail to store uploaded file")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))

		c.JSON(http.StatusInternalServerError, http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	c.SSEvent("progress", gin.H{"status": "pending"})
	c.Writer.Flush()

	err = aws.UploadFileToS3(
		dr.s3Bucket,
		tempFilePath,
		fmt.Sprintf("%s/archive.tar.gz", id.String()),
	)
	if err != nil {
		logger.Ctx(c).Err(err).Stack().Str("id", id.String()).Msg("fail to upload to s3")
		c.SSEvent(
			"finish",
			gin.H{"status": "error"},
		)
		c.Writer.Flush()
		return
	}

	defer os.RemoveAll(tempDir)

	ctx := context.Background()
	tt := tarfile.NewTarfileTransformer(
		dr.clusterClient,
		dr.ghApp,
		dr.db,
		dr.envVarsProvider,
		dr.privRegistryProvider,
		dr.s3Bucket,
		dr.awsAccessKey,
		dr.awsSecretAccessKey,
		tempFilePath,
		&builder.GitOptions{},
		dr.dockerhubPullSecretName,
	)

	prepare, err := tt.Prepare(ctx, id)
	if err != nil {
		logger.Ctx(c).Err(err).Stack().Str("id", id.String()).Msg("fail to prepare preview")
		c.SSEvent(
			"finish",
			gin.H{"status": "error"},
		)
		c.Writer.Flush()
		return
	}

	if prepare.ValidationError != nil {
		c.SSEvent(
			"finish",
			gin.H{"status": "validation", "reason": prepare.ValidationError.Message},
		)
		c.Writer.Flush()
		return
	}

	if prepare.Skip {
		c.SSEvent(
			"finish",
			gin.H{"status": "validation", "reason": "Your project has no .ergomake"},
		)
		c.Writer.Flush()
		return
	}

	c.SSEvent("progress", gin.H{"status": "building"})
	c.Writer.Flush()

	tr, err := tt.Transform(ctx, id)
	if err != nil {
		logger.Ctx(c).Err(err).Stack().Str("id", id.String()).Msg("fail to transform view")
		c.SSEvent(
			"finish",
			gin.H{"status": "error"},
		)
		c.Writer.Flush()
		return
	}

	if !tr.IsCompose {
		c.SSEvent(
			"finish",
			gin.H{"status": "validation", "reason": "Did not find a compose file in your project."},
		)
		c.Writer.Flush()
		return
	}

	if tr.IsCompose {
		c.SSEvent("progress", gin.H{"status": "deploying"})
		c.Writer.Flush()

		err = cluster.Deploy(ctx, dr.clusterClient, tr.ClusterEnv)
		if err != nil {
			logger.Ctx(c).Err(err).Stack().Str("id", id.String()).Msg("fail to deploy preview into clustert")
			c.SSEvent(
				"finish",
				gin.H{"status": "error"},
			)
			c.Writer.Flush()
			return
		}

		deploymentsCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
		defer cancel()
		err = dr.clusterClient.WaitDeployments(deploymentsCtx, tr.ClusterEnv.Namespace)
		if err != nil {
			logger.Ctx(c).Err(err).Stack().Str("id", id.String()).Msg("fail to wait for deployment")
			c.SSEvent(
				"finish",
				gin.H{"status": "error"},
			)
			c.Writer.Flush()
			return
		}

		c.SSEvent(
			"finish",
			gin.H{
				"status": "success",
				"url":    fmt.Sprintf("https://%s", tr.Environment.FirstService().Url),
			},
		)
		c.Writer.Flush()
	}
}
