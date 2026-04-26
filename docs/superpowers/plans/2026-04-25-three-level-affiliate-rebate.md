# Three-Level Affiliate Rebate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为余额包订单实现三层邀请返利、T+7 冻结转可提现、满 100 人工提现申请、退款冲正与负债抵扣。

**Architecture:** 保留现有注册绑定邀请人主链路，新增“返利明细账本 + 提现申请 + 负债抵扣”三块核心数据模型。支付成功后按订单生成多条返利记录，冻结到期后解锁；退款根据返利状态做取消、冲正或负债化。用户侧仅展示聚合结果与提现申请入口，后台负责审核与打款确认。

**Tech Stack:** Go、Ent、PostgreSQL、Gin、Vue 3、TypeScript、Vitest、Go testing

---

## File Map

### Backend schema / migration
- Create: `backend/migrations/131_add_three_level_affiliate_rebates.sql`
- Modify: `backend/ent/schema/user_affiliate.go` *(如已有 schema 文件名不同，按实际 ent schema 调整)*
- Create: `backend/ent/schema/affiliate_rebate_record.go`
- Create: `backend/ent/schema/affiliate_withdrawal_request.go`
- Create: `backend/ent/schema/affiliate_balance_snapshot.go` *(如果采用独立聚合表；若直接在 user_affiliates 扩字段，则改现有 schema)*

### Backend service / repository / handler
- Modify: `backend/internal/repository/affiliate_repo.go`
- Modify: `backend/internal/service/affiliate_service.go`
- Modify: `backend/internal/service/payment_fulfillment.go`
- Modify: `backend/internal/service/payment_refund.go`
- Create: `backend/internal/service/affiliate_rebate_release_job.go`
- Create: `backend/internal/handler/admin/affiliate_handler.go`
- Modify: `backend/internal/handler/user_handler.go`
- Modify: `backend/internal/server/routes/user.go`
- Modify: `backend/internal/server/routes/admin.go` *(或实际 admin 路由注册文件)*
- Modify: `backend/internal/handler/dto/settings.go` *(如需透出固定规则文案或门槛配置)*

