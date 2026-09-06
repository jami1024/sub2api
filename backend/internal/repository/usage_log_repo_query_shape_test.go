package repository

import (
	"strings"
	"testing"
)

type usageLogScanShapeRecorder struct {
	destinationCount int
}

func (r *usageLogScanShapeRecorder) Scan(destinations ...any) error {
	r.destinationCount = len(destinations)
	return nil
}

func TestUsageLogSelectColumnsMatchScanShape(t *testing.T) {
	recorder := &usageLogScanShapeRecorder{}
	if _, err := scanUsageLog(recorder); err != nil {
		t.Fatalf("scan usage log: %v", err)
	}

	columns := strings.Split(usageLogSelectColumns, ", ")
	if got, want := len(columns), recorder.destinationCount; got != want {
		t.Fatalf("usage log SELECT has %d columns, Scan has %d destinations", got, want)
	}

	selectList := ", " + usageLogSelectColumns + ", "
	for _, column := range []string{"upstream_latency_ms", "upstream_request_id"} {
		if !strings.Contains(selectList, ", "+column+", ") {
			t.Errorf("usage log SELECT is missing %s", column)
		}
	}
}
