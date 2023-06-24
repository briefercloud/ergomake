package payment

import "context"

type PaymentPlan string

const (
	PaymentPlanFree         PaymentPlan = "free"
	PaymentPlanStandard     PaymentPlan = "standard"
	PaymentPlanProfessional PaymentPlan = "professional"
)

func (plan *PaymentPlan) ActiveEnvironmentsLimit() int {
	switch *plan {
	case PaymentPlanFree:
		return 1
	case PaymentPlanStandard:
		return 3
	case PaymentPlanProfessional:
		return 8
	}

	panic("unreachable")
}

func (plan *PaymentPlan) PermanentEnvironmentsLimit() int {
	switch *plan {
	case PaymentPlanFree:
		return 0
	case PaymentPlanStandard:
		return 1
	case PaymentPlanProfessional:
		return 4
	}

	panic("unreachable")
}

const StandardPlanEnvLimit = 10

type PaymentProvider interface {
	SaveSubscription(ctx context.Context, owner, subscriptionID string) error
	GetOwnerPlan(ctx context.Context, owner string) (PaymentPlan, error)
}
