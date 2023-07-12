package v2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"

	"github.com/ergomake/ergomake/e2e/testutils"
	"github.com/ergomake/ergomake/internal/api"
	"github.com/ergomake/ergomake/internal/cluster"
	environmentsMocks "github.com/ergomake/ergomake/mocks/environments"
	envvarsMocks "github.com/ergomake/ergomake/mocks/envvars"
	ghAppMocks "github.com/ergomake/ergomake/mocks/github/ghapp"
	ghlauncherMocks "github.com/ergomake/ergomake/mocks/github/ghlauncher"
	paymentMocks "github.com/ergomake/ergomake/mocks/payment"
	permanentbranchesMocks "github.com/ergomake/ergomake/mocks/permanentbranches"
	privregistryMocks "github.com/ergomake/ergomake/mocks/privregistry"
	servicelogsMocks "github.com/ergomake/ergomake/mocks/servicelogs"
	usersMocks "github.com/ergomake/ergomake/mocks/users"
)

func TestV2Health(t *testing.T) {
	testCases := []struct {
		name   string
		status int
	}{
		{
			name:   "returns ok",
			status: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clusterClient, err := cluster.NewK8sClient()
			require.NoError(t, err)

			db := testutils.CreateRandomDB(t)

			ghApp := ghAppMocks.NewGHAppClient(t)
			apiServer := api.NewServer(
				ghlauncherMocks.NewGHLauncher(t),
				privregistryMocks.NewPrivRegistryProvider(t),
				db,
				servicelogsMocks.NewLogStreamer(t),
				ghApp,
				clusterClient,
				envvarsMocks.NewEnvVarsProvider(t),
				environmentsMocks.NewEnvironmentsProvider(t),
				usersMocks.NewService(t),
				paymentMocks.NewPaymentProvider(t),
				permanentbranchesMocks.NewPermanentBranchesProvider(t),
				&api.Config{},
			)
			server := httptest.NewServer(apiServer)

			e := httpexpect.Default(t, server.URL)
			e.GET("/v2/health").Expect().Status(tc.status)
		})
	}
}
