package service

import (
	"context"
	"testing"
	"time"
)

func TestGetClientFailureStatsNormalizesLimitAndUTCWindow(t *testing.T) {
	start := time.Date(2026, 6, 17, 8, 0, 0, 0, time.FixedZone("CST", 8*3600))
	end := start.Add(24 * time.Hour)

	var got *OpsClientFailureStatsFilter
	repo := &opsRepoMock{
		GetClientFailureStatsFn: func(ctx context.Context, filter *OpsClientFailureStatsFilter) (*OpsClientFailureStatsResponse, error) {
			got = filter
			return &OpsClientFailureStatsResponse{StartTime: filter.StartTime, EndTime: filter.EndTime}, nil
		},
	}
	svc := NewOpsService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	_, err := svc.GetClientFailureStats(context.Background(), &OpsClientFailureStatsFilter{
		StartTime: start,
		EndTime:   end,
		Limit:     999,
	})
	if err != nil {
		t.Fatalf("GetClientFailureStats returned error: %v", err)
	}
	if got == nil {
		t.Fatal("repo was not called")
	}
	if got.Limit != OpsClientFailureStatsMaxLimit {
		t.Fatalf("limit = %d, want cap %d", got.Limit, OpsClientFailureStatsMaxLimit)
	}
	if got.StartTime.Location() != time.UTC || got.EndTime.Location() != time.UTC {
		t.Fatalf("times must be normalized to UTC: %#v", got)
	}
}

func TestGetClientFailureStatsRejectsInvalidWindow(t *testing.T) {
	svc := NewOpsService(&opsRepoMock{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	now := time.Now()
	if _, err := svc.GetClientFailureStats(context.Background(), &OpsClientFailureStatsFilter{StartTime: now, EndTime: now}); err == nil {
		t.Fatal("expected error for empty window")
	}
}
