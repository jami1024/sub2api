package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

type packageScopeUserRepoStub struct {
	user       *User
	updated    []*User
	balanceOps []float64
}

func (s *packageScopeUserRepoStub) Create(ctx context.Context, user *User) error {
	panic("unexpected Create call")
}
func (s *packageScopeUserRepoStub) GetByID(ctx context.Context, id int64) (*User, error) {
	if s.user == nil {
		return nil, ErrUserNotFound
	}
	return s.user, nil
}
func (s *packageScopeUserRepoStub) GetByEmail(ctx context.Context, email string) (*User, error) {
	panic("unexpected GetByEmail call")
}
func (s *packageScopeUserRepoStub) GetFirstAdmin(ctx context.Context) (*User, error) {
	panic("unexpected GetFirstAdmin call")
}
func (s *packageScopeUserRepoStub) Update(ctx context.Context, user *User) error {
	clone := *user
	s.updated = append(s.updated, &clone)
	s.user = &clone
	return nil
}
func (s *packageScopeUserRepoStub) Delete(ctx context.Context, id int64) error {
	panic("unexpected Delete call")
}
func (s *packageScopeUserRepoStub) GetUserAvatar(ctx context.Context, userID int64) (*UserAvatar, error) {
	panic("unexpected GetUserAvatar call")
}
func (s *packageScopeUserRepoStub) UpsertUserAvatar(ctx context.Context, userID int64, input UpsertUserAvatarInput) (*UserAvatar, error) {
	panic("unexpected UpsertUserAvatar call")
}
func (s *packageScopeUserRepoStub) DeleteUserAvatar(ctx context.Context, userID int64) error {
	panic("unexpected DeleteUserAvatar call")
}
func (s *packageScopeUserRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (s *packageScopeUserRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, filters UserListFilters) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}
func (s *packageScopeUserRepoStub) GetLatestUsedAtByUserIDs(ctx context.Context, userIDs []int64) (map[int64]*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserIDs call")
}
func (s *packageScopeUserRepoStub) GetLatestUsedAtByUserID(ctx context.Context, userID int64) (*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserID call")
}
func (s *packageScopeUserRepoStub) UpdateUserLastActiveAt(ctx context.Context, userID int64, activeAt time.Time) error {
	panic("unexpected UpdateUserLastActiveAt call")
}
func (s *packageScopeUserRepoStub) UpdateBalance(ctx context.Context, id int64, amount float64) error {
	s.balanceOps = append(s.balanceOps, amount)
	if s.user == nil {
		return ErrUserNotFound
	}
	s.user.Balance += amount
	return nil
}
func (s *packageScopeUserRepoStub) DeductBalance(ctx context.Context, id int64, amount float64) error {
	panic("unexpected DeductBalance call")
}
func (s *packageScopeUserRepoStub) UpdateConcurrency(ctx context.Context, id int64, amount int) error {
	panic("unexpected UpdateConcurrency call")
}
func (s *packageScopeUserRepoStub) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	panic("unexpected ExistsByEmail call")
}
func (s *packageScopeUserRepoStub) RemoveGroupFromAllowedGroups(ctx context.Context, groupID int64) (int64, error) {
	panic("unexpected RemoveGroupFromAllowedGroups call")
}
func (s *packageScopeUserRepoStub) AddGroupToAllowedGroups(ctx context.Context, userID int64, groupID int64) error {
	panic("unexpected AddGroupToAllowedGroups call")
}
func (s *packageScopeUserRepoStub) RemoveGroupFromUserAllowedGroups(ctx context.Context, userID int64, groupID int64) error {
	panic("unexpected RemoveGroupFromUserAllowedGroups call")
}
func (s *packageScopeUserRepoStub) ListUserAuthIdentities(ctx context.Context, userID int64) ([]UserAuthIdentityRecord, error) {
	panic("unexpected ListUserAuthIdentities call")
}
func (s *packageScopeUserRepoStub) UnbindUserAuthProvider(ctx context.Context, userID int64, provider string) error {
	panic("unexpected UnbindUserAuthProvider call")
}
func (s *packageScopeUserRepoStub) UpdateTotpSecret(ctx context.Context, userID int64, encryptedSecret *string) error {
	panic("unexpected UpdateTotpSecret call")
}
func (s *packageScopeUserRepoStub) EnableTotp(ctx context.Context, userID int64) error {
	panic("unexpected EnableTotp call")
}
func (s *packageScopeUserRepoStub) DisableTotp(ctx context.Context, userID int64) error {
	panic("unexpected DisableTotp call")
}

