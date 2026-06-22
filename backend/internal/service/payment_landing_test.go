package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"

	"github.com/stretchr/testify/require"
)

func TestGetLandingPackageShowcaseCombinesPackagesAndOpenAIRates(t *testing.T) {
	client := newPaymentConfigServiceTestClient(t)
	svc := NewPaymentConfigService(client, &paymentConfigSettingRepoStub{values: map[string]string{}}, nil)
	ctx := context.Background()

	_, err := client.BalancePackage.Create().
		SetName("专属包-试用级").
		SetPrice(10).
		SetCreditAmount(15).
		SetPackageScope(PackageScopeCodex).
		SetDisplayTags([]string{"试用"}).
		SetForSale(true).
		SetSortOrder(1).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.BalancePackage.Create().
		SetName("专属包-进阶级").
		SetPrice(100).
		SetCreditAmount(400).
		SetPackageScope(PackageScopeCodex).
		SetDisplayTags([]string{"推荐"}).
		SetForSale(true).
		SetSortOrder(2).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.BalancePackage.Create().
		SetName("下架包").
		SetPrice(1).
		SetCreditAmount(1).
		SetPackageScope(PackageScopeCodex).
		SetForSale(false).
		SetSortOrder(3).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Group.Create().
		SetName("gpt pro").
		SetPlatform(domain.PlatformOpenAI).
		SetStatus(domain.StatusActive).
		SetRateMultiplier(0.8).
		SetSortOrder(1).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.Group.Create().
		SetName("gpt pro 高价").
		SetPlatform(domain.PlatformOpenAI).
		SetStatus(domain.StatusActive).
		SetRateMultiplier(1.3).
		SetSortOrder(2).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.Group.Create().
		SetName("停用分组").
		SetPlatform(domain.PlatformOpenAI).
		SetStatus(domain.StatusDisabled).
		SetRateMultiplier(0.1).
		Save(ctx)
	require.NoError(t, err)

	showcase, err := svc.GetLandingPackageShowcase(ctx)
	require.NoError(t, err)
	require.Len(t, showcase.Packages, 2)
	require.Len(t, showcase.UsageRates, 2)

	trial := showcase.Packages[0]
	require.Equal(t, "专属包-试用级", trial.Name)
	require.Equal(t, 1.5, trial.ArrivalMultiplier)
	require.Equal(t, 6.7, trial.ArrivalDiscount)
	require.Equal(t, 18.75, trial.EffectiveCreditAmount)
	require.Equal(t, 5.3, trial.EffectiveDiscount)
	require.Equal(t, "约 6.7 折", trial.ArrivalDiscountLabel)
	require.Equal(t, "综合约 5.3 折", trial.EffectiveDiscountLabel)

	pro := showcase.Packages[1]
	require.Equal(t, "专属包-进阶级", pro.Name)
	require.Equal(t, 4.0, pro.ArrivalMultiplier)
	require.Equal(t, 2.5, pro.ArrivalDiscount)
	require.Equal(t, 500.0, pro.EffectiveCreditAmount)
	require.Equal(t, 2.0, pro.EffectiveDiscount)
	require.Equal(t, "低至 2.5 折", pro.ArrivalDiscountLabel)
	require.Equal(t, "综合低至 2 折", pro.EffectiveDiscountLabel)

	require.Equal(t, "gpt pro", showcase.PrimaryUsageRate.GroupName)
	require.Equal(t, 0.8, showcase.PrimaryUsageRate.RateMultiplier)
	require.Equal(t, "同样余额可多用约 25%", showcase.PrimaryUsageRate.ValueLiftLabel)
}
