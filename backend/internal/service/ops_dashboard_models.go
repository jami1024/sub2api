package service

import "time"

type OpsDashboardFilter struct {
	StartTime time.Time
	EndTime   time.Time

	Platform string
	GroupID  *int64

	// QueryMode controls whether dashboard queries should use raw logs or pre-aggregated tables.
	// Expected values: auto/raw/preagg (see OpsQueryMode).
	QueryMode OpsQueryMode
}

type OpsRateSummary struct {
	Current float64 `json:"current"`
	Peak    float64 `json:"peak"`
	Avg     float64 `json:"avg"`
}

type OpsPercentiles struct {
	P50 *int `json:"p50_ms"`
	P90 *int `json:"p90_ms"`
	P95 *int `json:"p95_ms"`
	P99 *int `json:"p99_ms"`
	Avg *int `json:"avg_ms"`
	Max *int `json:"max_ms"`
}

type OpsDashboardOverview struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Platform  string    `json:"platform"`
	GroupID   *int64    `json:"group_id"`

	// HealthScore is a backend-computed overall health score (0-100).
	// It is derived from the monitored metrics in this overview, plus best-effort system metrics/job heartbeats.
	HealthScore int `json:"health_score"`

	// Latest system-level snapshot (window=1m, global).
	SystemMetrics *OpsSystemMetricsSnapshot `json:"system_metrics"`

	// Background jobs health (heartbeats).
	JobHeartbeats []*OpsJobHeartbeat `json:"job_heartbeats"`

	SuccessCount         int64 `json:"success_count"`
	ErrorCountTotal      int64 `json:"error_count_total"`
	BusinessLimitedCount int64 `json:"business_limited_count"`

	ErrorCountSLA     int64 `json:"error_count_sla"`
	RequestCountTotal int64 `json:"request_count_total"`
	RequestCountSLA   int64 `json:"request_count_sla"`

	TokenConsumed int64 `json:"token_consumed"`

	SLA                          float64 `json:"sla"`
	ErrorRate                    float64 `json:"error_rate"`
	UpstreamErrorRate            float64 `json:"upstream_error_rate"`
	UpstreamErrorCountExcl429529 int64   `json:"upstream_error_count_excl_429_529"`
	Upstream429Count             int64   `json:"upstream_429_count"`
	Upstream529Count             int64   `json:"upstream_529_count"`

	QPS OpsRateSummary `json:"qps"`
	TPS OpsRateSummary `json:"tps"`

	Duration OpsPercentiles `json:"duration"`
	TTFT     OpsPercentiles `json:"ttft"`
}

type OpsLatencyHistogramBucket struct {
	Range string `json:"range"`
	Count int64  `json:"count"`
}

// OpsLatencyHistogramResponse is a coarse latency distribution histogram (success requests only).
// It is used by the Ops dashboard to quickly identify tail latency regressions.
type OpsLatencyHistogramResponse struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Platform  string    `json:"platform"`
	GroupID   *int64    `json:"group_id"`

	TotalRequests int64                        `json:"total_requests"`
	Buckets       []*OpsLatencyHistogramBucket `json:"buckets"`
}

type OpsProviderStatusFilter struct {
	StartTime     time.Time
	EndTime       time.Time
	BucketSeconds int
	Limit         int
}