type packageScopeRedeemRepoStub struct {
	code *RedeemCode
}

func (s *packageScopeRedeemRepoStub) Create(ctx context.Context, code *RedeemCode) error {
	panic("unexpected Create call")
}
func (s *packageScopeRedeemRepoStub) CreateBatch(ctx context.Context, codes []RedeemCode) error {
	panic("unexpected CreateBatch call")
}
func (s *packageScopeRedeemRepoStub) GetByID(ctx context.Context, id int64) (*RedeemCode, error) {
	if s.code == nil {
		return nil, ErrRedeemCodeNotFound
	}
	return s.code, nil
}
func (s *packageScopeRedeemRepoStub) GetByCode(ctx context.Context, code string) (*RedeemCode, error) {
	if s.code == nil || s.code.Code != code {
		return nil, ErrRedeemCodeNotFound
	}
	return s.code, nil
}
func (s *packageScopeRedeemRepoStub) Update(ctx context.Context, code *RedeemCode) error {
	panic("unexpected Update call")
}
func (s *packageScopeRedeemRepoStub) Delete(ctx context.Context, id int64) error {
	panic("unexpected Delete call")
}
func (s *packageScopeRedeemRepoStub) Use(ctx context.Context, id, userID int64) error {
	if s.code == nil || s.code.ID != id {
		return ErrRedeemCodeNotFound
	}
	now := time.Now()
	s.code.Status = StatusUsed
	s.code.UsedBy = &userID
	s.code.UsedAt = &now
	return nil
}
func (s *packageScopeRedeemRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (s *packageScopeRedeemRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, codeType, status, search string) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}
func (s *packageScopeRedeemRepoStub) ListByUser(ctx context.Context, userID int64, limit int) ([]RedeemCode, error) {
	panic("unexpected ListByUser call")
}
func (s *packageScopeRedeemRepoStub) ListByUserPaginated(ctx context.Context, userID int64, params pagination.PaginationParams, codeType string) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected ListByUserPaginated call")
}
func (s *packageScopeRedeemRepoStub) SumPositiveBalanceByUser(ctx context.Context, userID int64) (float64, error) {
	panic("unexpected SumPositiveBalanceByUser call")
}

type packageScopeAdminRedeemRepoStub struct {
	created []*RedeemCode
}

func (s *packageScopeAdminRedeemRepoStub) Create(ctx context.Context, code *RedeemCode) error {
	if code == nil {
		return nil
	}
	clone := *code
	s.created = append(s.created, &clone)
	return nil
}
func (s *packageScopeAdminRedeemRepoStub) CreateBatch(ctx context.Context, codes []RedeemCode) error {
	panic("unexpected CreateBatch call")
}
func (s *packageScopeAdminRedeemRepoStub) GetByID(ctx context.Context, id int64) (*RedeemCode, error) {
	panic("unexpected GetByID call")
}
func (s *packageScopeAdminRedeemRepoStub) GetByCode(ctx context.Context, code string) (*RedeemCode, error) {
	panic("unexpected GetByCode call")
}
func (s *packageScopeAdminRedeemRepoStub) Update(ctx context.Context, code *RedeemCode) error {
	panic("unexpected Update call")
}
func (s *packageScopeAdminRedeemRepoStub) Delete(ctx context.Context, id int64) error {
	panic("unexpected Delete call")
}
func (s *packageScopeAdminRedeemRepoStub) Use(ctx context.Context, id, userID int64) error {
	panic("unexpected Use call")
}
func (s *packageScopeAdminRedeemRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (s *packageScopeAdminRedeemRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, codeType, status, search string) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}
func (s *packageScopeAdminRedeemRepoStub) ListByUser(ctx context.Context, userID int64, limit int) ([]RedeemCode, error) {
	panic("unexpected ListByUser call")
}
func (s *packageScopeAdminRedeemRepoStub) ListByUserPaginated(ctx context.Context, userID int64, params pagination.PaginationParams, codeType string) ([]RedeemCode, *pagination.PaginationResult, error) {
	panic("unexpected ListByUserPaginated call")
}
func (s *packageScopeAdminRedeemRepoStub) SumPositiveBalanceByUser(ctx context.Context, userID int64) (float64, error) {
	panic("unexpected SumPositiveBalanceByUser call")
}

