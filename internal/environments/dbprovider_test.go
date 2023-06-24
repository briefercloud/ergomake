package environments

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ergomake/ergomake/e2e/testutils"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/payment"
	paymentMocks "github.com/ergomake/ergomake/mocks/payment"
)

func TestDBEnvironmentsProvider_IsOwnerLimited(t *testing.T) {
	t.Parallel()

	setup := func(t *testing.T) *database.DB {
		db := testutils.CreateRandomDB(t)

		limitedEnv := database.Environment{
			Owner:  "owner",
			Status: database.EnvLimited,
		}
		err := db.Save(&limitedEnv).Error
		require.NoError(t, err)

		nonLimitedEnv := database.Environment{
			Owner:  "owner",
			Status: database.EnvSuccess,
		}
		err = db.Save(&nonLimitedEnv).Error
		require.NoError(t, err)

		otherOwnerEnv := database.Environment{
			Owner:  "other-owner",
			Status: database.EnvSuccess,
		}
		err = db.Save(&otherOwnerEnv).Error
		require.NoError(t, err)

		return db
	}

	paymentProvider := paymentMocks.NewPaymentProvider(t)
	paymentProvider.EXPECT().GetOwnerPlan(mock.Anything, mock.Anything).Return(payment.PaymentPlanFree, nil)

	tt := []struct {
		name           string
		limit          int
		want           bool
		ownerLimit     int
		paymetProvider payment.PaymentProvider
	}{
		{
			name:           "when non limited envs count lower than envLimitAmount",
			limit:          2,
			want:           false,
			paymetProvider: paymentProvider,
		},
		{
			name:           "when non limited envs count equal than envLimitAmount",
			limit:          1,
			want:           true,
			paymetProvider: paymentProvider,
		},
		{
			name:           "when non limited envs count greater than envLimitAmount",
			limit:          0,
			want:           true,
			paymetProvider: paymentProvider,
		},
		{
			name:           "when owner has specific configuration",
			limit:          2,
			ownerLimit:     1,
			want:           true,
			paymetProvider: paymentMocks.NewPaymentProvider(t),
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := setup(t)

			if tc.ownerLimit > 0 {
				err := db.Save(&environmentLimits{
					Owner:    "owner",
					EnvLimit: tc.ownerLimit,
				}).Error
				require.NoError(t, err)
			}

			ep := NewDBEnvironmentsProvider(db, paymentProvider, tc.limit)
			limited, err := ep.IsOwnerLimited(context.Background(), "owner")
			require.NoError(t, err)
			assert.Equal(t, tc.want, limited)
		})
	}
}
