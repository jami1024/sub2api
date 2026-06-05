package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// 渠道监控聚合层：把 latest + availability 拼成 admin/user 视图所需的 summary / detail。
// 所有方法都遵守"失败仅日志，返回零值"的原则，避免 N+1 查询失败拖垮列表渲染。

// BatchMonitorStatusSummary 批量聚合多个监控的 latest + 7d 可用率（admin/user list 用，消除 N+1）。
// 失败时返回空 map，错误仅日志，不影响列表渲染。
//
// 参数：
//   - ids: 要聚合的 monitor ID 列表
//   - primaryByID: monitor ID -> primary model（用于读 7d 可用率与 latest 状态）
//   - extrasByID: monitor ID -> extra models 列表（用于读 latest 状态填充 ExtraModels）
func (s *ChannelMonitorService) BatchMonitorStatusSummary(
	ctx context.Context,
	ids []int64,
	providerByID map[int64]string,
	primaryByID map[int64]string,
	extrasByID map[int64][]string,
	intervalByID map[int64]int,
) map[int64]MonitorStatusSummary {
	out := make(map[int64]MonitorStatusSummary, len(ids))
	if len(ids) == 0 {
		return out
	}
	now := time.Now()
	usageHealth := s.batchOpenAIUsageHealth(ctx, ids, providerByID, primaryByID, extrasByID, now.Add(-monitorUserRealtimeWindow))
	historyIDs := nonOpenAIMonitorIDs(ids, providerByID)
	latestMap := map[int64][]*ChannelMonitorLatest{}
	availMap := map[int64][]*ChannelMonitorAvailability{}
	if len(historyIDs) > 0 {
		var err error
		latestMap, err = s.repo.ListLatestForMonitorIDs(ctx, historyIDs)
		if err != nil {
			slog.Warn("channel_monitor: batch load latest failed", "error", err)
			latestMap = map[int64][]*ChannelMonitorLatest{}
		}
		availMap, err = s.repo.ComputeAvailabilityForMonitors(ctx, historyIDs, monitorAvailability7Days)
		if err != nil {
			slog.Warn("channel_monitor: batch compute availability failed", "error", err)
			availMap = map[int64][]*ChannelMonitorAvailability{}
		}
	}

	for _, id := range ids {
		if providerByID[id] == MonitorProviderOpenAI {
			out[id] = buildUsageHealthStatusSummary(
				primaryByID[id],
				extrasByID[id],
				usageHealth,
			)
			continue
		}
		out[id] = buildStatusSummary(
			indexLatestByModel(latestMap[id]),
			indexAvailabilityByModel(availMap[id]),
			primaryByID[id],
			extrasByID[id],
		)
	}
	return out
}

func (s *ChannelMonitorService) batchOpenAIUsageHealth(
	ctx context.Context,
	ids []int64,
	providerByID map[int64]string,
	primaryByID map[int64]string,
	extrasByID map[int64][]string,
	since time.Time,
) map[string]*ChannelMonitorUsageHealth {
	models := openAIUsageModels(ids, providerByID, primaryByID, extrasByID)
	if len(models) == 0 {
		return map[string]*ChannelMonitorUsageHealth{}
	}
	health, err := s.repo.ComputeOpenAIUsageHealthByModels(ctx, models, since)
	if err != nil {
		slog.Warn("channel_monitor: batch load openai usage health failed", "error", err)
		return map[string]*ChannelMonitorUsageHealth{}
	}
	return health
}

func (s *ChannelMonitorService) batchOpenAIUsageLatest(
	ctx context.Context,
	ids []int64,
	providerByID map[int64]string,
	primaryByID map[int64]string,
	extrasByID map[int64][]string,
	intervalByID map[int64]int,
) map[string]*ChannelMonitorUsageLogLatest {
	models := make([]string, 0)
	longestInterval := 0
	for _, id := range ids {
		if providerByID[id] != MonitorProviderOpenAI {
			continue
		}
		models = append(models, primaryByID[id])
		models = append(models, extrasByID[id]...)
		interval := intervalByID[id]
		if interval > longestInterval {
			longestInterval = interval
		}
	}
	if len(models) == 0 {
		return map[string]*ChannelMonitorUsageLogLatest{}
	}
	latest, err := s.repo.ListLatestSuccessfulOpenAIUsageByModels(ctx, models, channelMonitorUsageLogSince(time.Now(), longestInterval))
	if err != nil {
		slog.Warn("channel_monitor: batch load openai usage logs failed", "error", err)
		return map[string]*ChannelMonitorUsageLogLatest{}
	}
	return latest
}