### Frontend
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/api/user.ts`
- Create: `frontend/src/api/admin/affiliate.ts`
- Modify: `frontend/src/views/user/AffiliateView.vue`
- Create: `frontend/src/views/admin/AffiliateWithdrawalsView.vue`
- Modify: `frontend/src/router/index.ts`
- Modify: `frontend/src/i18n/locales/zh.ts`
- Modify: `frontend/src/i18n/locales/en.ts`

### Tests
- Create: `backend/internal/service/affiliate_multi_level_rebate_test.go`
- Create: `backend/internal/service/affiliate_withdrawal_test.go`
- Create: `backend/internal/service/affiliate_refund_reversal_test.go`
- Modify: `backend/internal/service/payment_balance_package_test.go`
- Modify: `backend/internal/repository/affiliate_repo_integration_test.go`
- Create: `frontend/src/views/user/__tests__/AffiliateView.rebate.spec.ts`
- Create: `frontend/src/views/admin/__tests__/AffiliateWithdrawalsView.spec.ts`

---

### Task 1: 建立三层返利与提现的数据模型

**Files:**
- Create: `backend/migrations/131_add_three_level_affiliate_rebates.sql`
- Create: `backend/ent/schema/affiliate_rebate_record.go`
- Create: `backend/ent/schema/affiliate_withdrawal_request.go`
- Modify: `backend/ent/schema/user_affiliate.go`
- Test: `backend/internal/repository/affiliate_repo_integration_test.go`

- [ ] **Step 1: 写失败测试，明确新账本字段与状态约束**

```go
func TestAffiliateRepository_CreatePendingRebateRecords(t *testing.T) {
    t.Skip("RED: implement affiliate rebate record persistence")
}
```

Run: `go test ./backend/internal/repository -run TestAffiliateRepository_CreatePendingRebateRecords`
Expected: FAIL，因为返利记录模型和持久化接口尚不存在。

- [ ] **Step 2: 新增数据库迁移，定义返利记录与提现申请表**

```sql
CREATE TABLE affiliate_rebate_records (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  source_user_id BIGINT NOT NULL,
  source_order_id BIGINT NOT NULL,
  level SMALLINT NOT NULL,
  rate NUMERIC(10,4) NOT NULL,
  base_amount NUMERIC(20,8) NOT NULL,
  rebate_amount NUMERIC(20,8) NOT NULL,
  status VARCHAR(32) NOT NULL,
  available_at TIMESTAMPTZ NOT NULL,
  reversed_amount NUMERIC(20,8) NOT NULL DEFAULT 0,
  debt_amount NUMERIC(20,8) NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE affiliate_withdrawal_requests (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  amount NUMERIC(20,8) NOT NULL,
  status VARCHAR(32) NOT NULL,
  applicant_note TEXT NOT NULL DEFAULT '',
  admin_note TEXT NOT NULL DEFAULT '',
  reviewed_by BIGINT,
  reviewed_at TIMESTAMPTZ,
  paid_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- [ ] **Step 3: 在 Ent schema 中映射状态与索引**

```go
field.String("status").MaxLen(32)
field.Time("available_at")
index.Fields("user_id", "status")
index.Fields("source_order_id")
index.Fields("user_id", "level")
```

- [ ] **Step 4: 为 `user_affiliates` 增补新的聚合字段（冻结、可提现、提现中、已提现、负债）或明确由查询实时聚合**

```go
field.Float("pending_quota").Default(0)
field.Float("available_quota").Default(0)
field.Float("withdrawing_quota").Default(0)
field.Float("withdrawn_history_quota").Default(0)
field.Float("debt_quota").Default(0)
```

- [ ] **Step 5: 运行 ent 生成与相关仓储测试**

Run: `make -C backend test`
Expected: 新 schema 可编译，旧 affiliate repo 测试因未实现新逻辑仍可能失败，但迁移与 ent 生成通过。

---

### Task 2: 把单层邀请关系扩展为最多三层追溯

**Files:**
- Modify: `backend/internal/repository/affiliate_repo.go`
- Modify: `backend/internal/service/affiliate_service.go`
- Test: `backend/internal/service/affiliate_multi_level_rebate_test.go`

- [ ] **Step 1: 写失败测试，验证能从 D 追溯到 C/B/A 三层**

```go
func TestResolveAffiliateAncestors_MaxThreeLevels(t *testing.T) {
    t.Skip("RED: should resolve inviter chain depth <= 3")
}
```

Run: `go test ./backend/internal/service -run TestResolveAffiliateAncestors_MaxThreeLevels`
Expected: FAIL，因为当前只有单层 `InviterID` 使用，没有多层解析函数。

- [ ] **Step 2: 在 repository 中实现按 `inviter_id` 递归/循环查询的祖先解析接口**

```go
type AffiliateAncestor struct {
    UserID int64
    Level  int
}
```

- [ ] **Step 3: 在 service 中增加统一解析函数，防止自环和脏数据死循环**

```go
func (s *AffiliateService) ResolveAffiliateAncestors(ctx context.Context, userID int64, maxDepth int) ([]AffiliateAncestor, error)
```

- [ ] **Step 4: 跑测试确认 A<-B<-C<-D 只返回 3 层，超出截断**

Run: `go test ./backend/internal/service -run TestResolveAffiliateAncestors_MaxThreeLevels`
Expected: PASS

---

### Task 3: 余额包支付成功后，为每笔订单生成三层冻结返利

**Files:**
- Modify: `backend/internal/service/payment_fulfillment.go`
- Modify: `backend/internal/service/affiliate_service.go`
- Modify: `backend/internal/service/payment_balance_package_test.go`
- Test: `backend/internal/service/affiliate_multi_level_rebate_test.go`

- [ ] **Step 1: 写失败测试，证明余额包完成后会生成 6%/3%/1% 三条冻结返利**

```go
func TestBalancePackageFulfillmentCreatesThreeLevelPendingRebates(t *testing.T) {
    t.Skip("RED: should create pending rebate records for balance_package")
}
```

Run: `go test ./backend/internal/service -run TestBalancePackageFulfillmentCreatesThreeLevelPendingRebates`
Expected: FAIL，因为当前余额包不参与返利。

- [ ] **Step 2: 将返利入口从“仅 balance”扩成“balance_package 首版可用”**

```go
if o == nil || o.Amount <= 0 {
    return
}
if o.OrderType != payment.OrderTypeBalancePackage {
    return
}
```

- [ ] **Step 3: 新增按订单生成返利明细的 service 接口，固定比例为 6/3/1，冻结 7 天**

```go
func (s *AffiliateService) CreatePendingRebatesForOrder(ctx context.Context, sourceOrderID, sourceUserID int64, baseAmount float64, paidAt time.Time) error
```

- [ ] **Step 4: 对每个上级分别落账，不存在的层级跳过，且每个订单只生成一次**

```go
availableAt := paidAt.Add(7 * 24 * time.Hour)
records := []rebateRecordInput{
  {Level: 1, Rate: 6, ...},
  {Level: 2, Rate: 3, ...},
  {Level: 3, Rate: 1, ...},
}
```

- [ ] **Step 5: 跑测试确认“每次购买余额包都返利”，不是仅首购一次**

Run: `go test ./backend/internal/service -run 'TestBalancePackageFulfillmentCreatesThreeLevelPendingRebates|TestBalancePackageRepeatedPurchasesCreateRepeatedRebateRecords'`
Expected: PASS

---

### Task 4: 实现 T+7 解冻，冻结返利转可提现返利

**Files:**
- Create: `backend/internal/service/affiliate_rebate_release_job.go`
- Modify: `backend/internal/service/affiliate_service.go`
- Test: `backend/internal/service/affiliate_multi_level_rebate_test.go`

- [ ] **Step 1: 写失败测试，模拟到期返利从 pending 变为 available**

```go
func TestReleaseDueAffiliateRebates(t *testing.T) {
    t.Skip("RED: due pending rebate should become available")
}
```

Run: `go test ./backend/internal/service -run TestReleaseDueAffiliateRebates`
Expected: FAIL，因为当前没有解冻任务。

- [ ] **Step 2: 增加 release job/service 方法，仅处理 `status=pending` 且 `available_at <= now()` 的记录**

```go
func (s *AffiliateService) ReleaseDueRebates(ctx context.Context, now time.Time) (int, error)
```

- [ ] **Step 3: 同步维护用户聚合视图：减少 pending、增加 available**

```go
pending_quota = pending_quota - rebate_amount
available_quota = available_quota + rebate_amount
```

- [ ] **Step 4: 运行测试并验证重复执行幂等**

Run: `go test ./backend/internal/service -run TestReleaseDueAffiliateRebates`
Expected: PASS，重复执行不会二次入账。

---

### Task 5: 实现满 100 元的人工提现申请流程

**Files:**
- Modify: `backend/internal/service/affiliate_service.go`
- Modify: `backend/internal/handler/user_handler.go`
- Create: `backend/internal/handler/admin/affiliate_handler.go`
- Modify: `backend/internal/server/routes/user.go`
- Modify: `backend/internal/server/routes/admin.go`
- Test: `backend/internal/service/affiliate_withdrawal_test.go`

- [ ] **Step 1: 写失败测试，未满 100 元不能提现，满 100 元可发起申请**

```go
func TestCreateAffiliateWithdrawalRequestRequiresMin100(t *testing.T) {
    t.Skip("RED: available rebate must be >= 100")
}
```

Run: `go test ./backend/internal/service -run TestCreateAffiliateWithdrawalRequestRequiresMin100`
Expected: FAIL，因为当前无提现申请模型。

- [ ] **Step 2: 增加用户侧申请接口**

```go
POST /api/v1/user/aff/withdrawals
GET  /api/v1/user/aff/withdrawals
```

- [ ] **Step 3: 发起申请时冻结用户可提现余额，转入 withdrawing**

```go
if available < 100 { return ErrAffiliateWithdrawThreshold }
available_quota -= amount
withdrawing_quota += amount
```

- [ ] **Step 4: 增加后台审核接口（通过 / 驳回 / 标记已打款）**

```go
GET  /api/v1/admin/affiliate/withdrawals
POST /api/v1/admin/affiliate/withdrawals/:id/approve
POST /api/v1/admin/affiliate/withdrawals/:id/reject
POST /api/v1/admin/affiliate/withdrawals/:id/mark-paid
```

- [ ] **Step 5: 跑服务测试验证状态流转**

Run: `go test ./backend/internal/service -run TestCreateAffiliateWithdrawalRequestRequiresMin100`
Expected: PASS

---

### Task 6: 退款时按返利状态做取消、冲正或负债抵扣

**Files:**
- Modify: `backend/internal/service/payment_refund.go`
- Modify: `backend/internal/service/affiliate_service.go`
- Test: `backend/internal/service/affiliate_refund_reversal_test.go`

- [ ] **Step 1: 写失败测试，覆盖三种退款返利场景**

```go
func TestRefundCancelsPendingAffiliateRebates(t *testing.T) { t.Skip("RED") }
func TestRefundReversesAvailableAffiliateRebates(t *testing.T) { t.Skip("RED") }
func TestRefundCreatesAffiliateDebtWhenAlreadyWithdrawn(t *testing.T) { t.Skip("RED") }
```

Run: `go test ./backend/internal/service -run 'TestRefund(CancelsPendingAffiliateRebates|ReversesAvailableAffiliateRebates|CreatesAffiliateDebtWhenAlreadyWithdrawn)'`
Expected: FAIL，因为当前没有返利状态冲正逻辑。

- [ ] **Step 2: 在退款流程中，针对 source_order_id 查询返利记录并按状态处理**

```go
switch record.Status {
case pending:
    record.Status = cancelled
case available:
    record.Status = reversed
case withdraw_requested:
    record.Status = reversed
case withdraw_paid:
    record.DebtAmount += remaining
}
```

- [ ] **Step 3: 对已提现返利生成用户级 debt_quota，并阻止后续提现通过**

```go
if userDebt > 0 {
    return ErrAffiliateDebtOutstanding
}
```

- [ ] **Step 4: 验证退款冲正幂等，重复退款处理不会二次扣减**

Run: `go test ./backend/internal/service -run 'TestRefund(CancelsPendingAffiliateRebates|ReversesAvailableAffiliateRebates|CreatesAffiliateDebtWhenAlreadyWithdrawn)'`
Expected: PASS

---

### Task 7: 前端改造邀请页与后台提现审核页

**Files:**
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/api/user.ts`
- Create: `frontend/src/api/admin/affiliate.ts`
- Modify: `frontend/src/views/user/AffiliateView.vue`
- Create: `frontend/src/views/admin/AffiliateWithdrawalsView.vue`
- Modify: `frontend/src/router/index.ts`
- Modify: `frontend/src/i18n/locales/zh.ts`
- Modify: `frontend/src/i18n/locales/en.ts`
- Test: `frontend/src/views/user/__tests__/AffiliateView.rebate.spec.ts`
- Test: `frontend/src/views/admin/__tests__/AffiliateWithdrawalsView.spec.ts`

- [ ] **Step 1: 写失败测试，用户页展示冻结/可提现/提现中/已提现，且未满 100 不可申请提现**

```ts
it('disables withdrawal request when available rebate is below 100', async () => {
  // RED
})
```

- [ ] **Step 2: 扩充用户端类型与接口返回**

```ts
interface UserAffiliateDetail {
  pending_quota: number
  available_quota: number
  withdrawing_quota: number
  withdrawn_history_quota: number
  debt_quota: number
  rebate_records: AffiliateRebateRecord[]
}
```

- [ ] **Step 3: 将当前“联系管理员”文案改为真实提现申请入口与明细列表**

```vue
<button :disabled="detail.available_quota < 100 || detail.debt_quota > 0">申请提现</button>
```

- [ ] **Step 4: 新增后台提现审核页，支持通过/驳回/标记已打款**

```ts
await adminAffiliateAPI.markWithdrawalPaid(id, { admin_note })
```

- [ ] **Step 5: 跑前端测试与构建**

Run: `pnpm --dir frontend run test:run -- src/views/user/__tests__/AffiliateView.rebate.spec.ts src/views/admin/__tests__/AffiliateWithdrawalsView.spec.ts`
Expected: PASS

Run: `pnpm --dir frontend run build`
Expected: PASS

---

### Task 8: 全链路验证与回归

**Files:**
- Test: `backend/internal/service/affiliate_multi_level_rebate_test.go`
- Test: `backend/internal/service/affiliate_withdrawal_test.go`
- Test: `backend/internal/service/affiliate_refund_reversal_test.go`
- Test: `frontend/src/views/user/__tests__/AffiliateView.rebate.spec.ts`
- Test: `frontend/src/views/admin/__tests__/AffiliateWithdrawalsView.spec.ts`

- [ ] **Step 1: 跑后端相关测试**

Run: `go test ./backend/internal/service -run 'Test(BalancePackageFulfillmentCreatesThreeLevelPendingRebates|BalancePackageRepeatedPurchasesCreateRepeatedRebateRecords|ReleaseDueAffiliateRebates|CreateAffiliateWithdrawalRequestRequiresMin100|RefundCancelsPendingAffiliateRebates|RefundReversesAvailableAffiliateRebates|RefundCreatesAffiliateDebtWhenAlreadyWithdrawn)'`
Expected: PASS

- [ ] **Step 2: 跑前端邀请页与后台提现页测试**

Run: `pnpm --dir frontend run test:run -- src/views/user/__tests__/AffiliateView.rebate.spec.ts src/views/admin/__tests__/AffiliateWithdrawalsView.spec.ts`
Expected: PASS

- [ ] **Step 3: 手动验收关键业务链**

验证清单：
- D 购买余额包后，C/B/A 三层都生成冻结返利
- 同一被邀请人第二次购买余额包，再次独立生成返利
- 到 T+7 后自动解冻
- 用户可提现满 100 后能发起申请
- 后台可标记已打款
- 冻结期退款取消返利
- 已可提现退款冲正返利
- 已提现退款生成负债并阻止后续提现

- [ ] **Step 4: 提交变更**

```bash
git add backend frontend docs/superpowers/specs/2026-04-25-three-level-affiliate-rebate-design.md docs/superpowers/plans/2026-04-25-three-level-affiliate-rebate.md
git commit -m "feat: add three-level affiliate rebate workflow"
```

