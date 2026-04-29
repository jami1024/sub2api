package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBillingErrorDetails_MapsGroupRPMExceededToTooManyRequests(t *testing.T) {
	status, code, msg, retryAfter := billingErrorDetails(service.ErrGroupRPMExceeded)
	require.Equal(t, http.StatusTooManyRequests, status)
	require.Equal(t, "rate_limit_exceeded", code)
	require.NotEmpty(t, msg)
	require.Greater(t, retryAfter, 0, "RPM exceeded should return positive Retry-After")
	require.LessOrEqual(t, retryAfter, 60)
}

func TestBillingErrorDetails_MapsUserRPMExceededToTooManyRequests(t *testing.T) {
	status, code, msg, retryAfter := billingErrorDetails(service.ErrUserRPMExceeded)
	require.Equal(t, http.StatusTooManyRequests, status)
	require.Equal(t, "rate_limit_exceeded", code)
	require.NotEmpty(t, msg)
	require.Greater(t, retryAfter, 0, "RPM exceeded should return positive Retry-After")
	require.LessOrEqual(t, retryAfter, 60)
}

func TestBillingErrorDetails_APIKeyRateLimitStillMaps(t *testing.T) {
	// 回归保护：加 RPM 分支后不应影响已有 APIKey rate limit 的映射。
	for _, err := range []error{
		service.ErrAPIKeyRateLimit5hExceeded,
		service.ErrAPIKeyRateLimit1dExceeded,
		service.ErrAPIKeyRateLimit7dExceeded,
	} {
		status, code, _, _ := billingErrorDetails(err)
		require.Equal(t, http.StatusTooManyRequests, status, "status for %v", err)
		require.Equal(t, "rate_limit_exceeded", code)
	}
}

func TestBillingErrorDetails_BillingServiceUnavailableMapsTo503(t *testing.T) {
	status, code, _, retryAfter := billingErrorDetails(service.ErrBillingServiceUnavailable)
	require.Equal(t, http.StatusServiceUnavailable, status)
	require.Equal(t, "billing_service_error", code)
	require.Equal(t, 3, retryAfter, "temporary billing failures should ask clients to retry")
}

func TestBillingErrorDetails_UnknownErrorFallsBackTo403(t *testing.T) {
	status, code, msg, _ := billingErrorDetails(service.ErrInsufficientBalance)
	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, "billing_error", code)
	require.NotEmpty(t, msg)
}

func TestCheckBillingEligibilityWithRetry_RetriesTemporaryBillingFailure(t *testing.T) {
	calls := 0
	var delays []time.Duration

	err := checkBillingEligibilityWithRetry(context.Background(), func(context.Context) error {
		calls++
		if calls < 3 {
			return service.ErrBillingServiceUnavailable
		}
		return nil
	}, func(ctx context.Context, delay time.Duration) error {
		delays = append(delays, delay)
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 3, calls)
	require.Len(t, delays, 2)
	require.GreaterOrEqual(t, delays[0], 100*time.Millisecond)
	require.GreaterOrEqual(t, delays[1], 300*time.Millisecond)
}

func TestCheckBillingEligibilityWithRetry_DoesNotRetryBusinessErrors(t *testing.T) {
	calls := 0

	err := checkBillingEligibilityWithRetry(context.Background(), func(context.Context) error {
		calls++
		return service.ErrInsufficientBalance
	}, func(context.Context, time.Duration) error {
		t.Fatal("business errors should not sleep or retry")
		return nil
	})

	require.ErrorIs(t, err, service.ErrInsufficientBalance)
	require.Equal(t, 1, calls)
}

func TestCheckBillingEligibilityWithRetry_ReturnsTemporaryFailureAfterRetries(t *testing.T) {
	calls := 0
	rootErr := errors.New("redis timeout")

	err := checkBillingEligibilityWithRetry(context.Background(), func(context.Context) error {
		calls++
		return service.ErrBillingServiceUnavailable.WithCause(rootErr)
	}, func(context.Context, time.Duration) error {
		return nil
	})

	require.ErrorIs(t, err, service.ErrBillingServiceUnavailable)
	require.Equal(t, 3, calls)
}
