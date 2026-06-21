package service

import (
	"context"
	"sort"
	"strings"
)

const (
	opsGroupRateDefaultModel        = "gpt-5.4"
	opsGroupRateDefaultPackageScope = "codex"
	opsGroupRateDefaultProfitMargin = 0.20
	opsGroupRateDefaultSafetyFactor = 1.20
	opsGroupRateDefaultUsageDays    = 7
	opsGroupRateMaxUsageDays        = 30
)

func (s *OpsService) GetGroupRateRecommendations(ctx context.Context, filter *OpsGroupRateRecommendationFilter) (*OpsGroupRateRecommendationResponse, error) {
	normalized := normalizeOpsGroupRateRecommendationFilter(filter)
	if s == nil || s.opsRepo == nil {
		return &OpsGroupRateRecommendationResponse{Params: normalized, Groups: []*OpsGroupRateRecommendationGroup{}}, nil
	}
	source, err := s.opsRepo.GetGroupRateRecommendationSourceData(ctx, &normalized)
	if err != nil {
		return nil, err
	}
	basis := selectGroupRatePackageBasis(source.Packages, normalized.PackageScope)
	groups := buildGroupRateRecommendations(source, basis, normalized)
	return &OpsGroupRateRecommendationResponse{Params: normalized, PackageBasis: basis, Groups: groups}, nil
}

func normalizeOpsGroupRateRecommendationFilter(filter *OpsGroupRateRecommendationFilter) OpsGroupRateRecommendationFilter {
	out := OpsGroupRateRecommendationFilter{}
	if filter != nil {
		out = *filter
	}
	out.Model = normalizeOpsUpstreamMultiplierModel(out.Model)
	if strings.TrimSpace(out.Model) == "" {
		out.Model = opsGroupRateDefaultModel
	}
	out.PackageScope = strings.TrimSpace(out.PackageScope)
	if out.PackageScope == "" {
		out.PackageScope = opsGroupRateDefaultPackageScope
	}
	if out.ProfitMargin <= 0 || out.ProfitMargin >= 0.95 {
		out.ProfitMargin = opsGroupRateDefaultProfitMargin
	}
	if out.SafetyFactor <= 0 {
		out.SafetyFactor = opsGroupRateDefaultSafetyFactor
	}
	if out.UsageDays <= 0 {
		out.UsageDays = opsGroupRateDefaultUsageDays
	}
	if out.UsageDays > opsGroupRateMaxUsageDays {
		out.UsageDays = opsGroupRateMaxUsageDays
	}
	return out
}

func selectGroupRatePackageBasis(packages []*OpsGroupRateRecommendationPackageBasis, packageScope string) *OpsGroupRateRecommendationPackageBasis {
	var best *OpsGroupRateRecommendationPackageBasis
	for _, pkg := range packages {
		if pkg == nil || pkg.Price <= 0 || pkg.CreditAmount <= 0 || pkg.RevenuePerCredit <= 0 {
			continue
		}
		if packageScope != "" && pkg.PackageScope != packageScope {
			continue
		}
		if best == nil || pkg.RevenuePerCredit < best.RevenuePerCredit || (pkg.RevenuePerCredit == best.RevenuePerCredit && pkg.PackageID < best.PackageID) {
			copyPkg := *pkg
			best = &copyPkg
		}
	}
	return best
}

func buildGroupRateRecommendations(source *OpsGroupRateRecommendationSourceData, basis *OpsGroupRateRecommendationPackageBasis, filter OpsGroupRateRecommendationFilter) []*OpsGroupRateRecommendationGroup {
	if source == nil {
		return []*OpsGroupRateRecommendationGroup{}
	}
	out := make([]*OpsGroupRateRecommendationGroup, 0, len(source.Groups))
	for _, group := range source.Groups {
		if group == nil || group.AllowImageGeneration {
			continue
		}
		item := buildOneGroupRateRecommendation(group, source, basis, filter)
		out = append(out, item)
	}
	return out
}

