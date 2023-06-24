package github

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/go-github/v52/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ergomake/ergomake/e2e/testutils"
	"github.com/ergomake/ergomake/internal/api"
	"github.com/ergomake/ergomake/internal/database"
	clusterMocks "github.com/ergomake/ergomake/mocks/cluster"
	environmentsMocks "github.com/ergomake/ergomake/mocks/environments"
	envvarsMocks "github.com/ergomake/ergomake/mocks/envvars"
	ghAppMocks "github.com/ergomake/ergomake/mocks/github/ghapp"
	paymentMocks "github.com/ergomake/ergomake/mocks/payment"
	servicelogsMocks "github.com/ergomake/ergomake/mocks/servicelogs"
	usersMocks "github.com/ergomake/ergomake/mocks/users"
)

func getEvent(account string) *github.MarketplacePurchaseEvent {
	action := "purchased"
	sender := "sender"

	return &github.MarketplacePurchaseEvent{
		Action: &action,
		MarketplacePurchase: &github.MarketplacePurchase{
			Account: &github.MarketplacePurchaseAccount{
				Login: &account,
			},
		},
		Sender: &github.User{
			Login: &sender,
		},
	}
}

func TestMarketplaceWebhook(t *testing.T) {
	type tc struct {
		name           string
		headers        map[string]string
		payload        interface{}
		expectedStatus int
		assertFn       func(t *testing.T, tc tc, db *database.DB)
	}
	testCases := []tc{
		{
			name: "valid payload gives 204 and creates event in database",
			headers: map[string]string{
				"X-Hub-Signature-256": genSignature(t, getEvent("account")),
				"X-GitHub-Event":      "marketplace_purchase",
			},
			payload:        getEvent("account"),
			expectedStatus: http.StatusNoContent,
			assertFn: func(t *testing.T, tc tc, db *database.DB) {
				event := tc.payload.(*github.MarketplacePurchaseEvent)
				owner := event.MarketplacePurchase.GetAccount().GetLogin()

				var events []*database.MarketplaceEvent
				result := db.Find(&events, "owner = ?", owner)
				require.NoError(t, result.Error)

				assert.Len(t, events, 1)
				assert.Equal(t, events[0].Owner, owner)
				assert.Equal(t, events[0].Action, event.GetAction())
			},
		},
		{
			name: "missing signature gives 401",
			headers: map[string]string{
				"X-GitHub-Event": "marketplace_purchase",
			},
			payload:        "whatever",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid signature gives 401",
			headers: map[string]string{
				"X-Hub-Signature-256": "sha256=invalid",
				"X-GitHub-Event":      "marketplace_purchase",
			},
			payload:        "whatever",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid payload gives 400",
			headers: map[string]string{
				"X-Hub-Signature-256": genSignature(t, "invalid"),
				"X-GitHub-Event":      "marketplace_purchase",
			},
			payload:        "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clusterClient := clusterMocks.NewClient(t)

			cfg := &api.Config{
				GithubWebhookSecret: "secret",
			}

			db := testutils.CreateRandomDB(t)

			ghApp := ghAppMocks.NewGHAppClient(t)
			apiServer := api.NewServer(db, servicelogsMocks.NewLogStreamer(t), ghApp, clusterClient, envvarsMocks.NewEnvVarsProvider(t),
				environmentsMocks.NewEnvironmentsProvider(t), usersMocks.NewService(t), paymentMocks.NewPaymentProvider(t), cfg)

			server := httptest.NewServer(apiServer)
			e := httpexpect.Default(t, server.URL)
			e.POST("/v2/github/marketplace/webhook").WithHeaders(tc.headers).
				WithJSON(tc.payload).Expect().Status(tc.expectedStatus)

			if tc.assertFn != nil {
				tc.assertFn(t, tc, db)
			}
		})
	}
}