func openAIUsageModels(
	ids []int64,
	providerByID map[int64]string,
	primaryByID map[int64]string,
	extrasByID map[int64][]string,
) []string {
	models := make([]string, 0)
	for _, id := range ids {
		if providerByID[id] != MonitorProviderOpenAI {
			continue
		}
		models = append(models, primaryByID[id])
		models = append(models, extrasByID[id]...)
	}
	return normalizeMonitorModels(models)
}

func normalizeMonitorModels(models []string) []string {
	out := make([]string, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		out = append(out, model)
	}
	return out
}

func buildUsageHealthStatusSummary(
	primary string,
	extras []string,
	health map[string]*ChannelMonitorUsageHealth,
) MonitorStatusSummary {
	summary := MonitorStatusSummary{ExtraModels: make([]ExtraModelStatus, 0, len(extras))}
	if h := health[strings.TrimSpace(primary)]; h != nil {
		summary.PrimaryStatus = h.LatestStatus
		summary.PrimaryLatencyMs = h.LatestLatencyMs
		summary.Availability7d = h.AvailabilityPct
	}
	for _, model := range extras {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		entry := ExtraModelStatus{Model: model}
		if h := health[model]; h != nil {
			entry.Status = h.LatestStatus
			entry.LatencyMs = h.LatestLatencyMs
		}
		summary.ExtraModels = append(summary.ExtraModels, entry)
	}
	return summary
}

func buildUsageLogStatusSummary(
	primary string,
	extras []string,
	latest map[string]*ChannelMonitorUsageLogLatest,
	since time.Time,
	now time.Time,
) MonitorStatusSummary {
	summary := MonitorStatusSummary{ExtraModels: make([]ExtraModelStatus, 0, len(extras))}
	if l := usageLogLatestWithinWindow(latest[strings.TrimSpace(primary)], since); l != nil {
		res := usageLogLatestToCheckResult(primary, l, now)
		summary.PrimaryStatus = res.Status
		summary.PrimaryLatencyMs = res.LatencyMs
		summary.Availability7d = 100
	}
	for _, model := range extras {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		entry := ExtraModelStatus{Model: model}
		if l := usageLogLatestWithinWindow(latest[model], since); l != nil {
			res := usageLogLatestToCheckResult(model, l, now)
			entry.Status = res.Status
			entry.LatencyMs = res.LatencyMs
		}
		summary.ExtraModels = append(summary.ExtraModels, entry)
	}
	return summary
}

func usageLogLatestWithinWindow(latest *ChannelMonitorUsageLogLatest, since time.Time) *ChannelMonitorUsageLogLatest {
	if latest == nil || latest.CreatedAt.Before(since) {
		return nil
	}
	return latest
}

// ListUserView 用户只读视图：列出所有 enabled 监控的概览。
// 使用批量聚合接口避免 N+1：
//
//	1 次查 monitors；
//	1 次批量 latest（含 ping_latency_ms）；
//	1 次批量 7d availability；
//	1 次批量 timeline（主模型最近 N 条）。
func (s *ChannelMonitorService) ListUserView(ctx context.Context) ([]*UserMonitorView, error) {
	if cached := s.getCachedUserList(); cached != nil {
		return cached, nil
	}
	v, err, _ := s.userListGroup.Do("user-list", func() (any, error) {
		if cached := s.getCachedUserList(); cached != nil {
			return cached, nil
		}
		views, err := s.loadUserView(ctx)
		if err != nil {
			return nil, err
		}
		s.setCachedUserList(views)
		return cloneUserMonitorViews(views), nil
	})
	if err != nil {
		return nil, err
	}
	views, ok := v.([]*UserMonitorView)
	if !ok {
		return nil, fmt.Errorf("load user monitor list: unexpected cache value")
	}
	return views, nil
}

