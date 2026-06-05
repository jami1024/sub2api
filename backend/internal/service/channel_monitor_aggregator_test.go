package service

import (
	"testing"
	"time"
)

func TestTimelineEntriesForUserViewKeepsOpenAISLAEvents(t *testing.T) {
	now := time.Date(2026, 6, 5, 8, 30, 0, 0, time.UTC)
	m := &ChannelMonitor{Provider: MonitorProviderOpenAI}
	entries := []*ChannelMonitorHistoryEntry{
		{Status: MonitorStatusOperational, CheckedAt: now},
		{Status: MonitorStatusFailed, CheckedAt: now.Add(-time.Minute)},
		{Status: MonitorStatusError, CheckedAt: now.Add(-2 * time.Minute)},
		{Status: MonitorStatusDegraded, CheckedAt: now.Add(-3 * time.Minute)},
	}

	got := timelineEntriesForUserView(m, entries)

	if len(got) != len(entries) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(entries))
	}
	if got[1].Status != MonitorStatusFailed || got[2].Status != MonitorStatusError {
		t.Fatalf("statuses = %#v, want OpenAI SLA error events preserved", got)
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