type packageScopeAuthCacheInvalidatorStub struct{}

func (s *packageScopeAuthCacheInvalidatorStub) InvalidateAuthCacheByKey(ctx context.Context, key string) {
}
func (s *packageScopeAuthCacheInvalidatorStub) InvalidateAuthCacheByUserID(ctx context.Context, userID int64) {
}
func (s *packageScopeAuthCacheInvalidatorStub) InvalidateAuthCacheByGroupID(ctx context.Context, groupID int64) {
}

type packageScopeAffiliateRepoStub struct {
	summaries map[int64]*AffiliateSummary
	records   []AffiliateRebateRecordInput
}

func (s *packageScopeAffiliateRepoStub) EnsureUserAffiliate(ctx context.Context, userID int64) (*AffiliateSummary, error) {
	if summary, ok := s.summaries[userID]; ok {
		return summary, nil
	}
	return nil, ErrAffiliateProfileNotFound
}
func (s *packageScopeAffiliateRepoStub) GetAffiliateByCode(ctx context.Context, code string) (*AffiliateSummary, error) {
	panic("unexpected GetAffiliateByCode call")
}
func (s *packageScopeAffiliateRepoStub) BindInviter(ctx context.Context, userID, inviterID int64) (bool, error) {
	panic("unexpected BindInviter call")
}
func (s *packageScopeAffiliateRepoStub) AccrueQuota(ctx context.Context, inviterID, inviteeUserID int64, amount float64) (bool, error) {
	panic("unexpected AccrueQuota call")
}
func (s *packageScopeAffiliateRepoStub) TransferQuotaToBalance(ctx context.Context, userID int64) (float64, float64, error) {
	panic("unexpected TransferQuotaToBalance call")
}
func (s *packageScopeAffiliateRepoStub) ListInvitees(ctx context.Context, inviterID int64, limit int) ([]AffiliateInvitee, error) {
	panic("unexpected ListInvitees call")
}
func (s *packageScopeAffiliateRepoStub) ListAncestors(ctx context.Context, userID int64, maxDepth int) ([]AffiliateAncestor, error) {
	current := userID
	out := make([]AffiliateAncestor, 0, maxDepth)
	visited := map[int64]struct{}{userID: {}}
	for level := 1; level <= maxDepth; level++ {
		summary, ok := s.summaries[current]
		if !ok || summary.InviterID == nil {
			break
		}
		inviterID := *summary.InviterID
		if _, seen := visited[inviterID]; seen {
			break
		}
		visited[inviterID] = struct{}{}
		out = append(out, AffiliateAncestor{UserID: inviterID, Level: level})
		current = inviterID
	}
	return out, nil
}
func (s *packageScopeAffiliateRepoStub) CreatePendingRebateRecords(ctx context.Context, records []AffiliateRebateRecordInput) (int, error) {
	s.records = append(s.records, records...)
	return len(records), nil
}
func (s *packageScopeAffiliateRepoStub) ReleaseDuePendingRebateRecords(ctx context.Context, now time.Time) (int, error) {
	panic("unexpected ReleaseDuePendingRebateRecords call")
}
func (s *packageScopeAffiliateRepoStub) CreateWithdrawalRequest(ctx context.Context, userID int64, amount float64, applicantNote string) (*AffiliateWithdrawalRequest, error) {
	panic("unexpected CreateWithdrawalRequest call")
}
func (s *packageScopeAffiliateRepoStub) ListWithdrawalRequests(ctx context.Context, status string, limit int) ([]AffiliateWithdrawalRequest, error) {
	panic("unexpected ListWithdrawalRequests call")
}
func (s *packageScopeAffiliateRepoStub) ListUserWithdrawalRequests(ctx context.Context, userID int64, limit int) ([]AffiliateWithdrawalRequest, error) {
	panic("unexpected ListUserWithdrawalRequests call")
}
func (s *packageScopeAffiliateRepoStub) RejectWithdrawalRequest(ctx context.Context, requestID int64, reviewerID int64, adminNote string) (*AffiliateWithdrawalRequest, error) {
	panic("unexpected RejectWithdrawalRequest call")
}
func (s *packageScopeAffiliateRepoStub) MarkWithdrawalPaid(ctx context.Context, requestID int64, reviewerID int64, adminNote string) (*AffiliateWithdrawalRequest, error) {
	panic("unexpected MarkWithdrawalPaid call")
}
func (s *packageScopeAffiliateRepoStub) ReverseRebatesForOrder(ctx context.Context, sourceOrderID int64) error {
	panic("unexpected ReverseRebatesForOrder call")
}
func (s *packageScopeAffiliateRepoStub) SumPendingRebateByUser(ctx context.Context, userID int64) (float64, error) {
	panic("unexpected SumPendingRebateByUser call")
}
func (s *packageScopeAffiliateRepoStub) ListUserRebateRecords(ctx context.Context, userID int64, limit int) ([]AffiliateRebateRecord, error) {
	panic("unexpected ListUserRebateRecords call")
}

