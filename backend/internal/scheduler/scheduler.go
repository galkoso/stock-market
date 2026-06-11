package scheduler

import (
	"context"
	"log"
	"time"

	"stock-market/backend/internal/provider/marketdata"
	"stock-market/backend/internal/repositories"
	"stock-market/backend/internal/services"
)

type Scheduler struct {
	provider      marketdata.Provider
	alerts        *repositories.AlertsRepository
	alertEngine   *services.AlertEngine
}

func New(
	provider marketdata.Provider,
	_ *repositories.WatchlistRepository,
	alerts *repositories.AlertsRepository,
	_ *repositories.NotificationsRepository,
	alertEngine *services.AlertEngine,
) *Scheduler {
	return &Scheduler{
		provider:    provider,
		alerts:      alerts,
		alertEngine: alertEngine,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	go func() {
		s.runMorningJobs(ctx)
		s.runAlertEvaluation(ctx)

		morningTicker := time.NewTicker(24 * time.Hour)
		alertTicker := time.NewTicker(2 * time.Minute)
		defer morningTicker.Stop()
		defer alertTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-morningTicker.C:
				s.runMorningJobs(ctx)
			case <-alertTicker.C:
				s.runAlertEvaluation(ctx)
			}
		}
	}()
}

func (s *Scheduler) runMorningJobs(ctx context.Context) {
	log.Println("scheduler: running morning refresh jobs")

	now := time.Now().UTC()
	from := now.Format("2006-01-02")
	to := now.AddDate(0, 0, 14).Format("2006-01-02")

	if _, err := s.provider.GetEarningsCalendar(ctx, from, to); err != nil {
		log.Printf("scheduler: earnings refresh failed: %v", err)
	}
}

func (s *Scheduler) runAlertEvaluation(ctx context.Context) {
	log.Println("scheduler: evaluating alerts")
	s.alertEngine.Evaluate(ctx)
}
