package service

import (
	"context"
	"testing"
	"time"
)

type channelMonitorUsageRepoStub struct {
	monitor *ChannelMonitor
	latest  map[string]*ChannelMonitorUsageLogLatest
}

func (s *channelMonitorUsageRepoStub) Create(context.Context, *ChannelMonitor) error { return nil }
func (s *channelMonitorUsageRepoStub) GetByID(context.Context, int64) (*ChannelMonitor, error) {
	return s.monitor, nil
}
func (s *channelMonitorUsageRepoStub) Update(context.Context, *ChannelMonitor) error { return nil }
func (s *channelMonitorUsageRepoStub) Delete(context.Context, int64) error           { return nil }
func (s *channelMonitorUsageRepoStub) List(context.Context, ChannelMonitorListParams) ([]*ChannelMonitor, int64, error) {
	return nil, 0, nil
}
func (s *channelMonitorUsageRepoStub) ListEnabled(context.Context) ([]*ChannelMonitor, error) {
	return nil, nil
}
func (s *channelMonitorUsageRepoStub) MarkChecked(context.Context, int64, time.Time) error {
	return nil
}
func (s *channelMonitorUsageRepoStub) InsertHistoryBatch(context.Context, []*ChannelMonitorHistoryRow) error {
	return nil
}
func (s *channelMonitorUsageRepoStub) DeleteHistoryBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}
func (s *channelMonitorUsageRepoStub) ListHistory(context.Context, int64, string, int) ([]*ChannelMonitorHistoryEntry, error) {
	return nil, nil
}
func (s *channelMonitorUsageRepoStub) ListLatestPerModel(context.Context, int64) ([]*ChannelMonitorLatest, error) {
	return nil, nil
}
func (s *channelMonitorUsageRepoStub) ComputeAvailability(context.Context, int64, int) ([]*ChannelMonitorAvailability, error) {
	return nil, nil
}
func (s *channelMonitorUsageRepoStub) ListLatestForMonitorIDs(context.Context, []int64) (map[int64][]*ChannelMonitorLatest, error) {
	return nil, nil
}
func (s *channelMonitorUsageRepoStub) ComputeAvailabilityForMonitors(context.Context, []int64, int) (map[int64][]*ChannelMonitorAvailability, error) {
	return nil, nil
}
func (s *channelMonitorUsageRepoStub) ListLatestSuccessfulOpenAIUsageByModels(context.Context, []string) (map[string]*ChannelMonitorUsageLogLatest, error) {
	return s.latest, nil
}
func (s *channelMonitorUsageRepoStub) ListRecentHistoryForMonitors(context.Context, []int64, map[int64]string, int) (map[int64][]*ChannelMonitorHistoryEntry, error) {
	return nil, nil
}
func (s *channelMonitorUsageRepoStub) UpsertDailyRollupsFor(context.Context, time.Time) (int64, error) {
	return 0, nil
}
func (s *channelMonitorUsageRepoStub) DeleteRollupsBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}
func (s *channelMonitorUsageRepoStub) LoadAggregationWatermark(context.Context) (*time.Time, error) {
	return nil, nil
}
func (s *channelMonitorUsageRepoStub) UpdateAggregationWatermark(context.Context, time.Time) error {
	return nil
}

type channelMonitorPassEncryptor struct{}

func (channelMonitorPassEncryptor) Encrypt(s string) (string, error) { return s, nil }
func (channelMonitorPassEncryptor) Decrypt(s string) (string, error) { return s, nil }

func TestChannelMonitorRunCheckUsesUsageLogs(t *testing.T) {
	createdAt := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	durationMs := 120
	repo := &channelMonitorUsageRepoStub{
		monitor: &ChannelMonitor{
			ID:           9,
			Provider:     MonitorProviderOpenAI,
			APIKey:       "encrypted",
			PrimaryModel: "gpt-5.4",
			ExtraModels:  []string{"gpt-5.4-mini"},
		},
		latest: map[string]*ChannelMonitorUsageLogLatest{
			"gpt-5.4": {Model: "gpt-5.4", DurationMs: &durationMs, CreatedAt: createdAt},
		},
	}
	svc := NewChannelMonitorService(repo, channelMonitorPassEncryptor{})

	results, err := svc.RunCheck(context.Background(), 9)
	if err != nil {
		t.Fatalf("RunCheck returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Status != MonitorStatusOperational {
		t.Fatalf("primary status = %q, want %q", results[0].Status, MonitorStatusOperational)
	}
	if results[0].LatencyMs == nil || *results[0].LatencyMs != durationMs {
		t.Fatalf("primary latency = %v, want %d", results[0].LatencyMs, durationMs)
	}
	if !results[0].CheckedAt.Equal(createdAt) {
		t.Fatalf("primary checked_at = %s, want %s", results[0].CheckedAt, createdAt)
	}
	if results[1].Status != "" {
		t.Fatalf("missing usage log status = %q, want empty", results[1].Status)
	}
}

func TestUsageLogLatestToCheckResultDegraded(t *testing.T) {
	durationMs := int(monitorDegradedThreshold/time.Millisecond) + 1
	res := usageLogLatestToCheckResult("gpt-5.4", &ChannelMonitorUsageLogLatest{
		Model:      "gpt-5.4",
		DurationMs: &durationMs,
		CreatedAt:  time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC),
	}, time.Now())

	if res.Status != MonitorStatusDegraded {
		t.Fatalf("status = %q, want %q", res.Status, MonitorStatusDegraded)
	}
	if res.Message == "" {
		t.Fatalf("expected degraded message")
	}
}
