package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type apiKeyPackageScopeRepoStub struct {
	exists bool
	created *APIKey
}

func (s *apiKeyPackageScopeRepoStub) Create(ctx context.Context, key *APIKey) error {
	s.created = key
	return nil
}
func (s *apiKeyPackageScopeRepoStub) GetByID(ctx context.Context, id int64) (*APIKey, error) { panic("unexpected GetByID call") }
func (s *apiKeyPackageScopeRepoStub) GetKeyAndOwnerID(ctx context.Context, id int64) (string, int64, error) {
	panic("unexpected GetKeyAndOwnerID call")
}
func (s *apiKeyPackageScopeRepoStub) GetByKey(ctx context.Context, key string) (*APIKey, error) {
	panic("unexpected GetByKey call")
}
func (s *apiKeyPackageScopeRepoStub) GetByKeyForAuth(ctx context.Context, key string) (*APIKey, error) {
	panic("unexpected GetByKeyForAuth call")
}
func (s *apiKeyPackageScopeRepoStub) Update(ctx context.Context, key *APIKey) error { panic("unexpected Update call") }
func (s *apiKeyPackageScopeRepoStub) Delete(ctx context.Context, id int64) error { panic("unexpected Delete call") }
func (s *apiKeyPackageScopeRepoStub) ListByUserID(ctx context.Context, userID int64, params pagination.PaginationParams, filters APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByUserID call")
}
func (s *apiKeyPackageScopeRepoStub) VerifyOwnership(ctx context.Context, userID int64, apiKeyIDs []int64) ([]int64, error) {
	panic("unexpected VerifyOwnership call")
}
func (s *apiKeyPackageScopeRepoStub) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	panic("unexpected CountByUserID call")
}
func (s *apiKeyPackageScopeRepoStub) ExistsByKey(ctx context.Context, key string) (bool, error) {
	return s.exists, nil
}
func (s *apiKeyPackageScopeRepoStub) ListByGroupID(ctx context.Context, groupID int64, params pagination.PaginationParams) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByGroupID call")
}
func (s *apiKeyPackageScopeRepoStub) SearchAPIKeys(ctx context.Context, userID int64, keyword string, limit int) ([]APIKey, error) {
	panic("unexpected SearchAPIKeys call")
}
func (s *apiKeyPackageScopeRepoStub) ClearGroupIDByGroupID(ctx context.Context, groupID int64) (int64, error) {
	panic("unexpected ClearGroupIDByGroupID call")
}
func (s *apiKeyPackageScopeRepoStub) UpdateGroupIDByUserAndGroup(ctx context.Context, userID, oldGroupID, newGroupID int64) (int64, error) {
	panic("unexpected UpdateGroupIDByUserAndGroup call")
}
func (s *apiKeyPackageScopeRepoStub) CountByGroupID(ctx context.Context, groupID int64) (int64, error) {
	panic("unexpected CountByGroupID call")
}
func (s *apiKeyPackageScopeRepoStub) ListKeysByUserID(ctx context.Context, userID int64) ([]string, error) {
	panic("unexpected ListKeysByUserID call")
}
func (s *apiKeyPackageScopeRepoStub) ListKeysByGroupID(ctx context.Context, groupID int64) ([]string, error) {
	panic("unexpected ListKeysByGroupID call")
}
func (s *apiKeyPackageScopeRepoStub) IncrementQuotaUsed(ctx context.Context, id int64, amount float64) (float64, error) {
	panic("unexpected IncrementQuotaUsed call")
}
func (s *apiKeyPackageScopeRepoStub) UpdateLastUsed(ctx context.Context, id int64, usedAt time.Time) error {
	panic("unexpected UpdateLastUsed call")
}
func (s *apiKeyPackageScopeRepoStub) IncrementRateLimitUsage(ctx context.Context, id int64, cost float64) error {
	panic("unexpected IncrementRateLimitUsage call")
}
func (s *apiKeyPackageScopeRepoStub) ResetRateLimitWindows(ctx context.Context, id int64) error {
	panic("unexpected ResetRateLimitWindows call")
}
func (s *apiKeyPackageScopeRepoStub) GetRateLimitData(ctx context.Context, id int64) (*APIKeyRateLimitData, error) {
	panic("unexpected GetRateLimitData call")
}

type apiKeyPackageScopeGroupRepoStub struct {
	byID map[int64]*Group
	active []Group
}