type OpsProviderStatusSummaryItem struct {
	Provider             string                            `json:"provider"`
	RequestCount         int64                             `json:"request_count"`
	SuccessCount         int64                             `json:"success_count"`
	FailureCount         int64                             `json:"failure_count"`
	BusinessLimitedCount int64                             `json:"business_limited_count"`
	Availability         float64                           `json:"availability"`
	ErrorRate            float64                           `json:"error_rate"`
	CacheReadRate        *float64                          `json:"cache_read_rate,omitempty"`
	P50Ms                *int                              `json:"p50_ms"`
	P95Ms                *int                              `json:"p95_ms"`
	P99Ms                *int                              `json:"p99_ms"`
	DurationAvgMs        *int                              `json:"duration_avg_ms"`
	DurationMaxMs        *int                              `json:"duration_max_ms"`
	TTFTAvgMs            *int                              `json:"ttft_avg_ms"`
	TTFTP95Ms            *int                              `json:"ttft_p95_ms"`
	TTFTSampleCount      int64                             `json:"ttft_sample_count"`
	UpstreamTTFTAvgMs    *int                              `json:"upstream_ttft_avg_ms,omitempty"`
	GatewayTTFTAvgMs     *int                              `json:"gateway_ttft_avg_ms,omitempty"`
	Timeout524Count      int64                             `json:"timeout_524_count"`
	Timeout524AvgMs      *int                              `json:"timeout_524_avg_ms"`
	LastSeen             *time.Time                        `json:"last_seen"`
	Fingerprint          *OpsProviderStatusFingerprint     `json:"fingerprint,omitempty"`
	Timeline             []*OpsProviderStatusTimelinePoint `json:"timeline,omitempty"`
}

type OpsProviderStatusFingerprint struct {
	Headers  map[string]string `json:"headers,omitempty"`
	LastSeen *time.Time        `json:"last_seen,omitempty"`
}

type OpsProviderStatusTimelinePoint struct {
	Provider          string    `json:"provider,omitempty"`
	BucketStart       time.Time `json:"bucket_start"`
	RequestCount      int64     `json:"request_count"`
	SuccessCount      int64     `json:"success_count"`
	FailureCount      int64     `json:"failure_count"`
	Availability      float64   `json:"availability"`
	P50Ms             *int      `json:"p50_ms"`
	P95Ms             *int      `json:"p95_ms"`
	P99Ms             *int      `json:"p99_ms"`
	DurationAvgMs     *int      `json:"duration_avg_ms"`
	TTFTAvgMs         *int      `json:"ttft_avg_ms"`
	TTFTSampleCount   int64     `json:"ttft_sample_count"`
	UpstreamTTFTAvgMs *int      `json:"upstream_ttft_avg_ms,omitempty"`
	GatewayTTFTAvgMs  *int      `json:"gateway_ttft_avg_ms,omitempty"`
	Timeout524Count   int64     `json:"timeout_524_count"`
	Timeout524AvgMs   *int      `json:"timeout_524_avg_ms"`
}

type OpsProviderStatusResponse struct {
	StartTime     time.Time                         `json:"start_time"`
	EndTime       time.Time                         `json:"end_time"`
	BucketSeconds int                               `json:"bucket_seconds"`
	Items         []*OpsProviderStatusSummaryItem   `json:"items"`
	Timeline      []*OpsProviderStatusTimelinePoint `json:"timeline"`
}