func (s *ChannelMonitorService) loadUserView(ctx context.Context) ([]*UserMonitorView, error) {
	monitors, err := s.repo.ListEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enabled monitors: %w", err)
	}
	if len(monitors) == 0 {
		return []*UserMonitorView{}, nil
	}

	ids, providerByID, primaryByID, extrasByID, intervalByID := collectMonitorIndexes(monitors)
	summaries := s.BatchMonitorStatusSummary(ctx, ids, providerByID, primaryByID, extrasByID, intervalByID)
	latestMap := s.batchLatest(ctx, nonOpenAIMonitorIDs(ids, providerByID))
	timelineMap := s.batchTimeline(ctx, nonOpenAIMonitorIDs(ids, providerByID), primaryByID)
	openAITimelineMap := s.batchOpenAIUsageTimeline(ctx, ids, providerByID, primaryByID)

	views := make([]*UserMonitorView, 0, len(monitors))
	for _, m := range monitors {
		primaryLatest := pickLatest(latestMap[m.ID], m.PrimaryModel)
		timeline := timelineMap[m.ID]
		if m.Provider == MonitorProviderOpenAI {
			timeline = openAITimelineMap[m.ID]
		}
		views = append(views, buildUserViewFromSummary(
			m,
			summaries[m.ID],
			primaryLatest,
			timelineEntriesForUserView(m, timeline),
		))
	}
	return views, nil
}

// collectMonitorIndexes 把 monitors 列表按 ID 展开为聚合查询所需的索引结构。
func collectMonitorIndexes(monitors []*ChannelMonitor) ([]int64, map[int64]string, map[int64]string, map[int64][]string, map[int64]int) {
	ids := make([]int64, 0, len(monitors))
	providerByID := make(map[int64]string, len(monitors))
	primaryByID := make(map[int64]string, len(monitors))
	extrasByID := make(map[int64][]string, len(monitors))
	intervalByID := make(map[int64]int, len(monitors))
	for _, m := range monitors {
		ids = append(ids, m.ID)
		providerByID[m.ID] = m.Provider
		primaryByID[m.ID] = m.PrimaryModel
		extrasByID[m.ID] = m.ExtraModels
		intervalByID[m.ID] = m.IntervalSeconds
	}
	return ids, providerByID, primaryByID, extrasByID, intervalByID
}

func nonOpenAIMonitorIDs(ids []int64, providerByID map[int64]string) []int64 {
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if providerByID[id] != MonitorProviderOpenAI {
			out = append(out, id)
		}
	}
	return out
}

// batchLatest 批量取 latest per model，失败仅日志（与现有 BatchMonitorStatusSummary 一致，不阻断列表渲染）。
func (s *ChannelMonitorService) batchLatest(ctx context.Context, ids []int64) map[int64][]*ChannelMonitorLatest {
	if len(ids) == 0 {
		return map[int64][]*ChannelMonitorLatest{}
	}
	latestMap, err := s.repo.ListLatestForMonitorIDs(ctx, ids)
	if err != nil {
		slog.Warn("channel_monitor: user view batch latest failed", "error", err)
		return map[int64][]*ChannelMonitorLatest{}
	}
	return latestMap
}

// batchTimeline 批量取每个 monitor 主模型最近 monitorTimelineMaxPoints 条历史。
func (s *ChannelMonitorService) batchTimeline(
	ctx context.Context,
	ids []int64,
	primaryByID map[int64]string,
) map[int64][]*ChannelMonitorHistoryEntry {
	timelineMap, err := s.repo.ListRecentHistoryForMonitors(ctx, ids, primaryByID, monitorTimelineMaxPoints)
	if err != nil {
		slog.Warn("channel_monitor: user view batch timeline failed", "error", err)
		return map[int64][]*ChannelMonitorHistoryEntry{}
	}
	return timelineMap
}

