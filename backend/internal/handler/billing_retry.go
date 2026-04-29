package handler

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const billingServiceUnavailableRetryAfterSeconds = 3

var billingEligibilityRetryBaseDelays = []time.Duration{
	100 * time.Millisecond,
	300 * time.Millisecond,
}

type billingEligibilityCheckFunc func(context.Context) error
type billingEligibilitySleepFunc func(context.Context, time.Duration) error

func checkBillingEligibilityWithRetry(
	ctx context.Context,
	check billingEligibilityCheckFunc,
	sleep billingEligibilitySleepFunc,
) error {
	if check == nil {
		return nil
	}
	if sleep == nil {
		sleep = sleepBillingEligibilityRetry
	}

	var lastErr error
	for attempt := 0; attempt <= len(billingEligibilityRetryBaseDelays); attempt++ {
		err := check(ctx)
		if err == nil {
			return nil
		}
		lastErr = err
		if !errors.Is(err, service.ErrBillingServiceUnavailable) {
			return err
		}
		if attempt == len(billingEligibilityRetryBaseDelays) {
			break
		}
		if sleepErr := sleep(ctx, billingRetryDelayWithJitter(billingEligibilityRetryBaseDelays[attempt])); sleepErr != nil {
			return err
		}
	}
	return lastErr
}

func sleepBillingEligibilityRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func billingRetryDelayWithJitter(base time.Duration) time.Duration {
	const maxJitter = 50 * time.Millisecond
	if base <= 0 {
		return 0
	}

	var buf [8]byte
	if _, err := crand.Read(buf[:]); err != nil {
		return base
	}
	jitter := time.Duration(binary.LittleEndian.Uint64(buf[:]) % uint64(maxJitter))
	return base + jitter
}
