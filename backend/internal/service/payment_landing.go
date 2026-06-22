package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/internal/domain"
)

// LandingPackageShowcase is a public, read-only projection for the marketing
// landing page. It intentionally contains no payment provider details or keys.
type LandingPackageShowcase struct {
	Packages         []LandingBalancePackage `json:"packages"`
	UsageRates       []LandingUsageRate      `json:"usage_rates"`
	PrimaryUsageRate *LandingUsageRate       `json:"primary_usage_rate,omitempty"`
}

type LandingBalancePackage struct {
	ID                     int64    `json:"id"`
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	Price                  float64  `json:"price"`
	CreditAmount           float64  `json:"credit_amount"`
	PackageScope           string   `json:"package_scope"`
	ProductName            string   `json:"product_name"`
	DisplayTags            []string `json:"display_tags"`
	SortOrder              int      `json:"sort_order"`
	ArrivalMultiplier      float64  `json:"arrival_multiplier"`
	ArrivalDiscount        float64  `json:"arrival_discount"`
	ArrivalDiscountLabel   string   `json:"arrival_discount_label"`
	EffectiveCreditAmount  float64  `json:"effective_credit_amount"`
	EffectiveDiscount      float64  `json:"effective_discount"`
	EffectiveDiscountLabel string   `json:"effective_discount_label"`
}

type LandingUsageRate struct {
	GroupID          int64   `json:"group_id"`
	GroupName        string  `json:"group_name"`
	RateMultiplier   float64 `json:"rate_multiplier"`
	RateLabel        string  `json:"rate_label"`
	ValueLiftPercent float64 `json:"value_lift_percent"`
	ValueLiftLabel   string  `json:"value_lift_label"`
}

func (s *PaymentConfigService) GetLandingPackageShowcase(ctx context.Context) (*LandingPackageShowcase, error) {
	packages, err := s.ListBalancePackages(ctx, true)
	if err != nil {
		return nil, err
	}

	groups, err := s.entClient.Group.Query().
		Where(group.PlatformEQ(domain.PlatformOpenAI), group.StatusEQ(domain.StatusActive)).
		Order(group.BySortOrder(), group.ByID()).
		All(ctx)
	if err != nil {
		return nil, err
	}

	usageRates := make([]LandingUsageRate, 0, len(groups))
	for _, g := range groups {
		usageRates = append(usageRates, buildLandingUsageRate(g))
	}
	primary := selectPrimaryLandingUsageRate(usageRates)

	primaryRate := 1.0
	if primary != nil && primary.RateMultiplier > 0 {
		primaryRate = primary.RateMultiplier
	}
	items := make([]LandingBalancePackage, 0, len(packages))
	for _, pkg := range packages {
		items = append(items, buildLandingBalancePackage(pkg, primaryRate))
	}

	return &LandingPackageShowcase{
		Packages:         items,
		UsageRates:       usageRates,
		PrimaryUsageRate: primary,
	}, nil
}

func buildLandingUsageRate(g *dbent.Group) LandingUsageRate {
	rate := round1(g.RateMultiplier)
	lift := 0.0
	label := ""
	if rate > 0 && rate < 1 {
		lift = round1((1/rate - 1) * 100)
		label = fmt.Sprintf("同样余额可多用约 %s", formatPercent(lift))
	}
	return LandingUsageRate{
		GroupID:          int64(g.ID),
		GroupName:        g.Name,
		RateMultiplier:   rate,
		RateLabel:        fmt.Sprintf("%s 使用倍率 %sx", g.Name, formatNumber(rate)),
		ValueLiftPercent: lift,
		ValueLiftLabel:   label,
	}
}

func buildLandingBalancePackage(pkg *dbent.BalancePackage, usageRate float64) LandingBalancePackage {
	arrivalMultiplier := safeDivide(pkg.CreditAmount, pkg.Price)
	arrivalDiscount := safeDivide(pkg.Price, pkg.CreditAmount) * 10
	effectiveCredit := pkg.CreditAmount
	if usageRate > 0 {
		effectiveCredit = pkg.CreditAmount / usageRate
	}
	effectiveDiscount := safeDivide(pkg.Price, effectiveCredit) * 10
	return LandingBalancePackage{
		ID:                     int64(pkg.ID),
		Name:                   pkg.Name,
		Description:            pkg.Description,
		Price:                  pkg.Price,
		CreditAmount:           pkg.CreditAmount,
		PackageScope:           pkg.PackageScope,
		ProductName:            pkg.ProductName,
		DisplayTags:            pkg.DisplayTags,
		SortOrder:              pkg.SortOrder,
		ArrivalMultiplier:      round1(arrivalMultiplier),
		ArrivalDiscount:        round1(arrivalDiscount),
		ArrivalDiscountLabel:   discountLabel(round1(arrivalDiscount), false),
		EffectiveCreditAmount:  round2(effectiveCredit),
		EffectiveDiscount:      round1(effectiveDiscount),
		EffectiveDiscountLabel: discountLabel(round1(effectiveDiscount), true),
	}
}

func selectPrimaryLandingUsageRate(rates []LandingUsageRate) *LandingUsageRate {
	if len(rates) == 0 {
		return nil
	}
	for i := range rates {
		if strings.EqualFold(strings.TrimSpace(rates[i].GroupName), "gpt pro") {
			return &rates[i]
		}
	}
	return &rates[0]
}

func safeDivide(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }
func round2(v float64) float64 { return math.Round(v*100) / 100 }

func formatNumber(v float64) string {
	if math.Abs(v-math.Round(v)) < 0.0000001 {
		return fmt.Sprintf("%.0f", v)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", v), "0"), ".")
}

func formatPercent(v float64) string {
	return formatNumber(v) + "%"
}

func discountLabel(discount float64, effective bool) string {
	prefix := "约 "
	if effective {
		prefix = "综合约 "
	}
	if discount <= 2.5 {
		prefix = "低至 "
		if effective {
			prefix = "综合低至 "
		}
	}
	return prefix + formatNumber(discount) + " 折"
}