func (s *ChannelMonitorService) batchOpenAIUsageTimeline(
	ctx context.Context,
	ids []int64,
	providerByID map[int64]string,
	primaryByID map[int64]string,
) map[int64][]*ChannelMonitorHistoryEntry {
	models := make([]string, 0)
	for _, id := range ids {
		if providerByID[id] == MonitorProviderOpenAI {
			models = append(models, primaryByID[id])
		}
	}
	models = normalizeMonitorModels(models)
	out := make(map[int64][]*ChannelMonitorHistoryEntry)
	if len(models) == 0 {
		return out
	}
	events, err := s.repo.ListRecentOpenAIUsageEventsByModels(ctx, models, time.Now().Add(-monitorUserRealtimeWindow), monitorTimelineMaxPoints)
	if err != nil {
		slog.Warn("channel_monitor: user view batch openai usage timeline failed", "error", err)
		return out
	}
	for _, id := range ids {
		if providerByID[id] != MonitorProviderOpenAI {
			continue
		}
		out[id] = events[strings.TrimSpace(primaryByID[id])]
	}
	return out
}

// pickLatest 从 latest 切片中挑出指定 model 对应项，未命中返回 nil。
func pickLatest(rows []*ChannelMonitorLatest, model string) *ChannelMonitorLatest {
	if model == "" {
		return nil
	}
	for _, r := range rows {
		if r.Model == model {
			return r
		}
	}
	return nil
}

func timelineEntriesForUserView(m *ChannelMonitor, entries []*ChannelMonitorHistoryEntry) []*ChannelMonitorHistoryEntry {
	return entries
}

// GetUserDetail 用户只读视图：单个监控详情（每个模型 7d/15d/30d 可用率与平均延迟）。
// 不暴露 api_key。
func (s *ChannelMonitorService) GetUserDetail(ctx context.Context, id int64) (*UserMonitorDetail, error) {
	if cached := s.getCachedUserDetail(id); cached != nil {
		return cached, nil
	}
	v, err, _ := s.userDetailGroup.Do(fmt.Sprintf("user-detail:%d", id), func() (any, error) {
		if cached := s.getCachedUserDetail(id); cached != nil {
			return cached, nil
		}
		detail, err := s.loadUserDetail(ctx, id)
		if err != nil {
			return nil, err
		}
		s.setCachedUserDetail(id, detail)
		return cloneUserMonitorDetail(detail), nil
	})
	if err != nil {
		return nil, err
	}
	detail, ok := v.(*UserMonitorDetail)
	if !ok || detail == nil {
		return nil, fmt.Errorf("load user monitor detail: unexpected cache value")
	}
	return detail, nil
}

func (s *ChannelMonitorService) loadUserDetail(ctx context.Context, id int64) (*UserMonitorDetail, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !m.Enabled {
		return nil, ErrChannelMonitorNotFound
	}
	if m.Provider == MonitorProviderOpenAI {
		return s.getOpenAIUsageHealthUserDetail(ctx, m)
	}

	latest, err := s.repo.ListLatestPerModel(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list latest per model: %w", err)
	}
	availMap, err := s.collectAvailabilityWindows(ctx, id)
	if err != nil {
		return nil, err
	}

	models := mergeModelDetails(m, latest, availMap)
	return &UserMonitorDetail{
		ID:        m.ID,
		Name:      m.Name,
		Provider:  m.Provider,
		GroupName: m.GroupName,
		Models:    models,
	}, nil
}

func (s *ChannelMonitorService) getCachedUserDetail(id int64) *UserMonitorDetail {
	if s == nil {
		return nil
	}
	s.userDetailMu.RLock()
	entry, ok := s.userDetailCache[id]
	s.userDetailMu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		if ok {
			s.invalidateUserDetailCache(id)
		}
		return nil
	}
	return cloneUserMonitorDetail(entry.detail)
}