func newPackageScopeEntClient(t *testing.T) *dbent.Client {
	t.Helper()
	db, err := sql.Open("sqlite", "file:package_scope_service?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func seedUserWithBalance(t *testing.T, balance float64) *User {
	t.Helper()
	return &User{ID: 1, Balance: balance}
}

func seedUserWithNilScopeAndZeroBalance(t *testing.T) *User {
	t.Helper()
	return &User{ID: 7, Balance: 0, PackageScope: nil}
}

func TestPackageScopeMatchesGroup_ExactMatchOnly(t *testing.T) {
	require.True(t, PackageScopeMatchesGroup("codex", "codex"))
	require.True(t, PackageScopeMatchesGroup("general", "general"))
	require.False(t, PackageScopeMatchesGroup("codex", "general"))
	require.False(t, PackageScopeMatchesGroup("general", "codex"))
	require.False(t, PackageScopeMatchesGroup("codex", ""))
}

func TestBackfillScopeForExistingPositiveBalanceUser(t *testing.T) {
	user := seedUserWithBalance(t, 100)
	scope := ResolveBackfillPackageScope(user)
	require.Equal(t, PackageScopeCodex, scope)
}

func TestRedeemBalanceDefaultsScopeToCodex(t *testing.T) {
	repo := &packageScopeUserRepoStub{user: seedUserWithNilScopeAndZeroBalance(t)}
	redeemRepo := &packageScopeRedeemRepoStub{code: &RedeemCode{
		ID:     1,
		Code:   "CODE-123",
		Type:   RedeemTypeBalance,
		Value:  30,
		Status: StatusUnused,
	}}
	svc := NewRedeemService(redeemRepo, repo, nil, nil, nil, newPackageScopeEntClient(t), &packageScopeAuthCacheInvalidatorStub{})

	_, err := svc.Redeem(context.Background(), repo.user.ID, "CODE-123")
	require.NoError(t, err)
	updated, err := repo.GetByID(context.Background(), repo.user.ID)
	require.NoError(t, err)
	require.Equal(t, PackageScopeCodex, psStringValue(updated.PackageScope))
}

func TestAdminAdjustBalanceDefaultsScopeToCodex(t *testing.T) {
	repo := &packageScopeUserRepoStub{user: seedUserWithNilScopeAndZeroBalance(t)}
	redeemRepo := &packageScopeAdminRedeemRepoStub{}
	svc := &adminServiceImpl{
		userRepo:             repo,
		redeemCodeRepo:       redeemRepo,
		authCacheInvalidator: &packageScopeAuthCacheInvalidatorStub{},
		billingCacheService:  &BillingCacheService{},
	}

	_, err := svc.UpdateUserBalance(context.Background(), repo.user.ID, 50, "add", "seed")
	require.NoError(t, err)
	updated, err := repo.GetByID(context.Background(), repo.user.ID)
	require.NoError(t, err)
	require.Equal(t, PackageScopeCodex, psStringValue(updated.PackageScope))
}

func TestCreateAndListBalancePackages(t *testing.T) {
	client := newPackageScopeEntClient(t)
	svc := NewPaymentConfigService(client, &paymentConfigSettingRepoStub{values: map[string]string{}}, nil)

	created, err := svc.CreateBalancePackage(context.Background(), CreateBalancePackageRequest{
		Name:         "Codex 100 包",
		Price:        100,
		CreditAmount: 100,
		PackageScope: PackageScopeCodex,
		ForSale:      true,
	})
	require.NoError(t, err)

	items, err := svc.ListBalancePackages(context.Background(), true)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, created.ID, items[0].ID)
	require.Equal(t, PackageScopeCodex, items[0].PackageScope)
}

