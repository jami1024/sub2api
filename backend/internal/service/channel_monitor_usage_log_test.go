package service

import (
	"context"
	"sync"
	"testing"
	"time"
)

type channelMonitorUsageRepoStub struct {
	monitor             *ChannelMonitor
	monitors            []*ChannelMonitor
	latest              map[string]*ChannelMonitorUsageLogLatest
	health              map[string]*ChannelMonitorUsageHealth
	healthResponses     []map[string]*ChannelMonitorUsageHealth
	events              map[string][]*ChannelMonitorHistoryEntry
	latestForMonitors   map[int64][]*ChannelMonitorLatest
	insertedRows        []*ChannelMonitorHistoryRow
	markedChecked       []time.Time
	latestPerModelN     int
	availabilityN       int
	latestForIDsN       int
	availabilityForIDsN int
	latestUsageN        int
	usageHealthN        int
	usageEventsN        int
	healthSince         []time.Time
	eventsSince         []time.Time
	eventsLimit         []int
	latestUsageBlock    chan struct{}
	mu                  sync.Mutex
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
	if s.monitors != nil {
		return s.monitors, nil
	}
	return nil, nil
}
func (s *channelMonitorUsageRepoStub) MarkChecked(_ context.Context, _ int64, checkedAt time.Time) error {
	s.markedChecked = append(s.markedChecked, checkedAt)
	return nil
}
func (s *channelMonitorUsageRepoStub) InsertHistoryBatch(_ context.Context, rows []*ChannelMonitorHistoryRow) error {
	s.insertedRows = append(s.insertedRows, rows...)
	return nil
}
func (s *channelMonitorUsageRepoStub) DeleteHistoryBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}
func (s *channelMonitorUsageRepoStub) ListHistory(context.Context, int64, string, int) ([]*ChannelMonitorHistoryEntry, error) {
	return nil, nil
}
func (s *channelMonitorUsageRepoStub) ListLatestPerModel(context.Context, int64) ([]*ChannelMonitorLatest, error) {
	s.latestPerModelN++
	return nil, nil
}
func (s *channelMonitorUsageRepoStub) ComputeAvailability(context.Context, int64, int) ([]*ChannelMonitorAvailability, error) {
	s.availabilityN++
	return nil, nil
}
func (s *channelMonitorUsageRepoStub) ListLatestForMonitorIDs(context.Context, []int64) (map[int64][]*ChannelMonitorLatest, error) {
	s.latestForIDsN++
	return s.latestForMonitors, nil
}
func (s *channelMonitorUsageRepoStub) ComputeAvailabilityForMonitors(context.Context, []int64, int) (map[int64][]*ChannelMonitorAvailability, error) {
	s.availabilityForIDsN++
	return nil, nil
}
func (s *channelMonitorUsageRepoStub) ListLatestSuccessfulOpenAIUsageByModels(context.Context, []string, time.Time) (map[string]*ChannelMonitorUsageLogLatest, error) {
	s.mu.Lock()
	s.latestUsageN++
	s.mu.Unlock()
	if s.latestUsageBlock != nil {
		<-s.latestUsageBlock
	}
	return s.latest, nil
}
func (s *channelMonitorUsageRepoStub) ComputeOpenAIUsageHealthByModels(_ context.Context, _ []string, since time.Time) (map[string]*ChannelMonitorUsageHealth, error) {
	s.mu.Lock()
	s.usageHealthN++
	n := s.usageHealthN
	s.healthSince = append(s.healthSince, since)
	s.mu.Unlock()
	if s.latestUsageBlock != nil {
		<-s.latestUsageBlock
	}
	if len(s.healthResponses) >= n {
		return s.healthResponses[n-1], nil
	}
	return s.health, nil
}
func (s *channelMonitorUsageRepoStub) ListRecentOpenAIUsageEventsByModels(_ context.Context, _ []string, since time.Time, limit int) (map[string][]*ChannelMonitorHistoryEntry, error) {
	s.usageEventsN++
	s.eventsSince = append(s.eventsSince, since)
	s.eventsLimit = append(s.eventsLimit, limit)
	return s.events, nil
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
	durationMs := 1200
	firstTokenMs := 120
	avgFirstTokenMs := 120
	repo := &channelMonitorUsageRepoStub{
		monitor: &ChannelMonitor{
			ID:           9,
			Provider:     MonitorProviderOpenAI,
			APIKey:       "encrypted",
			PrimaryModel: "gpt-5.4",
			ExtraModels:  []string{"gpt-5.4-mini"},
		},
		latest: map[string]*ChannelMonitorUsageLogLatest{
			"gpt-5.4": {Model: "gpt-5.4", DurationMs: &durationMs, FirstTokenMs: &firstTokenMs, AvgFirstTokenMs: &avgFirstTokenMs, CreatedAt: createdAt},
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
	if results[0].LatencyMs == nil || *results[0].LatencyMs != firstTokenMs {
		t.Fatalf("primary latency = %v, want first token %d", results[0].LatencyMs, firstTokenMs)
	}
	if !results[0].CheckedAt.Equal(createdAt) {
		t.Fatalf("primary checked_at = %s, want %s", results[0].CheckedAt, createdAt)
	}
	if results[1].Status != "" {
		t.Fatalf("missing usage log status = %q, want empty", results[1].Status)
	}
}

func TestChannelMonitorRunCheckPersistsHistoryOnlyForUsageBackedStatuses(t *testing.T) {
	createdAt := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	durationMs := 1200
	firstTokenMs := 120
	avgFirstTokenMs := 120
	repo := &channelMonitorUsageRepoStub{
		monitor: &ChannelMonitor{
			ID:           9,
			Provider:     MonitorProviderOpenAI,
			APIKey:       "encrypted",
			PrimaryModel: "gpt-5.4",
			ExtraModels:  []string{"gpt-5.4-mini"},
		},
		latest: map[string]*ChannelMonitorUsageLogLatest{
			"gpt-5.4": {Model: "gpt-5.4", DurationMs: &durationMs, FirstTokenMs: &firstTokenMs, AvgFirstTokenMs: &avgFirstTokenMs, CreatedAt: createdAt},
		},
	}
	svc := NewChannelMonitorService(repo, channelMonitorPassEncryptor{})

	_, err := svc.RunCheck(context.Background(), 9)
	if err != nil {
		t.Fatalf("RunCheck returned error: %v", err)
	}
	if len(repo.insertedRows) != 1 {
		t.Fatalf("len(insertedRows) = %d, want 1", len(repo.insertedRows))
	}
	row := repo.insertedRows[0]
	if row.Model != "gpt-5.4" || row.Status != MonitorStatusOperational || !row.CheckedAt.Equal(createdAt) {
		t.Fatalf("inserted row = %#v, want gpt-5.4 operational at usage log time", row)
	}
	if len(repo.markedChecked) != 1 {
		t.Fatalf("len(markedChecked) = %d, want 1", len(repo.markedChecked))
	}
}

func TestChannelMonitorRunCheckSkipsDuplicateUsageHistory(t *testing.T) {
	createdAt := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	durationMs := 1200
	firstTokenMs := 120
	avgFirstTokenMs := 120
	repo := &channelMonitorUsageRepoStub{
		monitor: &ChannelMonitor{
			ID:           9,
			Provider:     MonitorProviderOpenAI,
			APIKey:       "encrypted",
			PrimaryModel: "gpt-5.4",
		},
		latest: map[string]*ChannelMonitorUsageLogLatest{
			"gpt-5.4": {Model: "gpt-5.4", DurationMs: &durationMs, FirstTokenMs: &firstTokenMs, AvgFirstTokenMs: &avgFirstTokenMs, CreatedAt: createdAt},
		},
		latestForMonitors: map[int64][]*ChannelMonitorLatest{
			9: {{Model: "gpt-5.4", Status: MonitorStatusOperational, CheckedAt: createdAt}},
		},
	}
	svc := NewChannelMonitorService(repo, channelMonitorPassEncryptor{})

	_, err := svc.RunCheck(context.Background(), 9)
	if err != nil {
		t.Fatalf("RunCheck returned error: %v", err)
	}
	if len(repo.insertedRows) != 0 {
		t.Fatalf("len(insertedRows) = %d, want 0 for duplicate usage log", len(repo.insertedRows))
	}
	if len(repo.markedChecked) != 0 {
		t.Fatalf("len(markedChecked) = %d, want 0 for duplicate usage log", len(repo.markedChecked))
	}
}

func TestChannelMonitorRunCheckDoesNotPersistWhenNoUsageLog(t *testing.T) {
	repo := &channelMonitorUsageRepoStub{
		monitor: &ChannelMonitor{
			ID:           9,
			Provider:     MonitorProviderOpenAI,
			APIKey:       "encrypted",
			PrimaryModel: "gpt-5.4",
		},
		latest: map[string]*ChannelMonitorUsageLogLatest{},
	}
	svc := NewChannelMonitorService(repo, channelMonitorPassEncryptor{})

	_, err := svc.RunCheck(context.Background(), 9)
	if err != nil {
		t.Fatalf("RunCheck returned error: %v", err)
	}
	if len(repo.insertedRows) != 0 {
		t.Fatalf("len(insertedRows) = %d, want 0 without usage log", len(repo.insertedRows))
	}
	if len(repo.markedChecked) != 0 {
		t.Fatalf("len(markedChecked) = %d, want 0 without usage log", len(repo.markedChecked))
	}
	if repo.latestForIDsN != 0 {
		t.Fatalf("latest history queries = %d, want 0 without usage-backed rows", repo.latestForIDsN)
	}
}

func TestChannelMonitorBatchSummaryUsesOpenAIUsageHealthForAvailability(t *testing.T) {
	firstTokenMs := 900
	avgFirstTokenMs := 450
	repo := &channelMonitorUsageRepoStub{
		latest: map[string]*ChannelMonitorUsageLogLatest{
			"gpt-5.4": {
				Model:           "gpt-5.4",
				FirstTokenMs:    &firstTokenMs,
				AvgFirstTokenMs: &avgFirstTokenMs,
				CreatedAt:       time.Now(),
			},
		},
		health: map[string]*ChannelMonitorUsageHealth{
			"gpt-5.4": {
				Model:           "gpt-5.4",
				SuccessCount:    172,
				ErrorCountSLA:   45,
				AvailabilityPct: 79.2627,
				LatestStatus:    MonitorStatusOperational,
				LatestLatencyMs: &avgFirstTokenMs,
				LatestCheckedAt: time.Now(),
			},
		},
	}
	svc := NewChannelMonitorService(repo, channelMonitorPassEncryptor{})

	got := svc.BatchMonitorStatusSummary(
		context.Background(),
		[]int64{9},
		map[int64]string{9: MonitorProviderOpenAI},
		map[int64]string{9: "gpt-5.4"},
		map[int64][]string{9: []string{"gpt-5.4-mini"}},
		map[int64]int{9: monitorMinIntervalSeconds},
	)

	if got[9].PrimaryStatus != MonitorStatusOperational {
		t.Fatalf("primary status = %q, want %q", got[9].PrimaryStatus, MonitorStatusOperational)
	}
	if got[9].Availability7d != 79.2627 {
		t.Fatalf("availability = %.4f, want SLA-based 79.2627", got[9].Availability7d)
	}
	if got[9].PrimaryLatencyMs == nil || *got[9].PrimaryLatencyMs != avgFirstTokenMs {
		t.Fatalf("primary latency = %v, want %d", got[9].PrimaryLatencyMs, avgFirstTokenMs)
	}
	if repo.usageHealthN != 1 {
		t.Fatalf("usage health queries = %d, want 1", repo.usageHealthN)
	}
	if repo.latestForIDsN != 0 || repo.availabilityForIDsN != 0 || repo.latestUsageN != 0 {
		t.Fatalf("legacy queries latest=%d availability=%d successfulUsage=%d, want all 0 for OpenAI summary", repo.latestForIDsN, repo.availabilityForIDsN, repo.latestUsageN)
	}
	if len(repo.healthSince) != 1 {
		t.Fatalf("health since calls = %d, want 1", len(repo.healthSince))
	}
	age := time.Since(repo.healthSince[0])
	if age < 29*time.Minute || age > 31*time.Minute {
		t.Fatalf("health window age = %s, want about 30m", age)
	}
}

func TestChannelMonitorListUserViewCachesOpenAISummaryAndTimeline(t *testing.T) {
	now := time.Now().UTC()
	latency := 500
	repo := &channelMonitorUsageRepoStub{
		monitors: []*ChannelMonitor{{
			ID:              9,
			Provider:        MonitorProviderOpenAI,
			Enabled:         true,
			Name:            "OpenAI",
			PrimaryModel:    "gpt-5.5",
			IntervalSeconds: monitorMinIntervalSeconds,
		}},
		health: map[string]*ChannelMonitorUsageHealth{
			"gpt-5.5": {Model: "gpt-5.5", SuccessCount: 10, AvailabilityPct: 100, LatestStatus: MonitorStatusOperational, LatestLatencyMs: &latency, LatestCheckedAt: now},
		},
		events: map[string][]*ChannelMonitorHistoryEntry{
			"gpt-5.5": {{Model: "gpt-5.5", Status: MonitorStatusOperational, LatencyMs: &latency, CheckedAt: now}},
		},
	}
	svc := NewChannelMonitorService(repo, channelMonitorPassEncryptor{})

	if _, err := svc.ListUserView(context.Background()); err != nil {
		t.Fatalf("first ListUserView returned error: %v", err)
	}
	if _, err := svc.ListUserView(context.Background()); err != nil {
		t.Fatalf("second ListUserView returned error: %v", err)
	}

	if repo.usageHealthN != 1 {
		t.Fatalf("usage health queries = %d, want 1 due to list cache", repo.usageHealthN)
	}
	if repo.usageEventsN != 1 {
		t.Fatalf("usage event queries = %d, want 1 due to list cache", repo.usageEventsN)
	}
}

func TestChannelMonitorGetUserDetailUsesOpenAIUsageHealthWindows(t *testing.T) {
	createdAt := time.Now().UTC()
	avgFirstTokenMs := 450
	repo := &channelMonitorUsageRepoStub{
		monitor: &ChannelMonitor{
			ID:              9,
			Provider:        MonitorProviderOpenAI,
			Enabled:         true,
			Name:            "OpenAI",
			PrimaryModel:    "gpt-5.4",
			IntervalSeconds: monitorMinIntervalSeconds,
		},
		healthResponses: []map[string]*ChannelMonitorUsageHealth{
			{
				"gpt-5.4": {Model: "gpt-5.4", SuccessCount: 172, ErrorCountSLA: 45, AvailabilityPct: 79.2627, LatestStatus: MonitorStatusOperational, LatestLatencyMs: &avgFirstTokenMs, LatestCheckedAt: createdAt, AvgLatencyMs: &avgFirstTokenMs},
			},
			{
				"gpt-5.4": {Model: "gpt-5.4", SuccessCount: 300, ErrorCountSLA: 50, AvailabilityPct: 85.7143, LatestStatus: MonitorStatusOperational, LatestLatencyMs: &avgFirstTokenMs, LatestCheckedAt: createdAt, AvgLatencyMs: &avgFirstTokenMs},
			},
			{
				"gpt-5.4": {Model: "gpt-5.4", SuccessCount: 700, ErrorCountSLA: 100, AvailabilityPct: 87.5, LatestStatus: MonitorStatusOperational, LatestLatencyMs: &avgFirstTokenMs, LatestCheckedAt: createdAt, AvgLatencyMs: &avgFirstTokenMs},
			},
		},
		latestForMonitors: map[int64][]*ChannelMonitorLatest{
			9: {{Model: "gpt-5.4", Status: MonitorStatusFailed, CheckedAt: createdAt.Add(-time.Hour)}},
		},
	}
	svc := NewChannelMonitorService(repo, channelMonitorPassEncryptor{})

	detail, err := svc.GetUserDetail(context.Background(), 9)
	if err != nil {
		t.Fatalf("GetUserDetail returned error: %v", err)
	}
	if repo.latestPerModelN != 0 || repo.availabilityN != 0 {
		t.Fatalf("legacy history calls latest=%d availability=%d, want 0 for OpenAI usage-log detail", repo.latestPerModelN, repo.availabilityN)
	}
	if len(detail.Models) != 1 {
		t.Fatalf("len(models) = %d, want 1", len(detail.Models))
	}
	model := detail.Models[0]
	if model.LatestStatus != MonitorStatusOperational {
		t.Fatalf("latest status = %q, want %q", model.LatestStatus, MonitorStatusOperational)
	}
	if model.LatestLatencyMs == nil || *model.LatestLatencyMs != avgFirstTokenMs {
		t.Fatalf("latest latency = %v, want average first token %d", model.LatestLatencyMs, avgFirstTokenMs)
	}
	if model.Availability7d != 79.2627 || model.Availability15d != 85.7143 || model.Availability30d != 87.5 {
		t.Fatalf("availability = %.4f/%.4f/%.4f, want SLA health windows", model.Availability7d, model.Availability15d, model.Availability30d)
	}
	if model.AvgLatency7dMs == nil || *model.AvgLatency7dMs != avgFirstTokenMs {
		t.Fatalf("avg latency = %v, want average first token %d", model.AvgLatency7dMs, avgFirstTokenMs)
	}
}

func TestChannelMonitorListUserViewUsesOpenAIUsageEventsForTimeline(t *testing.T) {
	now := time.Now().UTC()
	latency := 7956
	repo := &channelMonitorUsageRepoStub{
		monitors: []*ChannelMonitor{{
			ID:              9,
			Provider:        MonitorProviderOpenAI,
			Enabled:         true,
			Name:            "OpenAI",
			PrimaryModel:    "gpt-5.5",
			IntervalSeconds: monitorMinIntervalSeconds,
		}},
		health: map[string]*ChannelMonitorUsageHealth{
			"gpt-5.5": {Model: "gpt-5.5", SuccessCount: 172, ErrorCountSLA: 45, AvailabilityPct: 79.2627, LatestStatus: MonitorStatusOperational, LatestLatencyMs: &latency, LatestCheckedAt: now},
		},
		events: map[string][]*ChannelMonitorHistoryEntry{
			"gpt-5.5": {
				{Model: "gpt-5.5", Status: MonitorStatusOperational, LatencyMs: &latency, CheckedAt: now},
				{Model: "gpt-5.5", Status: MonitorStatusFailed, CheckedAt: now.Add(-time.Minute)},
			},
		},
	}
	svc := NewChannelMonitorService(repo, channelMonitorPassEncryptor{})

	views, err := svc.ListUserView(context.Background())
	if err != nil {
		t.Fatalf("ListUserView returned error: %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("len(views) = %d, want 1", len(views))
	}
	if views[0].Availability7d != 79.2627 {
		t.Fatalf("availability = %.4f, want SLA-based 79.2627", views[0].Availability7d)
	}
	if len(views[0].Timeline) != 2 {
		t.Fatalf("timeline len = %d, want 2", len(views[0].Timeline))
	}
	if views[0].Timeline[0].Status != MonitorStatusOperational || views[0].Timeline[1].Status != MonitorStatusFailed {
		t.Fatalf("timeline statuses = %#v, want success then failed", views[0].Timeline)
	}
	if len(repo.eventsSince) != 1 {
		t.Fatalf("event since calls = %d, want 1", len(repo.eventsSince))
	}
	age := time.Since(repo.eventsSince[0])
	if age < 29*time.Minute || age > 31*time.Minute {
		t.Fatalf("event window age = %s, want about 30m", age)
	}
	if len(repo.eventsLimit) != 1 || repo.eventsLimit[0] != monitorTimelineMaxPoints {
		t.Fatalf("event limit = %#v, want [%d]", repo.eventsLimit, monitorTimelineMaxPoints)
	}
}

func TestChannelMonitorGetUserDetailCachesOpenAIUsageHealth(t *testing.T) {
	createdAt := time.Now().UTC()
	avgFirstTokenMs := 450
	repo := &channelMonitorUsageRepoStub{
		monitor: &ChannelMonitor{
			ID:              9,
			Provider:        MonitorProviderOpenAI,
			Enabled:         true,
			Name:            "OpenAI",
			PrimaryModel:    "gpt-5.4",
			IntervalSeconds: monitorMinIntervalSeconds,
		},
		health: map[string]*ChannelMonitorUsageHealth{
			"gpt-5.4": {
				Model:           "gpt-5.4",
				SuccessCount:    10,
				AvailabilityPct: 100,
				LatestStatus:    MonitorStatusOperational,
				LatestLatencyMs: &avgFirstTokenMs,
				LatestCheckedAt: createdAt,
				AvgLatencyMs:    &avgFirstTokenMs,
			},
		},
	}
	svc := NewChannelMonitorService(repo, channelMonitorPassEncryptor{})

	for i := 0; i < 2; i++ {
		if _, err := svc.GetUserDetail(context.Background(), 9); err != nil {
			t.Fatalf("GetUserDetail #%d returned error: %v", i+1, err)
		}
	}

	if repo.usageHealthN != 3 {
		t.Fatalf("usage health detail queries = %d, want 3 within cache TTL", repo.usageHealthN)
	}
}

func TestChannelMonitorGetUserDetailCoalescesConcurrentOpenAIUsageHealthLoads(t *testing.T) {
	createdAt := time.Now().UTC()
	avgFirstTokenMs := 450
	block := make(chan struct{})
	repo := &channelMonitorUsageRepoStub{
		monitor: &ChannelMonitor{
			ID:              9,
			Provider:        MonitorProviderOpenAI,
			Enabled:         true,
			Name:            "OpenAI",
			PrimaryModel:    "gpt-5.4",
			IntervalSeconds: monitorMinIntervalSeconds,
		},
		health: map[string]*ChannelMonitorUsageHealth{
			"gpt-5.4": {
				Model:           "gpt-5.4",
				SuccessCount:    10,
				AvailabilityPct: 100,
				LatestStatus:    MonitorStatusOperational,
				LatestLatencyMs: &avgFirstTokenMs,
				LatestCheckedAt: createdAt,
				AvgLatencyMs:    &avgFirstTokenMs,
			},
		},
		latestUsageBlock: block,
	}
	svc := NewChannelMonitorService(repo, channelMonitorPassEncryptor{})

	const callers = 8
	var wg sync.WaitGroup
	errs := make(chan error, callers)
	started := make(chan struct{})
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-started
			_, err := svc.GetUserDetail(context.Background(), 9)
			errs <- err
		}()
	}
	close(started)
	for {
		repo.mu.Lock()
		n := repo.usageHealthN
		repo.mu.Unlock()
		if n > 0 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	close(block)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("GetUserDetail returned error: %v", err)
		}
	}

	if repo.usageHealthN != 3 {
		t.Fatalf("usage health detail queries = %d, want 3 for concurrent callers", repo.usageHealthN)
	}
}

