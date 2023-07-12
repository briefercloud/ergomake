package stale

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/environments"
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/payment"
)

type server struct {
	*gin.Engine
	clusterClient        cluster.Client
	environmentsProvider environments.EnvironmentsProvider
	paymentProvider      payment.PaymentProvider
	frontendURL          string
	timeoutToStale       time.Duration
	ingressNamespace     string
	ingressServiceName   string
}

func NewServer(
	clusterClient cluster.Client,
	environmentsProvider environments.EnvironmentsProvider,
	paymentProvider payment.PaymentProvider,
	frontendURL string,
	timeoutToStale time.Duration,
	ingressNamespace string,
	ingressServiceName string,
) *server {
	router := gin.New()

	s := &server{
		router,
		clusterClient,
		environmentsProvider,
		paymentProvider,
		frontendURL,
		timeoutToStale,
		ingressNamespace,
		ingressServiceName,
	}

	router.Use(gin.Recovery())
	logger.Middleware(router)
	router.NoRoute(s.handle)

	return s
}

func (s *server) handle(c *gin.Context) {
	host := c.Request.Host

	logger.Ctx(c).Info().Interface("host", host).Msg("got a request for a stale environment")

	env, err := s.environmentsProvider.GetEnvironmentFromHost(c, host)
	if err != nil {
		if errors.Is(err, environments.ErrEnvironmentNotFound) {
			c.JSON(http.StatusNotFound, http.StatusText(http.StatusNotFound))
			return
		}

		logger.Ctx(c).Err(err).Str("host", host).Msg("fail to get environment from host")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	logger.Ctx(c).Info().Interface("env", env).Msg("stale env")

	if env.Status == database.EnvStale {
		namespace := env.ID.String()
		for _, svc := range env.Services {
			err := s.clusterClient.ScaleDeployment(c, namespace, svc.Name, 1)
			if err != nil {
				logger.Ctx(c).Err(err).Str("host", host).Str("service", svc.Name).Msg("fail to scale deployment up")
				c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				return
			}
			ingress, err := s.clusterClient.GetIngress(c, namespace, svc.Name)
			if errors.Is(err, cluster.ErrIngressNotFound) {
				continue
			}

			if err != nil {
				logger.Ctx(c).Err(err).Str("host", host).Str("service", svc.Name).Msg("fail to get ingress of service")
				c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				return
			}

			if len(ingress.Spec.Rules) > 0 {
				ingress.Spec.Rules[0].Host = svc.Url
			}

			err = s.clusterClient.UpdateIngress(c, ingress)
			if err != nil {
				logger.Ctx(c).Err(err).Str("host", host).Str("service", svc.Name).Msg("fail to get ingress of service")
				c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				return
			}
		}
	}

	go func() {
		log := logger.Get()
		ctx, cancel := context.WithTimeout(log.WithContext(context.Background()), time.Minute*10)
		defer cancel()

		err := s.clusterClient.WaitDeployments(ctx, env.ID.String())
		if err != nil {
			logger.Ctx(ctx).Err(err).Str("env", env.ID.String()).Msg("fail to wait deployments")
			return
		}

		env.Status = database.EnvSuccess
		err = s.environmentsProvider.SaveEnvironment(ctx, env)
		if err != nil {
			logger.Ctx(ctx).Err(err).Str("env", env.ID.String()).Msg("fail to set env status to success")
			return
		}
	}()

	c.Redirect(
		http.StatusTemporaryRedirect,
		fmt.Sprintf("%s/environments/%s?redirect=%s", s.frontendURL, env.ID, host),
	)
}

func (s *server) Listen(ctx context.Context, addr string) error {
	return s.Run(addr)
}

func parseNamespace(line string) string {
	re := regexp.MustCompile(`\[[^\]]*\]`)
	match := re.FindAll([]byte(line), -1)
	if len(match) < 2 {
		return ""
	}

	ns := string(match[len(match)-2])
	if len(ns) < 2 {
		return ""
	}

	ns = string(ns[1 : len(ns)-1])
	parts := strings.Split(ns, "-")
	if len(parts) < 3 {
		return ""
	}

	uuidParts := parts[0 : len(parts)-2]
	ns = strings.Join(uuidParts, "-")

	return ns
}

func parseTimestamp(logLine string) (time.Time, error) {
	re := regexp.MustCompile(`\[(.*?)\]`)
	match := re.FindStringSubmatch(logLine)
	if len(match) != 2 {
		return time.Time{}, errors.New("timestamp not found in the log line")
	}

	timestampStr := match[1]
	layout := "02/Jan/2006:15:04:05 -0700"
	timestamp, err := time.Parse(layout, timestampStr)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "fail to parse %s as time.Time", timestampStr)
	}

	return timestamp, nil
}