func ptrStringSlice(v []string) *[]string {
	return &v
}

func TestCreateBalancePackageOrderStoresSnapshot(t *testing.T) {
	client := newPackageScopeEntClient(t)
	createdUser, err := client.User.Create().
		SetEmail("u@example.com").
		SetPasswordHash("hash").
		SetRole(RoleUser).
		SetBalance(0).
		SetConcurrency(1).
		SetStatus(StatusActive).
		Save(context.Background())
	require.NoError(t, err)

	pkg, err := client.BalancePackage.Create().
		SetName("Codex 100 包").
		SetDescription("codex only").
		SetPrice(100).
		SetCreditAmount(100).
		SetPackageScope(PackageScopeCodex).
		SetForSale(true).
		Save(context.Background())
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}
	order, err := svc.createOrderInTx(context.Background(), CreateOrderRequest{
		UserID:           createdUser.ID,
		PaymentType:      "alipay",
		OrderType:        payment.OrderTypeBalancePackage,
		BalancePackageID: pkg.ID,
		ClientIP:         "127.0.0.1",
		SrcHost:          "example.com",
	}, &User{
		ID:       createdUser.ID,
		Email:    createdUser.Email,
		Username: createdUser.Username,
	}, nil, pkg, &PaymentConfig{MaxPendingOrders: 3, OrderTimeoutMin: 30}, pkg.CreditAmount, pkg.Price, 0, pkg.Price, nil)
	require.NoError(t, err)
	require.NotNil(t, order.BalancePackageID)
	require.Equal(t, pkg.ID, *order.BalancePackageID)
	require.Equal(t, PackageScopeCodex, order.PackageScopeSnapshot)
	require.False(t, order.ForceSwitchScope)
	require.Equal(t, pkg.CreditAmount, order.Amount)
	require.Equal(t, pkg.Price, order.PayAmount)
}

func TestCreateBalancePackage_PersistsDisplayTags(t *testing.T) {
	client := newPackageScopeEntClient(t)
	svc := NewPaymentConfigService(client, &paymentConfigSettingRepoStub{values: map[string]string{}}, nil)

	created, err := svc.CreateBalancePackage(context.Background(), CreateBalancePackageRequest{
		Name:         "Codex 100 包",
		Price:        100,
		CreditAmount: 100,
		PackageScope: PackageScopeCodex,
		DisplayTags:  []string{"新手推荐", "1x 倍率", "适合 Codex"},
		ForSale:      true,
	})
	require.NoError(t, err)
	require.Equal(t, []string{"新手推荐", "1x 倍率", "适合 Codex"}, created.DisplayTags)
}

func TestUpdateBalancePackage_PersistsDisplayTags(t *testing.T) {
	client := newPackageScopeEntClient(t)
	svc := NewPaymentConfigService(client, &paymentConfigSettingRepoStub{values: map[string]string{}}, nil)

	pkg, err := client.BalancePackage.Create().
		SetName("Codex 100 包").
		SetDescription("codex only").
		SetPrice(100).
		SetCreditAmount(100).
		SetPackageScope(PackageScopeCodex).
		SetForSale(true).
		Save(context.Background())
	require.NoError(t, err)

	updated, err := svc.UpdateBalancePackage(context.Background(), pkg.ID, UpdateBalancePackageRequest{
		DisplayTags: ptrStringSlice([]string{"高频使用", "适合编程"}),
	})
	require.NoError(t, err)
	require.Equal(t, []string{"高频使用", "适合编程"}, updated.DisplayTags)
}