func TestUsageLogLatestToCheckResultDegraded(t *testing.T) {
	firstTokenMs := int(monitorDegradedThreshold/time.Millisecond) + 1
	res := usageLogLatestToCheckResult("gpt-5.4", &ChannelMonitorUsageLogLatest{
		Model:           "gpt-5.4",
		FirstTokenMs:    &firstTokenMs,
		AvgFirstTokenMs: &firstTokenMs,
		CreatedAt:       time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC),
	}, time.Now())

	if res.Status != MonitorStatusDegraded {
		t.Fatalf("status = %q, want %q", res.Status, MonitorStatusDegraded)
	}
	if res.Message == "" {
		t.Fatalf("expected degraded message")
	}
}

func TestUsageLogLatestToCheckResultUsesAverageFirstTokenLatency(t *testing.T) {
	firstTokenMs := 900
	avgFirstTokenMs := 450
	res := usageLogLatestToCheckResult("gpt-5.4", &ChannelMonitorUsageLogLatest{
		Model:           "gpt-5.4",
		FirstTokenMs:    &firstTokenMs,
		AvgFirstTokenMs: &avgFirstTokenMs,
		CreatedAt:       time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC),
	}, time.Now())

	if res.LatencyMs == nil || *res.LatencyMs != avgFirstTokenMs {
		t.Fatalf("latency = %v, want average first token %d", res.LatencyMs, avgFirstTokenMs)
	}
	if res.Status != MonitorStatusOperational {
		t.Fatalf("status = %q, want %q", res.Status, MonitorStatusOperational)
	}
}

func TestUsageLogLatestToCheckResultDoesNotUseTotalDurationAsLatency(t *testing.T) {
	durationMs := int(monitorDegradedThreshold/time.Millisecond) + 1
	res := usageLogLatestToCheckResult("gpt-5.4", &ChannelMonitorUsageLogLatest{
		Model:      "gpt-5.4",
		DurationMs: &durationMs,
		CreatedAt:  time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC),
	}, time.Now())

	if res.Status != MonitorStatusOperational {
		t.Fatalf("status = %q, want %q when only total duration is slow", res.Status, MonitorStatusOperational)
	}
	if res.LatencyMs != nil {
		t.Fatalf("latency = %v, want nil when first_token_ms is missing", *res.LatencyMs)
	}
}
