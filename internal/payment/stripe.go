package payment

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/ergomake/ergomake/internal/database"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/subscription"
)

type stripeSubscription struct {
	ID             uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt      time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt      *gorm.DeletedAt `gorm:"index"`
	Owner          string          `gorm:"index;type:varchar(255);not null"`
	SubscriptionID string          `gorm:"index;type:varchar(255);not null"`
}

type stripePaymentProvider struct {
	db                        *database.DB
	secretKey                 string
	standardPlanProductID     string
	professionalPlanProductID string
	friends                   map[string]struct{}
}

func NewStripePaymentProvider(
	db *database.DB,
	secretKey, standardPlanProductID, professionalPlanProductID string,
	friendsArr []string,
) *stripePaymentProvider {
	friends := make(map[string]struct{})
	for _, friend := range friendsArr {
		friends[friend] = struct{}{}
	}

	return &stripePaymentProvider{db, secretKey, standardPlanProductID, professionalPlanProductID, friends}
}

func (stp *stripePaymentProvider) SaveSubscription(ctx context.Context, owner, subscriptionID string) error {
	subscription := stripeSubscription{
		Owner:          owner,
		SubscriptionID: subscriptionID,
	}

	err := stp.db.
		Where(subscription).
		Assign(stripeSubscription{UpdatedAt: time.Now()}).
		FirstOrCreate(&subscription).
		Error
	if err != nil {
		return errors.Wrap(err, "failed to save subscription to database")
	}

	return nil
}

func (stp *stripePaymentProvider) GetOwnerPlan(ctx context.Context, owner string) (PaymentPlan, error) {
	stripe.Key = stp.secretKey

	if _, isFriend := stp.friends[owner]; isFriend {
		return PaymentPlanProfessional, nil
	}

	var dbSubs []stripeSubscription
	err := stp.db.Where("owner = ?", owner).Find(&dbSubs).Error
	if err != nil {
		return PaymentPlanFree, errors.Wrapf(err, "fail to list %s subscriptions from db", owner)
	}

	plan := PaymentPlanFree

	for _, dbSub := range dbSubs {
		stripeSub, err := subscription.Get(
			dbSub.SubscriptionID,
			&stripe.SubscriptionParams{Params: stripe.Params{Context: ctx}},
		)
		if err != nil {
			return PaymentPlanFree, errors.Wrapf(err, "fail to get stripe subscription %s", dbSub.SubscriptionID)
		}

		if stripeSub.Status == stripe.SubscriptionStatusActive || stripeSub.Status == stripe.SubscriptionStatusTrialing {
			for _, item := range stripeSub.Items.Data {
				if item.Price.Product.ID == stp.professionalPlanProductID {
					return PaymentPlanProfessional, nil
				}

				if item.Price.Product.ID == stp.standardPlanProductID {
					plan = PaymentPlanStandard
				}
			}
		}
	}

	return plan, nil
}
