package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/core/billing"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres"
	"github.com/kitecloud/kite/kite-service/internal/logging"
	"github.com/urfave/cli/v2"
)

var billingCMD = cli.Command{
	Name:  "billing",
	Usage: "Billing commands",
	Subcommands: []*cli.Command{
		{
			Name: "reconcile",
			Usage: "Bring stored subscriptions back in line with LemonSqueezy. " +
				"LemonSqueezy cannot replay webhooks, so this is how missed status changes are recovered.",
			Action: billingReconcileCMD,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "dry_run",
					Usage: "Report what would change without writing anything.",
				},
				&cli.DurationFlag{
					Name:  "delay",
					Usage: "Wait this long between LemonSqueezy API calls to stay inside the rate limit.",
					Value: 250 * time.Millisecond,
				},
			},
		},
	},
}

func billingReconcileCMD(c *cli.Context) error {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logging.SetupLogger(cfg.Logging)

	pg, err := postgres.New(postgres.BuildConnectionDSN(cfg.Database.Postgres), 1)
	if err != nil {
		return fmt.Errorf("failed to create postgres client: %w", err)
	}

	planManager := plan.NewPlanManager(pg, pg, pg, plan.PlansFromConfig(cfg.Billing.Plans), plan.PlanManagerConfig{})

	subscriptionManager := billing.NewSubscriptionManager(pg, pg, planManager, billing.NewLemonSqueezyClient(
		cfg.Billing.LemonSqueezyAPIKey,
		cfg.Billing.LemonSqueezySigningSecret,
	))

	dryRun := c.Bool("dry_run")
	if dryRun {
		slog.Info("Running in dry run mode, nothing will be written")
	}

	res, err := subscriptionManager.Reconcile(context.Background(), dryRun, c.Duration("delay"))

	// The result is worth printing even on failure, because it says how far the
	// run got before it stopped.
	slog.Info(
		"Reconciled subscriptions",
		slog.Int("total", res.Total),
		slog.Int("changed", res.Changed),
		slog.Int("unchanged", res.Unchanged),
		slog.Int("skipped", res.Skipped),
		slog.Int("failed", res.Failed),
	)

	if err != nil {
		return fmt.Errorf("failed to reconcile subscriptions: %w", err)
	}

	return nil
}