func (s *ChannelMonitorService) setCachedUserDetail(id int64, detail *UserMonitorDetail) {
	if s == nil || detail == nil {
		return
	}
	s.userDetailMu.Lock()
	s.userDetailCache[id] = channelMonitorUserDetailCacheEntry{
		detail:    cloneUserMonitorDetail(detail),
		expiresAt: time.Now().Add(monitorUserDetailCacheTTL),
	}
	s.userDetailMu.Unlock()
}

func (s *ChannelMonitorService) invalidateUserDetailCache(id int64) {
	if s == nil {
		return
	}
	s.userDetailMu.Lock()
	delete(s.userDetailCache, id)
	s.userDetailMu.Unlock()
	s.userDetailGroup.Forget(fmt.Sprintf("user-detail:%d", id))
	s.invalidateUserListCache()
}

func (s *ChannelMonitorService) getCachedUserList() []*UserMonitorView {
	if s == nil {
		return nil
	}
	s.userListMu.RLock()
	entry := s.userListCache
	s.userListMu.RUnlock()
	if entry.views == nil || time.Now().After(entry.expiresAt) {
		if entry.views != nil {
			s.invalidateUserListCache()
		}
		return nil
	}
	return cloneUserMonitorViews(entry.views)
}

func (s *ChannelMonitorService) setCachedUserList(views []*UserMonitorView) {
	if s == nil {
		return
	}
	s.userListMu.Lock()
	s.userListCache = channelMonitorUserListCacheEntry{
		views:     cloneUserMonitorViews(views),
		expiresAt: time.Now().Add(monitorUserDetailCacheTTL),
	}
	s.userListMu.Unlock()
}

func (s *ChannelMonitorService) invalidateUserListCache() {
	if s == nil {
		return
	}
	s.userListMu.Lock()
	s.userListCache = channelMonitorUserListCacheEntry{}
	s.userListMu.Unlock()
	s.userListGroup.Forget("user-list")
}

func cloneUserMonitorDetail(in *UserMonitorDetail) *UserMonitorDetail {
	if in == nil {
		return nil
	}
	out := *in
	if in.Models != nil {
		out.Models = append([]ModelDetail(nil), in.Models...)
	}
	return &out
}

func cloneUserMonitorViews(in []*UserMonitorView) []*UserMonitorView {
	if in == nil {
		return nil
	}
	out := make([]*UserMonitorView, 0, len(in))
	for _, v := range in {
		if v == nil {
			out = append(out, nil)
			continue
		}
		cp := *v
		if v.ExtraModels != nil {
			cp.ExtraModels = append([]ExtraModelStatus(nil), v.ExtraModels...)
		}
		if v.Timeline != nil {
			cp.Timeline = append([]UserMonitorTimelinePoint(nil), v.Timeline...)
		}
		out = append(out, &cp)
	}
	return out
}

func (s *ChannelMonitorService) getOpenAIUsageLogUserDetail(ctx context.Context, m *ChannelMonitor) (*UserMonitorDetail, error) {
	models := channelMonitorModels(m)
	now := time.Now()
	since := channelMonitorUsageLogSince(now, m.IntervalSeconds)
	latest, err := s.repo.ListLatestSuccessfulOpenAIUsageByModels(ctx, models, since)
	if err != nil {
		return nil, fmt.Errorf("list latest openai usage logs: %w", err)
	}
	return &UserMonitorDetail{
		ID:        m.ID,
		Name:      m.Name,
		Provider:  m.Provider,
		GroupName: m.GroupName,
		Models:    buildOpenAIUsageLogModelDetails(models, latest, since, now),
	}, nil
}

func (s *ChannelMonitorService) getOpenAIUsageHealthUserDetail(ctx context.Context, m *ChannelMonitor) (*UserMonitorDetail, error) {
	models := channelMonitorModels(m)
	healthMap, err := s.collectOpenAIUsageHealthWindows(ctx, models)
	if err != nil {
		return nil, err
	}
	return &UserMonitorDetail{
		ID:        m.ID,
		Name:      m.Name,
		Provider:  m.Provider,
		GroupName: m.GroupName,
		Models:    buildOpenAIUsageHealthModelDetails(models, healthMap),
	}, nil
}