func (s *server) MonitorStaleServices(ctx context.Context) {
	lastRequestAtByEnvironments := make(map[string]time.Time)
	go func() {
		for {
			time.Sleep(time.Second * 2)

			ctx, cancel := context.WithCancel(ctx)
			sinceSeconds := int64(time.Hour.Seconds())
			logs, errCh, err := s.clusterClient.WatchServiceLogs(ctx, s.ingressNamespace, s.ingressServiceName, sinceSeconds)

			if err != nil {
				logger.Ctx(ctx).Err(err).Msg("something went wrong while watching nginx logs, will restart")
				cancel()
				continue
			}

		outer:
			for {
				select {
				case logEntry := <-logs:
					nsStr := parseNamespace(logEntry)
					if nsStr == "" {
						continue
					}

					if strings.HasPrefix(nsStr, "preview-core") {
						continue
					}

					ns, err := uuid.Parse(nsStr)
					if err != nil {
						logger.Ctx(ctx).Err(err).Str("namespace", nsStr).
							Msg("fail to parse namespace extracted from access logs")
						continue
					}

					timestamp, err := parseTimestamp(logEntry)
					if err != nil {
						logger.Ctx(ctx).Err(err).Msg("fail to parse timestamp extracted from access logs")
						continue
					}

					logger.Ctx(ctx).Info().Str("ns", ns.String()).Str("time", timestamp.String()).Msg("updating last request at")
					lastRequestAtByEnvironments[ns.String()] = timestamp

				case err := <-errCh:
					if err != nil {
						logger.Ctx(ctx).Err(err).Msg("something went wrong while watching nginx logs, will restart")
					}
					cancel()
					break outer
				}
			}
		}
	}()

	first := true
	for {
		if !first {
			time.Sleep(time.Second * 30)
		}
		first = false

		envs, err := s.environmentsProvider.ListSuccessEnvironments(ctx)
		if err != nil {
			logger.Ctx(ctx).Err(err).Msg("fail to list success environments")
			continue
		}

		envsByOwner := make(map[string][]*database.Environment)
		for _, env := range envs {
			if !env.PullRequest.Valid {
				// branch environments never die
				continue
			}

			ownerEnvs, ok := envsByOwner[env.Owner]
			if !ok {
				ownerEnvs = make([]*database.Environment, 0)
			}

			envsByOwner[env.Owner] = append(ownerEnvs, env)
		}

		envsToDownscale := []*database.Environment{}
		for owner, envs := range envsByOwner {
			plan, err := s.paymentProvider.GetOwnerPlan(ctx, owner)
			if err != nil {
				logger.Ctx(ctx).Err(err).Str("owner", owner).Msg("fail to get owner plan")
				continue
			}

			sort.SliceStable(envs, func(x, y int) bool {
				isXBranch := !envs[x].PullRequest.Valid
				isYBranch := !envs[y].PullRequest.Valid

				if isYBranch {
					return true
				}

				if isXBranch {
					return false
				}

				xLastUsedAt, ok := lastRequestAtByEnvironments[envs[x].ID.String()]
				if !ok {
					xLastUsedAt = envs[x].UpdatedAt
				}

				yLastUsedAt, ok := lastRequestAtByEnvironments[envs[y].ID.String()]
				if !ok {
					yLastUsedAt = envs[y].UpdatedAt
				}

				return xLastUsedAt.Before(yLastUsedAt)
			})

			activeLimit := plan.ActiveEnvironmentsLimit()
			permanentLimit := plan.PermanentEnvironmentsLimit()

			if len(envs) > activeLimit {
				ownerEnvsToDownscale := []*database.Environment{}
				for _, env := range envs {
					ownerEnvsToDownscale = append(ownerEnvsToDownscale, env)

					if len(envs)-len(ownerEnvsToDownscale) <= activeLimit {
						break
					}
				}
				envsToDownscale = append(envsToDownscale, ownerEnvsToDownscale...)
			} else if len(envs) > permanentLimit {
				ownerEnvsToDownscale := []*database.Environment{}
				for _, env := range envs {
					if time.Since(env.UpdatedAt) >= s.timeoutToStale {
						ownerEnvsToDownscale = append(ownerEnvsToDownscale, env)
					}

					if len(envs)-len(ownerEnvsToDownscale) <= permanentLimit {
						break
					}
				}
				envsToDownscale = append(envsToDownscale, ownerEnvsToDownscale...)
			}
		}

	outer:
		for _, env := range envsToDownscale {
			ns := env.ID.String()

			env.Status = database.EnvStale

			for _, svc := range env.Services {
				err = s.clusterClient.ScaleDeployment(ctx, ns, svc.Name, 0)
				if err != nil {
					logger.Ctx(ctx).Err(err).Str("service", svc.Name).Str("env", ns).
						Msg("fail to scale down deployment to stale environment")
					env.Status = database.EnvDegraded
					break
				}

				ingress, err := s.clusterClient.GetIngress(ctx, ns, svc.Name)
				if err != nil && !errors.Is(err, cluster.ErrIngressNotFound) {
					logger.Ctx(ctx).Err(err).Str("service", svc.Name).Str("env", ns).
						Msg("fail to get ingress to stale environment")
					continue outer
				}

				if errors.Is(err, cluster.ErrIngressNotFound) || len(ingress.Spec.Rules) <= 0 {
					continue
				}

				ingress.Spec.Rules[0].Host = fmt.Sprintf("stale-%s", ingress.Spec.Rules[0].Host)
				err = s.clusterClient.UpdateIngress(ctx, ingress)
				if err != nil {
					logger.Ctx(ctx).Err(err).Str("service", svc.Name).Str("env", ns).Msg("fail to update ingress to stale environment")
					continue outer
				}
			}

			err = s.environmentsProvider.SaveEnvironment(ctx, env)
			if err != nil {
				logger.Ctx(ctx).Err(err).Str("env", ns).Str("status", string(env.Status)).Msg("fail to update environment status")
			}
		}
	}
}
