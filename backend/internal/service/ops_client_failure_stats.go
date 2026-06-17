package service

import (
	"context"
	"fmt"
)

const (
	OpsClientFailureStatsDefaultLimit = 50
	OpsClientFailureStatsMaxLimit     = 100
)

func (s *OpsService) GetClientFailureStats(ctx context.Context, filter *OpsClientFailureStatsFilter) (*OpsClientFailureStatsResponse, error) {
	if err := s.RequireMonitoringEnabled(ctx); err != nil {
		return nil, err
	}
	if s.opsRepo == nil {
		return &OpsClientFailureStatsResponse{}, nil
	}
	normalized, err := normalizeClientFailureStatsFilter(filter)
	if err != nil {
		return nil, err
	}
	return s.opsRepo.GetClientFailureStats(ctx, normalized)
}

func normalizeClientFailureStatsFilter(filter *OpsClientFailureStatsFilter) (*OpsClientFailureStatsFilter, error) {
	if filter == nil {
		return nil, fmt.Errorf("nil filter")
	}
	if filter.StartTime.IsZero() || filter.EndTime.IsZero() || !filter.StartTime.Before(filter.EndTime) {
		return nil, fmt.Errorf("valid start_time/end_time required")
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = OpsClientFailureStatsDefaultLimit
	}
	if limit > OpsClientFailureStatsMaxLimit {
		limit = OpsClientFailureStatsMaxLimit
	}
	return &OpsClientFailureStatsFilter{
		StartTime: filter.StartTime.UTC(),
		EndTime:   filter.EndTime.UTC(),
		Limit:     limit,
	}, nil
}
