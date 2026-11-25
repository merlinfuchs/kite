package cmd

import (
	"context"
	"fmt"
	"log/slog"

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
			Usage: "Bring stored subscriptions back in line with LemonSqueezy and Paddle. " +
				"LemonSqueezy cannot replay webhooks, so this is how missed changes are recovered.",
			Action: billingReconcileCMD,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "dry_run",
					Usage: "Report what would change without writing anything.",
				},
				&cli.DurationFlag{
					Name:  "delay",
					Usage: "Wait this long between API calls to stay inside the rate limits.",
					Value: billing.ReconcileDelay,
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

	planManager := plan.NewPlanManager(pg, pg, pg, pg, plan.PlansFromConfig(cfg.Billing.Plans), plan.PlanManagerConfig{})

	paddleClient, err := billing.NewPaddleClient(cfg.Billing.Paddle)
	if err != nil {
		return fmt.Errorf("failed to create paddle client: %w", err)
	}

	subscriptionManager := billing.NewSubscriptionManager(pg, pg, planManager, billing.NewLemonSqueezyClient(
		cfg.Billing.LemonSqueezyAPIKey,
		cfg.Billing.LemonSqueezySigningSecret,
	), paddleClient)

	dryRun := c.Bool("dry_run")
	if dryRun {
		slog.Info("Running in dry run mode, nothing will be written")
	}

	_, err = subscriptionManager.Reconcile(context.Background(), dryRun, c.Duration("delay"))
	if err != nil {
		return fmt.Errorf("failed to reconcile subscriptions: %w", err)
	}

	return nil
}