type OpsClientFailureStatsFilter struct {
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

type OpsClientFailureStatsItem struct {
	UserID             *int64     `json:"user_id,omitempty"`
	UserEmail          string     `json:"user_email"`
	FailureCount       int64      `json:"failure_count"`
	AffectedKeyCount   int64      `json:"affected_key_count"`
	TopErrorMessage    string     `json:"top_error_message"`
	TopErrorCount      int64      `json:"top_error_count"`
	LastSeen           *time.Time `json:"last_seen,omitempty"`
	TopInboundEndpoint string     `json:"top_inbound_endpoint"`
	TopPlatform        string     `json:"top_platform"`
}

type OpsClientFailureStatsResponse struct {
	StartTime time.Time                    `json:"start_time"`
	EndTime   time.Time                    `json:"end_time"`
	Items     []*OpsClientFailureStatsItem `json:"items"`
}

const (
	OpsUpstreamMultiplierStatusSuccess = "success"
	OpsUpstreamMultiplierStatusSkipped = "skipped"
	OpsUpstreamMultiplierStatusError   = "error"
)

type OpsUpstreamMultiplierAccount struct {
	AccountID             int64                        `json:"account_id"`
	AccountName           string                       `json:"account_name"`
	Platform              string                       `json:"platform"`
	BaseURL               string                       `json:"base_url"`
	KeyPrefix             string                       `json:"key_prefix"`
	AccountRateMultiplier float64                      `json:"account_rate_multiplier"`
	Supported             bool                         `json:"supported"`
	SkipReason            string                       `json:"skip_reason,omitempty"`
	LatestSample          *OpsUpstreamMultiplierSample `json:"latest_sample,omitempty"`
}

type OpsUpstreamMultiplierSample struct {
	ID                  int64     `json:"id"`
	AccountID           int64     `json:"account_id"`
	AccountNameSnapshot string    `json:"account_name_snapshot"`
	Platform            string    `json:"platform"`
	BaseURLSnapshot     string    `json:"base_url_snapshot"`
	KeyPrefixSnapshot   string    `json:"key_prefix_snapshot"`
	Model               string    `json:"model"`
	Status              string    `json:"status"`
	HTTPStatus          *int      `json:"http_status,omitempty"`
	StandardCostDelta   *float64  `json:"standard_cost_delta,omitempty"`
	ActualCostDelta     *float64  `json:"actual_cost_delta,omitempty"`
	Multiplier          *float64  `json:"multiplier,omitempty"`
	BalanceBefore       *float64  `json:"balance_before,omitempty"`
	BalanceAfter        *float64  `json:"balance_after,omitempty"`
	ErrorMessage        string    `json:"error_message,omitempty"`
	MeasuredAt          time.Time `json:"measured_at"`
	CreatedAt           time.Time `json:"created_at"`
}

type OpsUpstreamMultiplierSamplesFilter struct {
	Model     string
	AccountID *int64
	Limit     int
}

type OpsUpstreamMultiplierAccountsResponse struct {
	Model    string                          `json:"model"`
	Accounts []*OpsUpstreamMultiplierAccount `json:"accounts"`
}

type OpsUpstreamMultiplierSamplesResponse struct {
	Model   string                         `json:"model"`
	Samples []*OpsUpstreamMultiplierSample `json:"samples"`
}

type OpsMeasureUpstreamMultiplierRequest struct {
	Model      string  `json:"model"`
	AccountIDs []int64 `json:"account_ids"`
}

type OpsMeasureUpstreamMultiplierResponse struct {
	Model   string                         `json:"model"`
	Samples []*OpsUpstreamMultiplierSample `json:"samples"`
}

type OpsApplyUpstreamMultiplierRequest struct {
	Model     string `json:"model"`
	AccountID int64  `json:"account_id"`
}

type OpsApplyUpstreamMultiplierResponse struct {
	Model          string                       `json:"model"`
	AccountID      int64                        `json:"account_id"`
	RateMultiplier float64                      `json:"rate_multiplier"`
	Sample         *OpsUpstreamMultiplierSample `json:"sample"`
}

const (
	OpsGroupRateRecommendationStatusSafe         = "safe"
	OpsGroupRateRecommendationStatusBasicSafe    = "basic_safe"
	OpsGroupRateRecommendationStatusLow          = "low"
	OpsGroupRateRecommendationStatusInsufficient = "insufficient_data"
)

type OpsGroupRateRecommendationFilter struct {
	Model                string  `json:"model"`
	PackageScope         string  `json:"package_scope"`
	ProfitMargin         float64 `json:"profit_margin"`
	SafetyFactor         float64 `json:"safety_factor"`
	UsageDays            int     `json:"usage_days"`
	IncludeUnschedulable bool    `json:"include_unschedulable"`
	IncludeSelfHosted    bool    `json:"include_self_hosted"`
}

type OpsGroupRateRecommendationPackageBasis struct {
	PackageID        int64   `json:"package_id"`
	Name             string  `json:"name"`
	Price            float64 `json:"price"`
	CreditAmount     float64 `json:"credit_amount"`
	PackageScope     string  `json:"package_scope"`
	RevenuePerCredit float64 `json:"revenue_per_credit"`
}

type OpsGroupRateRecommendationUsageShare struct {
	RequestCount      int64   `json:"request_count"`
	RequestShare      float64 `json:"request_share"`
	StandardCost      float64 `json:"standard_cost"`
	StandardCostShare float64 `json:"standard_cost_share"`
}

type OpsGroupRateRecommendationAccount struct {
	AccountID            int64      `json:"account_id"`
	AccountName          string     `json:"account_name"`
	BaseURL              string     `json:"base_url"`
	KeyPrefix            string     `json:"key_prefix"`
	Schedulable          bool       `json:"schedulable"`
	Status               string     `json:"status"`
	CurrentPriority      int        `json:"current_priority"`
	BindingPriority      int        `json:"binding_priority"`
	UpstreamMultiplier   *float64   `json:"upstream_multiplier,omitempty"`
	MultiplierStatus     string     `json:"multiplier_status"`
	MultiplierMeasuredAt *time.Time `json:"multiplier_measured_at,omitempty"`
	RequestCount         int64      `json:"request_count"`
	RequestShare         float64    `json:"request_share"`
	StandardCost         float64    `json:"standard_cost"`
	StandardCostShare    float64    `json:"standard_cost_share"`
	RecommendedWeight    float64    `json:"recommended_weight"`
	RecommendedPriority  int        `json:"recommended_priority"`
	ParticipatesInAdvice bool       `json:"participates_in_advice"`
	Note                 string     `json:"note"`
}

type OpsGroupRateRecommendationGroup struct {
	GroupID                      int64                                `json:"group_id"`
	GroupName                    string                               `json:"group_name"`
	CurrentGroupMultiplier       float64                              `json:"current_group_multiplier"`
	PackageScope                 string                               `json:"package_scope"`
	SchedulableAccountCount      int                                  `json:"schedulable_account_count"`
	ActualBlendedMultiplier      *float64                             `json:"actual_blended_multiplier,omitempty"`
	RecommendedBlendedMultiplier *float64                             `json:"recommended_blended_multiplier,omitempty"`
	WorstCaseMultiplier          *float64                             `json:"worst_case_multiplier,omitempty"`
	MinimumGroupMultiplier       *float64                             `json:"minimum_group_multiplier,omitempty"`
	SafeGroupMultiplier          *float64                             `json:"safe_group_multiplier,omitempty"`
	Status                       string                               `json:"status"`
	Notes                        []string                             `json:"notes,omitempty"`
	Accounts                     []*OpsGroupRateRecommendationAccount `json:"accounts"`
}

type OpsGroupRateRecommendationResponse struct {
	Params       OpsGroupRateRecommendationFilter        `json:"params"`
	PackageBasis *OpsGroupRateRecommendationPackageBasis `json:"package_basis,omitempty"`
	Groups       []*OpsGroupRateRecommendationGroup      `json:"groups"`
}

type OpsGroupRateRecommendationSourceData struct {
	Packages []*OpsGroupRateRecommendationPackageBasis
	Groups   []*OpsGroupRateRecommendationSourceGroup
	Usage    map[int64]map[int64]OpsGroupRateRecommendationUsageShare
	Samples  map[int64]*OpsUpstreamMultiplierSample
}

type OpsGroupRateRecommendationSourceGroup struct {
	GroupID              int64
	GroupName            string
	RateMultiplier       float64
	PackageScope         string
	AllowImageGeneration bool
	Accounts             []*OpsGroupRateRecommendationSourceAccount
}

type OpsGroupRateRecommendationSourceAccount struct {
	AccountID       int64
	AccountName     string
	Platform        string
	Type            string
	Status          string
	Schedulable     bool
	CurrentPriority int
	BindingPriority int
	BaseURL         string
	KeyPrefix       string
}
