package service

import (
	"context"
	"fmt"
	"time"
)

const (
	OpsProviderStatusDefaultBucketSeconds = 60
	OpsProviderStatusMaxBuckets           = 180
	OpsProviderStatusDefaultLimit         = 50
	OpsProviderStatusMaxLimit             = 100
)

func (s *OpsService) GetProviderStatus(ctx context.Context, filter *OpsProviderStatusFilter) (*OpsProviderStatusResponse, error) {
	if err := s.RequireMonitoringEnabled(ctx); err != nil {
		return nil, err
	}
	if s.opsRepo == nil {
		return &OpsProviderStatusResponse{}, nil
	}
	normalized, err := normalizeProviderStatusFilter(filter)
	if err != nil {
		return nil, err
	}
	return s.opsRepo.GetProviderStatus(ctx, normalized)
}

func normalizeProviderStatusFilter(filter *OpsProviderStatusFilter) (*OpsProviderStatusFilter, error) {
	if filter == nil {
		return nil, fmt.Errorf("nil filter")
	}
	if filter.StartTime.IsZero() || filter.EndTime.IsZero() || !filter.StartTime.Before(filter.EndTime) {
		return nil, fmt.Errorf("valid start_time/end_time required")
	}
	start := filter.StartTime.UTC()
	end := filter.EndTime.UTC()
	bucketSeconds := filter.BucketSeconds
	if bucketSeconds <= 0 {
		bucketSeconds = pickProviderStatusBucketSeconds(end.Sub(start))
	}
	if bucketSeconds <= 0 {
		bucketSeconds = OpsProviderStatusDefaultBucketSeconds
	}
	bucketCount := int(end.Sub(start).Seconds()/float64(bucketSeconds)) + 1
	if bucketCount > OpsProviderStatusMaxBuckets {
		bucketSeconds = int(end.Sub(start).Seconds()/float64(OpsProviderStatusMaxBuckets)) + 1
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = OpsProviderStatusDefaultLimit
	}
	if limit > OpsProviderStatusMaxLimit {
		limit = OpsProviderStatusMaxLimit
	}
	return &OpsProviderStatusFilter{
		StartTime:     start,
		EndTime:       end,
		BucketSeconds: bucketSeconds,
		Limit:         limit,
	}, nil
}

func pickProviderStatusBucketSeconds(window time.Duration) int {
	switch {
	case window <= 2*time.Hour:
		return 60
	case window <= 24*time.Hour:
		return 300
	default:
		return 3600
	}
}