func buildOneGroupRateRecommendation(group *OpsGroupRateRecommendationSourceGroup, source *OpsGroupRateRecommendationSourceData, basis *OpsGroupRateRecommendationPackageBasis, filter OpsGroupRateRecommendationFilter) *OpsGroupRateRecommendationGroup {
	item := &OpsGroupRateRecommendationGroup{
		GroupID:                group.GroupID,
		GroupName:              group.GroupName,
		CurrentGroupMultiplier: group.RateMultiplier,
		PackageScope:           group.PackageScope,
		Status:                 OpsGroupRateRecommendationStatusInsufficient,
		Accounts:               []*OpsGroupRateRecommendationAccount{},
	}
	if basis == nil || basis.RevenuePerCredit <= 0 {
		item.Notes = append(item.Notes, "缺少可用套餐口径")
	}

	usageByAccount := source.Usage[group.GroupID]
	participants := []*OpsGroupRateRecommendationAccount{}
	for _, acc := range group.Accounts {
		if acc == nil {
			continue
		}
		usage := usageByAccount[acc.AccountID]
		sample := source.Samples[acc.AccountID]
		account := &OpsGroupRateRecommendationAccount{
			AccountID:         acc.AccountID,
			AccountName:       acc.AccountName,
			BaseURL:           acc.BaseURL,
			KeyPrefix:         acc.KeyPrefix,
			Schedulable:       acc.Schedulable,
			Status:            acc.Status,
			CurrentPriority:   acc.CurrentPriority,
			BindingPriority:   acc.BindingPriority,
			RequestCount:      usage.RequestCount,
			RequestShare:      roundFloat(usage.RequestShare, 6),
			StandardCost:      roundFloat(usage.StandardCost, 8),
			StandardCostShare: roundFloat(usage.StandardCostShare, 6),
		}
		if sample != nil {
			account.MultiplierStatus = sample.Status
			account.UpstreamMultiplier = sample.Multiplier
			measuredAt := sample.MeasuredAt.UTC()
			account.MultiplierMeasuredAt = &measuredAt
		}
		if isGroupRateRecommendationParticipant(acc, account, filter) {
			account.ParticipatesInAdvice = true
			participants = append(participants, account)
			if acc.Schedulable {
				item.SchedulableAccountCount++
			}
		} else {
			account.Note = groupRateAccountSkipNote(acc, account, filter)
		}
		item.Accounts = append(item.Accounts, account)
	}

	assignRecommendedWeights(participants)
	actual := blendedMultiplierFromUsage(participants)
	recommended := blendedMultiplierFromRecommendedWeights(participants)
	worst := worstCaseMultiplier(participants)
	item.ActualBlendedMultiplier = actual
	item.RecommendedBlendedMultiplier = recommended
	item.WorstCaseMultiplier = worst
	if basis != nil {
		basisMultiplier := recommended
		if actual != nil {
			basisMultiplier = actual
		}
		if basisMultiplier != nil {
			min := recommendationRequiredGroupMultiplier(*basisMultiplier, basis.RevenuePerCredit, filter.ProfitMargin, filter.SafetyFactor)
			item.MinimumGroupMultiplier = &min
		}
	}
	if basis != nil && worst != nil {
		safe := recommendationRequiredGroupMultiplier(*worst, basis.RevenuePerCredit, filter.ProfitMargin, filter.SafetyFactor)
		item.SafeGroupMultiplier = &safe
	}
	item.Status = classifyGroupRateRecommendation(item)
	for _, acc := range item.Accounts {
		if acc.ParticipatesInAdvice && acc.Note == "" {
			acc.Note = groupRateParticipantNote(acc, len(participants))
		}
	}
	if len(participants) == 0 {
		item.Notes = append(item.Notes, "没有可参与建议的上游账号")
	}
	sortGroupRateAccounts(item.Accounts)
	return item
}

