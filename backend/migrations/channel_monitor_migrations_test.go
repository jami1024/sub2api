package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration146IndexesOpsErrorLogsForChannelStatus(t *testing.T) {
	content, err := FS.ReadFile("146_add_channel_monitor_ops_error_log_indexes_notx.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ops_error_logs_channel_monitor_model_created_id")
	require.Contains(t, sql, "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ops_error_logs_channel_monitor_requested_model_created_id")
	require.Contains(t, sql, "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ops_error_logs_channel_monitor_upstream_model_created_id")
	require.Contains(t, sql, "ON ops_error_logs (model, created_at DESC, id DESC)")
	require.Contains(t, sql, "ON ops_error_logs (requested_model, created_at DESC, id DESC)")
	require.Contains(t, sql, "ON ops_error_logs (upstream_model, created_at DESC, id DESC)")
	require.Contains(t, sql, "COALESCE(status_code, 0) >= 400")
	require.Contains(t, sql, "NOT is_business_limited")
	require.Contains(t, sql, "is_count_tokens = FALSE")
	require.Contains(t, sql, "model IS NOT NULL")
	require.Contains(t, sql, "requested_model IS NOT NULL")
	require.Contains(t, sql, "upstream_model IS NOT NULL")
	require.NotContains(t, strings.ToUpper(sql), "BEGIN")
	require.NotContains(t, strings.ToUpper(sql), "COMMIT")
}
