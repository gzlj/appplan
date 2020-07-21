package workloadplan

import (
	"context"

	planv1 "github.com/gzlj/appplan/pkg/apis/plan/v1"
)

func ConvertFromCr(plan *planv1.WorkLoadPlan) IPlan {
	p := NormalPlan{}
	p.WorkLoadPlan = plan
	p.waitSeconds = 15
	if plan.Spec.Cron != nil {


		if plan.Spec.Cron.Period == CRON_PERIOD_MONTH {
			p.waitSeconds = 30
		} else if plan.Spec.Cron.Period == CRON_PERIOD_YEAR {
			p.waitSeconds = 45
		}
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	p.ctx = ctx
	p.cancelFunc = cancelFunc
	return &p
}