func isGroupRateRecommendationParticipant(source *OpsGroupRateRecommendationSourceAccount, account *OpsGroupRateRecommendationAccount, filter OpsGroupRateRecommendationFilter) bool {
	if source == nil || account == nil {
		return false
	}
	if source.Platform != "" && source.Platform != PlatformOpenAI {
		return false
	}
	if source.Type != "" && source.Type != AccountTypeAPIKey {
		return false
	}
	if source.Status != StatusActive {
		return false
	}
	if !filter.IncludeUnschedulable && !source.Schedulable {
		return false
	}
	if !filter.IncludeSelfHosted && isSelfHostedGroupRateAccount(source.AccountName) {
		return false
	}
	if account.UpstreamMultiplier == nil || *account.UpstreamMultiplier <= 0 || account.MultiplierStatus != OpsUpstreamMultiplierStatusSuccess {
		return false
	}
	return true
}

func groupRateAccountSkipNote(source *OpsGroupRateRecommendationSourceAccount, account *OpsGroupRateRecommendationAccount, filter OpsGroupRateRecommendationFilter) string {
	if source == nil || account == nil {
		return "数据异常"
	}
	if source.Platform != "" && source.Platform != PlatformOpenAI {
		return "非 OpenAI 账号，不参与建议"
	}
	if source.Type != "" && source.Type != AccountTypeAPIKey {
		return "非 API Key 账号，不参与建议"
	}
	if source.Status != StatusActive {
		return "账号未启用，不参与建议"
	}
	if !filter.IncludeUnschedulable && !source.Schedulable {
		return "不可调度，不参与建议"
	}
	if !filter.IncludeSelfHosted && isSelfHostedGroupRateAccount(source.AccountName) {
		return "自建账号默认不参与建议"
	}
	if account.UpstreamMultiplier == nil || account.MultiplierStatus != OpsUpstreamMultiplierStatusSuccess {
		return "缺少成功倍率样本，请先检测"
	}
	return "不参与建议"
}

func assignRecommendedWeights(accounts []*OpsGroupRateRecommendationAccount) {
	if len(accounts) == 0 {
		return
	}
	if len(accounts) == 1 {
		accounts[0].RecommendedWeight = 1
		accounts[0].RecommendedPriority = 1
		return
	}
	inv := make([]float64, len(accounts))
	total := 0.0
	for i, account := range accounts {
		m := 0.0
		if account.UpstreamMultiplier != nil {
			m = *account.UpstreamMultiplier
		}
		if m <= 0 {
			continue
		}
		inv[i] = 1 / m
		total += inv[i]
	}
	if total <= 0 {
		return
	}
	for i, account := range accounts {
		account.RecommendedWeight = inv[i] / total
	}
	if len(accounts) >= 3 {
		capRecommendedWeights(accounts, 0.5)
	}
	for _, account := range accounts {
		account.RecommendedWeight = roundFloat(account.RecommendedWeight, 4)
		account.RecommendedPriority = recommendedPriorityFromWeight(account.RecommendedWeight)
	}
}

func capRecommendedWeights(accounts []*OpsGroupRateRecommendationAccount, capValue float64) {
	for iter := 0; iter < 5; iter++ {
		over := 0.0
		underTotal := 0.0
		for _, account := range accounts {
			if account.RecommendedWeight > capValue {
				over += account.RecommendedWeight - capValue
				account.RecommendedWeight = capValue
			} else {
				underTotal += account.RecommendedWeight
			}
		}
		if over <= 0 || underTotal <= 0 {
			break
		}
		for _, account := range accounts {
			if account.RecommendedWeight < capValue {
				account.RecommendedWeight += over * (account.RecommendedWeight / underTotal)
			}
		}
	}
	total := 0.0
	for _, account := range accounts {
		total += account.RecommendedWeight
	}
	if total > 0 {
		for _, account := range accounts {
			account.RecommendedWeight /= total
		}
	}
}

