package vuforia

import (
	"context"
	"errors"
	"strings"
	"time"
)

func WaitUntilProcessed(ctx context.Context, client Client, target string) error {
	defaultInterval := 5 * time.Second
	extendedInterval := 30 * time.Second
	interval := defaultInterval
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
			interval = defaultInterval

			output, err := client.GetTarget(ctx, &GetTargetRequest{TargetId: target})
			if err != nil {
				var ae APIError
				if errors.As(err, &ae) && strings.EqualFold(ae.ResultCode, "RequestQuotaReached") {
					interval = extendedInterval
					continue
				}

				return err
			}

			switch strings.ToLower(output.Status) {
			case "failed", "success":
				return nil
			case "processing":
				continue
			default:
				continue
			}
		}
	}
}