func (s *ChannelMonitorService) collectOpenAIUsageHealthWindows(ctx context.Context, models []string) (map[int]map[string]*ChannelMonitorUsageHealth, error) {
	out := make(map[int]map[string]*ChannelMonitorUsageHealth, 3)
	now := time.Now()
	windows := []int{monitorAvailability7Days, monitorAvailability15Days, monitorAvailability30Days}
	for _, w := range windows {
		rows, err := s.repo.ComputeOpenAIUsageHealthByModels(ctx, models, now.AddDate(0, 0, -w))
		if err != nil {
			return nil, fmt.Errorf("compute openai usage health %dd: %w", w, err)
		}
		out[w] = rows
	}
	return out, nil
}

func buildOpenAIUsageHealthModelDetails(
	models []string,
	healthMap map[int]map[string]*ChannelMonitorUsageHealth,
) []ModelDetail {
	out := make([]ModelDetail, 0, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		d := ModelDetail{Model: model}
		if h := healthMap[monitorAvailability7Days][model]; h != nil {
			d.LatestStatus = h.LatestStatus
			d.LatestLatencyMs = h.LatestLatencyMs
			d.Availability7d = h.AvailabilityPct
			d.AvgLatency7dMs = h.AvgLatencyMs
		}
		if h := healthMap[monitorAvailability15Days][model]; h != nil {
			d.Availability15d = h.AvailabilityPct
		}
		if h := healthMap[monitorAvailability30Days][model]; h != nil {
			d.Availability30d = h.AvailabilityPct
		}
		out = append(out, d)
	}
	return out
}

func buildOpenAIUsageLogModelDetails(
	models []string,
	latest map[string]*ChannelMonitorUsageLogLatest,
	since time.Time,
	now time.Time,
) []ModelDetail {
	out := make([]ModelDetail, 0, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		d := ModelDetail{Model: model}
		if l := usageLogLatestWithinWindow(latest[model], since); l != nil {
			res := usageLogLatestToCheckResult(model, l, now)
			d.LatestStatus = res.Status
			d.LatestLatencyMs = res.LatencyMs
			d.Availability7d = 100
			d.Availability15d = 100
			d.Availability30d = 100
			d.AvgLatency7dMs = res.LatencyMs
		}
		out = append(out, d)
	}
	return out
}

// collectAvailabilityWindows 一次性查询 7/15/30 天三个窗口，按模型组织。
func (s *ChannelMonitorService) collectAvailabilityWindows(ctx context.Context, monitorID int64) (map[int]map[string]*ChannelMonitorAvailability, error) {
	out := make(map[int]map[string]*ChannelMonitorAvailability, 3)
	windows := []int{monitorAvailability7Days, monitorAvailability15Days, monitorAvailability30Days}
	for _, w := range windows {
		rows, err := s.repo.ComputeAvailability(ctx, monitorID, w)
		if err != nil {
			return nil, fmt.Errorf("compute availability %dd: %w", w, err)
		}
		out[w] = indexAvailabilityByModel(rows)
	}
	return out, nil
}

// ---------- 纯函数 helper（无 IO，可在 batch / 单 monitor / detail 路径复用）----------

// indexLatestByModel 把 latest 切片按 model 索引（小工具，避免在 hot path 重复写）。
func indexLatestByModel(rows []*ChannelMonitorLatest) map[string]*ChannelMonitorLatest {
	m := make(map[string]*ChannelMonitorLatest, len(rows))
	for _, r := range rows {
		m[r.Model] = r
	}
	return m
}

// indexAvailabilityByModel 把 availability 切片按 model 索引。
func indexAvailabilityByModel(rows []*ChannelMonitorAvailability) map[string]*ChannelMonitorAvailability {
	m := make(map[string]*ChannelMonitorAvailability, len(rows))
	for _, r := range rows {
		m[r.Model] = r
	}
	return m
}