func recommendedPriorityFromWeight(weight float64) int {
	switch {
	case weight >= 0.40:
		return 1
	case weight >= 0.25:
		return 2
	case weight >= 0.15:
		return 4
	case weight >= 0.05:
		return 6
	default:
		return 8
	}
}

func blendedMultiplierFromUsage(accounts []*OpsGroupRateRecommendationAccount) *float64 {
	totalUsage := 0.0
	value := 0.0
	for _, account := range accounts {
		if account.UpstreamMultiplier == nil || account.StandardCostShare <= 0 {
			continue
		}
		totalUsage += account.StandardCostShare
		value += *account.UpstreamMultiplier * account.StandardCostShare
	}
	if totalUsage <= 0 {
		return blendedMultiplierFromRecommendedWeights(accounts)
	}
	result := roundFloat(value/totalUsage, 6)
	return &result
}

func blendedMultiplierFromRecommendedWeights(accounts []*OpsGroupRateRecommendationAccount) *float64 {
	value := 0.0
	total := 0.0
	for _, account := range accounts {
		if account.UpstreamMultiplier == nil || account.RecommendedWeight <= 0 {
			continue
		}
		value += *account.UpstreamMultiplier * account.RecommendedWeight
		total += account.RecommendedWeight
	}
	if total <= 0 {
		return nil
	}
	result := roundFloat(value/total, 6)
	return &result
}

func worstCaseMultiplier(accounts []*OpsGroupRateRecommendationAccount) *float64 {
	maxValue := 0.0
	for _, account := range accounts {
		if account.UpstreamMultiplier != nil && *account.UpstreamMultiplier > maxValue {
			maxValue = *account.UpstreamMultiplier
		}
	}
	if maxValue <= 0 {
		return nil
	}
	result := roundFloat(maxValue, 6)
	return &result
}

func recommendationRequiredGroupMultiplier(upstreamMultiplier, revenuePerCredit, profitMargin, safetyFactor float64) float64 {
	if upstreamMultiplier <= 0 || revenuePerCredit <= 0 || profitMargin >= 1 {
		return 0
	}
	return roundFloat(upstreamMultiplier*safetyFactor/(1-profitMargin)/revenuePerCredit, 6)
}

func classifyGroupRateRecommendation(group *OpsGroupRateRecommendationGroup) string {
	if group == nil || group.MinimumGroupMultiplier == nil {
		return OpsGroupRateRecommendationStatusInsufficient
	}
	if group.SafeGroupMultiplier != nil && group.CurrentGroupMultiplier >= *group.SafeGroupMultiplier {
		return OpsGroupRateRecommendationStatusSafe
	}
	if group.CurrentGroupMultiplier >= *group.MinimumGroupMultiplier {
		return OpsGroupRateRecommendationStatusBasicSafe
	}
	return OpsGroupRateRecommendationStatusLow
}

func groupRateParticipantNote(account *OpsGroupRateRecommendationAccount, participantCount int) string {
	if account == nil || account.UpstreamMultiplier == nil {
		return ""
	}
	if participantCount == 1 {
		return "唯一可用账号，建议补充备份账号"
	}
	if account.RecommendedWeight >= 0.35 {
		return "成本较低，建议主力"
	}
	if account.RecommendedWeight >= 0.15 {
		return "建议保留补充权重"
	}
	return "成本较高，建议热备"
}

func sortGroupRateAccounts(accounts []*OpsGroupRateRecommendationAccount) {
	sort.SliceStable(accounts, func(i, j int) bool {
		if accounts[i].ParticipatesInAdvice != accounts[j].ParticipatesInAdvice {
			return accounts[i].ParticipatesInAdvice
		}
		if accounts[i].RecommendedWeight != accounts[j].RecommendedWeight {
			return accounts[i].RecommendedWeight > accounts[j].RecommendedWeight
		}
		return accounts[i].AccountID < accounts[j].AccountID
	})
}

func isSelfHostedGroupRateAccount(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	return strings.Contains(name, "自建") || strings.Contains(name, "self-host") || strings.Contains(name, "selfhost")
}