func (s *apiKeyPackageScopeGroupRepoStub) Create(ctx context.Context, group *Group) error { panic("unexpected Create call") }
func (s *apiKeyPackageScopeGroupRepoStub) GetByID(ctx context.Context, id int64) (*Group, error) {
	if g, ok := s.byID[id]; ok {
		return g, nil
	}
	return nil, ErrGroupNotFound
}
func (s *apiKeyPackageScopeGroupRepoStub) GetByIDLite(ctx context.Context, id int64) (*Group, error) { return s.GetByID(ctx, id) }
func (s *apiKeyPackageScopeGroupRepoStub) Update(ctx context.Context, group *Group) error { panic("unexpected Update call") }
func (s *apiKeyPackageScopeGroupRepoStub) Delete(ctx context.Context, id int64) error { panic("unexpected Delete call") }
func (s *apiKeyPackageScopeGroupRepoStub) DeleteCascade(ctx context.Context, id int64) ([]int64, error) { panic("unexpected DeleteCascade call") }
func (s *apiKeyPackageScopeGroupRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (s *apiKeyPackageScopeGroupRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, platform, status, search string, isExclusive *bool) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}
func (s *apiKeyPackageScopeGroupRepoStub) ListActive(ctx context.Context) ([]Group, error) { return s.active, nil }
func (s *apiKeyPackageScopeGroupRepoStub) ListActiveByPlatform(ctx context.Context, platform string) ([]Group, error) {
	panic("unexpected ListActiveByPlatform call")
}
func (s *apiKeyPackageScopeGroupRepoStub) ExistsByName(ctx context.Context, name string) (bool, error) { panic("unexpected ExistsByName call") }
func (s *apiKeyPackageScopeGroupRepoStub) GetAccountCount(ctx context.Context, groupID int64) (int64, int64, error) {
	panic("unexpected GetAccountCount call")
}
func (s *apiKeyPackageScopeGroupRepoStub) DeleteAccountGroupsByGroupID(ctx context.Context, groupID int64) (int64, error) {
	panic("unexpected DeleteAccountGroupsByGroupID call")
}
func (s *apiKeyPackageScopeGroupRepoStub) GetAccountIDsByGroupIDs(ctx context.Context, groupIDs []int64) ([]int64, error) {
	panic("unexpected GetAccountIDsByGroupIDs call")
}
func (s *apiKeyPackageScopeGroupRepoStub) BindAccountsToGroup(ctx context.Context, groupID int64, accountIDs []int64) error {
	panic("unexpected BindAccountsToGroup call")
}
func (s *apiKeyPackageScopeGroupRepoStub) UpdateSortOrders(ctx context.Context, updates []GroupSortOrderUpdate) error {
	panic("unexpected UpdateSortOrders call")
}

type apiKeyPackageScopeSubRepoStub struct {
	activeByGroup map[int64]bool
}