// buildStatusSummary 由 latest + availability 字典构造 MonitorStatusSummary。
// 不做任何 IO，纯组装，便于在 batch 与单 monitor 路径复用。
func buildStatusSummary(
	latestByModel map[string]*ChannelMonitorLatest,
	availByModel map[string]*ChannelMonitorAvailability,
	primary string,
	extras []string,
) MonitorStatusSummary {
	summary := MonitorStatusSummary{ExtraModels: make([]ExtraModelStatus, 0, len(extras))}
	if primary != "" {
		if l, ok := latestByModel[primary]; ok {
			summary.PrimaryStatus = l.Status
			summary.PrimaryLatencyMs = l.LatencyMs
		}
		if a, ok := availByModel[primary]; ok {
			summary.Availability7d = a.AvailabilityPct
		}
	}
	for _, model := range extras {
		entry := ExtraModelStatus{Model: model}
		if l, ok := latestByModel[model]; ok {
			entry.Status = l.Status
			entry.LatencyMs = l.LatencyMs
		}
		summary.ExtraModels = append(summary.ExtraModels, entry)
	}
	return summary
}

// buildUserViewFromSummary 用预聚合好的 MonitorStatusSummary + 主模型 latest + timeline 装填 UserMonitorView（无 IO）。
// primaryLatest 可能为 nil（该监控尚无历史）；timelineEntries 可能为空。
func buildUserViewFromSummary(
	m *ChannelMonitor,
	summary MonitorStatusSummary,
	primaryLatest *ChannelMonitorLatest,
	timelineEntries []*ChannelMonitorHistoryEntry,
) *UserMonitorView {
	view := &UserMonitorView{
		ID:               m.ID,
		Name:             m.Name,
		Provider:         m.Provider,
		GroupName:        m.GroupName,
		PrimaryModel:     m.PrimaryModel,
		PrimaryStatus:    summary.PrimaryStatus,
		PrimaryLatencyMs: summary.PrimaryLatencyMs,
		Availability7d:   summary.Availability7d,
		ExtraModels:      summary.ExtraModels,
		Timeline:         buildTimelinePoints(timelineEntries),
	}
	if primaryLatest != nil {
		view.PrimaryPingLatencyMs = primaryLatest.PingLatencyMs
	}
	return view
}

// buildTimelinePoints 把 history entry 裁剪为 timeline 点（去除 message/ID/Model，减小响应体）。
func buildTimelinePoints(entries []*ChannelMonitorHistoryEntry) []UserMonitorTimelinePoint {
	out := make([]UserMonitorTimelinePoint, 0, len(entries))
	for _, e := range entries {
		out = append(out, UserMonitorTimelinePoint{
			Status:        e.Status,
			LatencyMs:     e.LatencyMs,
			PingLatencyMs: e.PingLatencyMs,
			CheckedAt:     e.CheckedAt,
		})
	}
	return out
}

// mergeModelDetails 合并 latest + availability 三个窗口为 ModelDetail 列表。
// 复用 indexLatestByModel，避免在多处重复写 build map 逻辑。
func mergeModelDetails(
	m *ChannelMonitor,
	latest []*ChannelMonitorLatest,
	availMap map[int]map[string]*ChannelMonitorAvailability,
) []ModelDetail {
	all := append([]string{m.PrimaryModel}, m.ExtraModels...)
	latestByModel := indexLatestByModel(latest)
	out := make([]ModelDetail, 0, len(all))
	for _, model := range all {
		d := ModelDetail{Model: model}
		if l, ok := latestByModel[model]; ok {
			d.LatestStatus = l.Status
			d.LatestLatencyMs = l.LatencyMs
		}
		if a, ok := availMap[monitorAvailability7Days][model]; ok {
			d.Availability7d = a.AvailabilityPct
			d.AvgLatency7dMs = a.AvgLatencyMs
		}
		if a, ok := availMap[monitorAvailability15Days][model]; ok {
			d.Availability15d = a.AvailabilityPct
		}
		if a, ok := availMap[monitorAvailability30Days][model]; ok {
			d.Availability30d = a.AvailabilityPct
		}
		out = append(out, d)
	}
	return out
}
