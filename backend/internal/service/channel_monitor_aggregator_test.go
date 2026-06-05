package service

import (
	"testing"
	"time"
)

func TestTimelineEntriesForUserViewFiltersStaleOpenAIProbeFailures(t *testing.T) {
	now := time.Date(2026, 6, 5, 8, 30, 0, 0, time.UTC)
	m := &ChannelMonitor{Provider: MonitorProviderOpenAI}
	entries := []*ChannelMonitorHistoryEntry{
		{Status: MonitorStatusOperational, CheckedAt: now},
		{Status: MonitorStatusFailed, CheckedAt: now.Add(-time.Minute)},
		{Status: MonitorStatusError, CheckedAt: now.Add(-2 * time.Minute)},
		{Status: MonitorStatusDegraded, CheckedAt: now.Add(-3 * time.Minute)},
	}

	got := timelineEntriesForUserView(m, entries)

	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0].Status != MonitorStatusOperational || got[1].Status != MonitorStatusDegraded {
		t.Fatalf("statuses = [%q, %q], want operational/degraded only", got[0].Status, got[1].Status)
	}
}

func TestTimelineEntriesForUserViewKeepsNonOpenAIHistory(t *testing.T) {
	now := time.Date(2026, 6, 5, 8, 30, 0, 0, time.UTC)
	m := &ChannelMonitor{Provider: MonitorProviderAnthropic}
	entries := []*ChannelMonitorHistoryEntry{
		{Status: MonitorStatusOperational, CheckedAt: now},
		{Status: MonitorStatusFailed, CheckedAt: now.Add(-time.Minute)},
	}

	got := timelineEntriesForUserView(m, entries)

	if len(got) != len(entries) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(entries))
	}
}
