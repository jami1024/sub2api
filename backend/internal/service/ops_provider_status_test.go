package service

import (
	"testing"
	"time"
)

func TestNormalizeProviderStatusFilterCapsBucketsAndLimit(t *testing.T) {
	start := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)
	end := start.Add(7 * 24 * time.Hour)

	got, err := normalizeProviderStatusFilter(&OpsProviderStatusFilter{
		StartTime: start,
		EndTime:   end,
		Limit:     1000,
	})
	if err != nil {
		t.Fatalf("normalizeProviderStatusFilter returned error: %v", err)
	}
	if got.Limit != OpsProviderStatusMaxLimit {
		t.Fatalf("limit = %d, want cap %d", got.Limit, OpsProviderStatusMaxLimit)
	}
	buckets := int(got.EndTime.Sub(got.StartTime).Seconds()/float64(got.BucketSeconds)) + 1
	if buckets > OpsProviderStatusMaxBuckets {
		t.Fatalf("bucket count = %d, want <= %d (bucket_seconds=%d)", buckets, OpsProviderStatusMaxBuckets, got.BucketSeconds)
	}
}

func TestNormalizeProviderStatusFilterRejectsInvalidWindow(t *testing.T) {
	now := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)
	if _, err := normalizeProviderStatusFilter(&OpsProviderStatusFilter{StartTime: now, EndTime: now}); err == nil {
		t.Fatal("expected invalid zero-length window error")
	}
	if _, err := normalizeProviderStatusFilter(&OpsProviderStatusFilter{StartTime: now, EndTime: now.Add(-time.Minute)}); err == nil {
		t.Fatal("expected invalid reversed window error")
	}
}