func (s *apiKeyPackageScopeSubRepoStub) Create(ctx context.Context, sub *UserSubscription) error { panic("unexpected Create call") }
func (s *apiKeyPackageScopeSubRepoStub) GetByID(ctx context.Context, id int64) (*UserSubscription, error) { panic("unexpected GetByID call") }
func (s *apiKeyPackageScopeSubRepoStub) GetByUserIDAndGroupID(ctx context.Context, userID, groupID int64) (*UserSubscription, error) { panic("unexpected GetByUserIDAndGroupID call") }
func (s *apiKeyPackageScopeSubRepoStub) GetActiveByUserIDAndGroupID(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	if s.activeByGroup[groupID] {
		return &UserSubscription{UserID: userID, GroupID: groupID, Status: SubscriptionStatusActive}, nil
	}
	return nil, ErrSubscriptionNotFound
}
func (s *apiKeyPackageScopeSubRepoStub) Update(ctx context.Context, sub *UserSubscription) error { panic("unexpected Update call") }
func (s *apiKeyPackageScopeSubRepoStub) Delete(ctx context.Context, id int64) error { panic("unexpected Delete call") }
func (s *apiKeyPackageScopeSubRepoStub) ListByUserID(ctx context.Context, userID int64) ([]UserSubscription, error) { panic("unexpected ListByUserID call") }
func (s *apiKeyPackageScopeSubRepoStub) ListActiveByUserID(ctx context.Context, userID int64) ([]UserSubscription, error) {
	var subs []UserSubscription
	for gid, ok := range s.activeByGroup {
		if ok {
			subs = append(subs, UserSubscription{UserID: userID, GroupID: gid, Status: SubscriptionStatusActive})
		}
	}
	return subs, nil
}
func (s *apiKeyPackageScopeSubRepoStub) ListByGroupID(ctx context.Context, groupID int64, params pagination.PaginationParams) ([]UserSubscription, *pagination.PaginationResult, error) {
	panic("unexpected ListByGroupID call")
}
func (s *apiKeyPackageScopeSubRepoStub) List(ctx context.Context, params pagination.PaginationParams, userID, groupID *int64, status, platform, sortBy, sortOrder string) ([]UserSubscription, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (s *apiKeyPackageScopeSubRepoStub) ExistsByUserIDAndGroupID(ctx context.Context, userID, groupID int64) (bool, error) { panic("unexpected ExistsByUserIDAndGroupID call") }
func (s *apiKeyPackageScopeSubRepoStub) ExtendExpiry(ctx context.Context, subscriptionID int64, newExpiresAt time.Time) error { panic("unexpected ExtendExpiry call") }
func (s *apiKeyPackageScopeSubRepoStub) UpdateStatus(ctx context.Context, subscriptionID int64, status string) error { panic("unexpected UpdateStatus call") }
func (s *apiKeyPackageScopeSubRepoStub) UpdateNotes(ctx context.Context, subscriptionID int64, notes string) error { panic("unexpected UpdateNotes call") }
func (s *apiKeyPackageScopeSubRepoStub) ActivateWindows(ctx context.Context, id int64, start time.Time) error { panic("unexpected ActivateWindows call") }
func (s *apiKeyPackageScopeSubRepoStub) ResetDailyUsage(ctx context.Context, id int64, newWindowStart time.Time) error { panic("unexpected ResetDailyUsage call") }
func (s *apiKeyPackageScopeSubRepoStub) ResetWeeklyUsage(ctx context.Context, id int64, newWindowStart time.Time) error { panic("unexpected ResetWeeklyUsage call") }
func (s *apiKeyPackageScopeSubRepoStub) ResetMonthlyUsage(ctx context.Context, id int64, newWindowStart time.Time) error { panic("unexpected ResetMonthlyUsage call") }
func (s *apiKeyPackageScopeSubRepoStub) IncrementUsage(ctx context.Context, id int64, costUSD float64) error { panic("unexpected IncrementUsage call") }
func (s *apiKeyPackageScopeSubRepoStub) BatchUpdateExpiredStatus(ctx context.Context) (int64, error) { panic("unexpected BatchUpdateExpiredStatus call") }

func TestCreateAPIKeyRejectsDifferentPackageScope(t *testing.T) {
	scope := PackageScopeCodex
	repo := &apiKeyPackageScopeRepoStub{}
	groupRepo := &apiKeyPackageScopeGroupRepoStub{
		byID: map[int64]*Group{
			2: {ID: 2, Name: "general", PackageScope: strPtr(PackageScopeGeneral), SubscriptionType: SubscriptionTypeStandard},
		},
	}
	svc := NewAPIKeyService(repo, &packageScopeUserRepoStub{user: &User{ID: 1, Status: StatusActive, PackageScope: &scope}}, groupRepo, &apiKeyPackageScopeSubRepoStub{}, nil, nil, &config.Config{})

	_, err := svc.Create(context.Background(), 1, CreateAPIKeyRequest{
		Name: "k1", GroupID: int64PtrPS(2), CustomKey: strPtrPS("custom_key_123456"),
	})
	require.ErrorIs(t, err, ErrPackageScopeNotAllowed)
}

func TestGetAvailableGroupsFiltersByPackageScopeButKeepsSubscribedGroups(t *testing.T) {
	scope := PackageScopeCodex
	userRepo := &packageScopeUserRepoStub{user: &User{ID: 1, Status: StatusActive, PackageScope: &scope}}
	groupRepo := &apiKeyPackageScopeGroupRepoStub{
		active: []Group{
			{ID: 1, Name: "codex-standard", PackageScope: strPtrPS(PackageScopeCodex), SubscriptionType: SubscriptionTypeStandard},
			{ID: 2, Name: "general-standard", PackageScope: strPtrPS(PackageScopeGeneral), SubscriptionType: SubscriptionTypeStandard},
			{ID: 3, Name: "subscribed-plan", SubscriptionType: SubscriptionTypeSubscription},
		},
	}
	subRepo := &apiKeyPackageScopeSubRepoStub{activeByGroup: map[int64]bool{3: true}}
	svc := NewAPIKeyService(&apiKeyPackageScopeRepoStub{}, userRepo, groupRepo, subRepo, nil, nil, &config.Config{})

	groups, err := svc.GetAvailableGroups(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, groups, 2)
	require.Equal(t, int64(1), groups[0].ID)
	require.Equal(t, int64(3), groups[1].ID)
}

func strPtrPS(v string) *string { return &v }
func int64PtrPS(v int64) *int64 { return &v }