func TestCreateBalancePackage_RejectsMoreThanThreeDisplayTags(t *testing.T) {
	client := newPackageScopeEntClient(t)
	svc := NewPaymentConfigService(client, &paymentConfigSettingRepoStub{values: map[string]string{}}, nil)

	_, err := svc.CreateBalancePackage(context.Background(), CreateBalancePackageRequest{
		Name:         "Codex 100 包",
		Price:        100,
		CreditAmount: 100,
		PackageScope: PackageScopeCodex,
		DisplayTags:  []string{"新手", "1x", "编程", "多一个"},
		ForSale:      true,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "BALANCE_PACKAGE_DISPLAY_TAGS_INVALID")
}

func TestCreateBalancePackage_NormalizesDisplayTags(t *testing.T) {
	client := newPackageScopeEntClient(t)
	svc := NewPaymentConfigService(client, &paymentConfigSettingRepoStub{values: map[string]string{}}, nil)

	created, err := svc.CreateBalancePackage(context.Background(), CreateBalancePackageRequest{
		Name:         "Codex 100 包",
		Price:        100,
		CreditAmount: 100,
		PackageScope: PackageScopeCodex,
		DisplayTags:  []string{" 新手推荐 ", "", "新手推荐", "1x 倍率"},
		ForSale:      true,
	})
	require.NoError(t, err)
	require.Equal(t, []string{"新手推荐", "1x 倍率"}, created.DisplayTags)
}

func TestCreateBalancePackageOrder_AllowsDifferentScopeWithForceSwitch(t *testing.T) {
	client := newPackageScopeEntClient(t)
	createdUser, err := client.User.Create().
		SetEmail("force-switch-user@example.com").
		SetPasswordHash("hash").
		SetRole(RoleUser).
		SetBalance(15).
		SetConcurrency(1).
		SetStatus(StatusActive).
		SetPackageScope(PackageScopeCodex).
		Save(context.Background())
	require.NoError(t, err)

	pkg, err := client.BalancePackage.Create().
		SetName("General 包").
		SetDescription("general only").
		SetPrice(88).
		SetCreditAmount(80).
		SetPackageScope(PackageScopeGeneral).
		SetForSale(true).
		Save(context.Background())
	require.NoError(t, err)

	codex := PackageScopeCodex
	svc := &PaymentService{entClient: client}
	order, err := svc.createOrderInTx(context.Background(), CreateOrderRequest{
		UserID:           createdUser.ID,
		PaymentType:      "alipay",
		OrderType:        payment.OrderTypeBalancePackage,
		BalancePackageID: pkg.ID,
		ForceSwitchScope: true,
		ClientIP:         "127.0.0.1",
		SrcHost:          "example.com",
	}, &User{
		ID:           createdUser.ID,
		Email:        createdUser.Email,
		Username:     createdUser.Username,
		Balance:      15,
		PackageScope: &codex,
	}, nil, pkg, &PaymentConfig{MaxPendingOrders: 3, OrderTimeoutMin: 30}, pkg.CreditAmount, pkg.Price, 0, pkg.Price, nil)
	require.NoError(t, err)
	require.True(t, order.ForceSwitchScope)
}

func TestExecuteBalancePackageFulfillmentCreditsUserBalance(t *testing.T) {
	ctx := context.Background()
	client := newPackageScopeEntClient(t)

	userEntity, err := client.User.Create().
		SetEmail("balance-package-user@example.com").
		SetPasswordHash("hash").
		SetUsername("bp-user").
		SetBalance(0).
		SetConcurrency(1).
		SetStatus(StatusActive).
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(userEntity.ID).
		SetUserEmail(userEntity.Email).
		SetUserName(userEntity.Username).
		SetAmount(100).
		SetPayAmount(100).
		SetFeeRate(0).
		SetRechargeCode("PAY-BP-1").
		SetOutTradeNo("sub2_bp_success").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-bp-1").
		SetOrderType(payment.OrderTypeBalancePackage).
		SetBalancePackageID(1).
		SetPackageScopeSnapshot(PackageScopeCodex).
		SetStatus(OrderStatusPaid).
		SetPaidAt(time.Now()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("example.com").
		Save(ctx)
	require.NoError(t, err)

	userRepo := &packageScopeUserRepoStub{user: &User{
		ID:           userEntity.ID,
		Email:        userEntity.Email,
		Username:     userEntity.Username,
		Balance:      0,
		PackageScope: nil,
		Status:       StatusActive,
	}}
	inviter1, inviter2, inviter3 := int64(101), int64(102), int64(103)
	affiliateRepo := &packageScopeAffiliateRepoStub{
		summaries: map[int64]*AffiliateSummary{
			userEntity.ID: {UserID: userEntity.ID, InviterID: &inviter1},
			inviter1:      {UserID: inviter1, InviterID: &inviter2},
			inviter2:      {UserID: inviter2, InviterID: &inviter3},
			inviter3:      {UserID: inviter3},
		},
	}
	svc := &PaymentService{
		entClient:        client,
		userRepo:         userRepo,
		affiliateService: &AffiliateService{repo: affiliateRepo},
	}

	err = svc.ExecuteBalancePackageFulfillment(ctx, order.ID)
	require.NoError(t, err)

	updated, err := userRepo.GetByID(ctx, userEntity.ID)
	require.NoError(t, err)
	require.Equal(t, 100.0, updated.Balance)
	require.Equal(t, PackageScopeCodex, psStringValue(updated.PackageScope))

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, reloaded.Status)
	require.Len(t, affiliateRepo.records, 3)
	require.InDelta(t, 6.0, affiliateRepo.records[0].RebateAmount, 1e-9)
	require.InDelta(t, 3.0, affiliateRepo.records[1].RebateAmount, 1e-9)
	require.InDelta(t, 1.0, affiliateRepo.records[2].RebateAmount, 1e-9)
}

func TestExecuteBalancePackageFulfillmentMarksFailedAfterPaidConflict(t *testing.T) {
	ctx := context.Background()
	client := newPackageScopeEntClient(t)

	userEntity, err := client.User.Create().
		SetEmail("balance-package-conflict@example.com").
		SetPasswordHash("hash").
		SetUsername("bp-conflict").
		SetBalance(20).
		SetConcurrency(1).
		SetStatus(StatusActive).
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(userEntity.ID).
		SetUserEmail(userEntity.Email).
		SetUserName(userEntity.Username).
		SetAmount(100).
		SetPayAmount(100).
		SetFeeRate(0).
		SetRechargeCode("PAY-BP-2").
		SetOutTradeNo("sub2_bp_conflict").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-bp-2").
		SetOrderType(payment.OrderTypeBalancePackage).
		SetBalancePackageID(2).
		SetPackageScopeSnapshot(PackageScopeGeneral).
		SetStatus(OrderStatusPaid).
		SetPaidAt(time.Now()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("example.com").
		Save(ctx)
	require.NoError(t, err)

	existingScope := PackageScopeCodex
	userRepo := &packageScopeUserRepoStub{user: &User{
		ID:           userEntity.ID,
		Email:        userEntity.Email,
		Username:     userEntity.Username,
		Balance:      20,
		PackageScope: &existingScope,
		Status:       StatusActive,
	}}
	svc := &PaymentService{entClient: client, userRepo: userRepo}

	err = svc.ExecuteBalancePackageFulfillment(ctx, order.ID)
	require.Error(t, err)

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusFailed, reloaded.Status)
	require.Contains(t, psStringValue(reloaded.FailedReason), "PACKAGE_SCOPE_CONFLICT_AFTER_PAYMENT")
}

func TestExecuteBalancePackageFulfillment_ForceSwitchClearsOldBalanceAndSwitchesScope(t *testing.T) {
	ctx := context.Background()
	client := newPackageScopeEntClient(t)

	userEntity, err := client.User.Create().
		SetEmail("balance-package-force@example.com").
		SetPasswordHash("hash").
		SetUsername("bp-force").
		SetBalance(35).
		SetConcurrency(1).
		SetStatus(StatusActive).
		SetPackageScope(PackageScopeCodex).
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(userEntity.ID).
		SetUserEmail(userEntity.Email).
		SetUserName(userEntity.Username).
		SetAmount(80).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("PAY-BP-FORCE").
		SetOutTradeNo("sub2_bp_force_switch").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-bp-force").
		SetOrderType(payment.OrderTypeBalancePackage).
		SetBalancePackageID(3).
		SetPackageScopeSnapshot(PackageScopeGeneral).
		SetForceSwitchScope(true).
		SetStatus(OrderStatusPaid).
		SetPaidAt(time.Now()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("example.com").
		Save(ctx)
	require.NoError(t, err)

	existingScope := PackageScopeCodex
	userRepo := &packageScopeUserRepoStub{user: &User{
		ID:           userEntity.ID,
		Email:        userEntity.Email,
		Username:     userEntity.Username,
		Balance:      35,
		PackageScope: &existingScope,
		Status:       StatusActive,
	}}
	svc := &PaymentService{entClient: client, userRepo: userRepo}

	err = svc.ExecuteBalancePackageFulfillment(ctx, order.ID)
	require.NoError(t, err)

	updated, err := userRepo.GetByID(ctx, userEntity.ID)
	require.NoError(t, err)
	require.Equal(t, 80.0, updated.Balance)
	require.Equal(t, PackageScopeGeneral, psStringValue(updated.PackageScope))
	require.Equal(t, []float64{-35, 80}, userRepo.balanceOps)

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, reloaded.Status)
}
