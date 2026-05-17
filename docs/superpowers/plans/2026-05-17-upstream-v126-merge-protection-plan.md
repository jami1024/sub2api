# Upstream v0.1.126 Merge Protection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在隔离分支中合并 `upstream/main` 到本地 `main` 的基础上解决 7 个冲突文件，并确保余额套餐、`package_scope`、强制切换模式、支付页/支付结果页、首页套餐展示、返利相关现有功能全部不回退。

**Architecture:** 先以测试锁住现有前端行为，再修复后端订单语义，最后让 Ent 生成代码与当前 schema 真相重新对齐。整个过程只在隔离 worktree 中进行，按“前端行为 → 后端语义 → 生成代码 → 全量验证”的顺序推进，每完成一组就做定向验证，避免最后集中爆雷。

**Tech Stack:** Git worktree、Vue 3 + TypeScript + Vitest、Go + Ent、Make、pnpm

---

### Task 1: 重新进入合并态并锁定当前保护基线

**Files:**
- Modify: `.worktrees/merge-upstream-main-v126-2026-05-17` 工作区 git 状态
- Verify: `docs/superpowers/specs/2026-05-17-upstream-v126-merge-protection-design.md`
- Output: 冲突文件清单与基线状态记录

- [ ] **Step 1: 确认设计文档已存在且当前分支干净**

Run:
```bash
git status --short --branch
ls docs/superpowers/specs/2026-05-17-upstream-v126-merge-protection-design.md
```
Expected:
```text
## merge-upstream-main-v126-2026-05-17
... design file path exists ...
```

- [ ] **Step 2: 重新进入上游合并态**

Run:
```bash
git merge --no-ff --no-commit upstream/main
```
Expected:
```text
Auto-merging ...
CONFLICT (content): Merge conflict in ...
Automatic merge failed; fix conflicts and then commit the result.
```

- [ ] **Step 3: 记录必须处理的 7 个冲突文件**

Run:
```bash
git diff --name-only --diff-filter=U
```
Expected:
```text
backend/ent/migrate/schema.go
backend/ent/mutation.go
backend/ent/runtime/runtime.go
backend/internal/service/payment_order.go
frontend/src/types/index.ts
frontend/src/views/user/PaymentResultView.vue
frontend/src/views/user/PaymentView.vue
```

- [ ] **Step 4: 确认保护边界仍与设计文档一致**

Check list:
```text
- balance_package
- package_scope
- force_switch_scope
- payment result fee fallback
- home package section
- affiliate behavior
```
Expected: 上述关键词都能在设计文档或现有代码/测试中找到对应证据。

- [ ] **Step 5: 提交一个仅记录“进入合并态”的工作笔记提交或保留在工作日志中，不改代码**

Run:
```bash
git status --short
```
Expected: 只有合并引入的文件变化，没有额外手工文件。

### Task 2: 先用前端测试锁住支付页和结果页现有行为

**Files:**
- Modify: `frontend/src/views/user/__tests__/PaymentView.spec.ts`
- Modify: `frontend/src/views/user/__tests__/PaymentResultView.spec.ts`
- Verify: `frontend/src/views/user/PaymentView.vue`
- Verify: `frontend/src/views/user/PaymentResultView.vue`
- Verify: `frontend/src/types/index.ts`

- [ ] **Step 1: 为支付页冲突补一个“混合套餐 + 强制切换”回归测试（若现有测试已覆盖，则补一个更具体的 payload 断言）**

Add to `frontend/src/views/user/__tests__/PaymentView.spec.ts`:
```ts
it('submits balance_package payload with balance_package_id and force_switch_scope for conflicting package', async () => {
  mockUser.package_scope = 'codex'
  getCheckoutInfo.mockResolvedValueOnce(checkoutInfoWithMixedBalancePackagesFixture())
  createOrder.mockResolvedValueOnce({
    order_id: 321,
    amount: 80,
    pay_amount: 88,
    fee_rate: 0,
    expires_at: '2099-01-01T00:10:00.000Z',
    payment_type: 'wxpay',
    out_trade_no: 'sub2_bp_force_321',
    result_type: 'jsapi_ready',
    resume_token: 'resume-force-321',
    jsapi: {
      appId: 'wx123',
      timeStamp: '1712345678',
      nonceStr: 'nonce',
      package: 'prepay_id=wx123',
      signType: 'RSA',
      paySign: 'sign',
    },
  })

  const wrapper = mount(PaymentView)
  await flushPromises()
  await wrapper.get('[data-testid="payment-tab-package"]').trigger('click')
  await wrapper.get('[data-testid="balance-package-force-switch-10"]').trigger('click')
  await wrapper.get('[data-testid="confirm-force-switch"]').trigger('click')
  await flushPromises()

  expect(createOrder).toHaveBeenCalledWith(expect.objectContaining({
    amount: 80,
    order_type: 'balance_package',
    balance_package_id: 10,
    force_switch_scope: true,
  }))
})
```

- [ ] **Step 2: 只运行这个新测试，确认它先失败**

Run:
```bash
pnpm --dir frontend exec vitest run frontend/src/views/user/__tests__/PaymentView.spec.ts -t "submits balance_package payload with balance_package_id and force_switch_scope for conflicting package"
```
Expected:
```text
FAIL
```
Failure reason should be one of:
```text
- merge conflict markers still present
- selector not found because package tab / dialog behavior broken
- createOrder payload missing balance_package_id / force_switch_scope
```

- [ ] **Step 3: 为支付结果页冲突补一个“fee_rate 缺失也不显示 NaN，balance_package 用美元前缀”的回归测试（如果已有同名测试，直接增强断言）**

Ensure `frontend/src/views/user/__tests__/PaymentResultView.spec.ts` contains a case like:
```ts
it('renders balance_package credited amount with dollar prefix and safe base amount when fee_rate is missing', async () => {
  routeState.query = { resume_token: 'resume-balance-package-1' }
  resolveOrderPublicByResumeToken.mockResolvedValue({
    data: {
      ...orderFactory('COMPLETED'),
      order_type: 'balance_package',
      amount: 1,
      pay_amount: 0.1,
      fee_rate: undefined,
    },
  })

  const wrapper = mount(PaymentResultView, {
    global: { stubs: { OrderStatusBadge: true } },
  })
  await flushPromises()

  expect(wrapper.text()).toContain('$1.00')
  expect(wrapper.text()).toContain('¥0.10')
  expect(wrapper.text()).not.toContain('NaN')
})
```

- [ ] **Step 4: 只运行结果页测试，确认它先失败**

Run:
```bash
pnpm --dir frontend exec vitest run frontend/src/views/user/__tests__/PaymentResultView.spec.ts -t "renders balance_package credited amount with dollar prefix and safe base amount when fee_rate is missing"
```
Expected:
```text
FAIL
```

- [ ] **Step 5: 解决 3 个前端冲突文件**

Modify:
- `frontend/src/views/user/PaymentView.vue`
- `frontend/src/views/user/PaymentResultView.vue`
- `frontend/src/types/index.ts`

Keep these exact outcomes:
```text
PaymentView.vue:
- package tab 保留
- 当前 scope 提示保留
- force-switch dialog 保留
- createOrder payload 继续写 balance_package_id / force_switch_scope
- help_text/help_image 只在未选套餐时显示

PaymentResultView.vue:
- normalizedFeeRate 保留
- baseAmount / feeAmount 对 undefined fee_rate 容错
- balance / balance_package credited amount 用 $ 前缀
- 其它支付金额维持 ¥ 前缀

types/index.ts:
- User.package_scope 保留
- Group.package_scope 保留
- PackageScope 类型保留
- 本地 affiliate 扩展字段保留
- 不丢失上游需要的认证 provider / 其它新增字段
```

- [ ] **Step 6: 运行支付页与结果页定向测试，确认转绿**

Run:
```bash
pnpm --dir frontend exec vitest run frontend/src/views/user/__tests__/PaymentView.spec.ts frontend/src/views/user/__tests__/PaymentResultView.spec.ts
```
Expected:
```text
PASS
```

- [ ] **Step 7: 运行 package-scope 相关前端测试，确认没有侧面回退**

Run:
```bash
pnpm --dir frontend exec vitest run frontend/src/views/user/__tests__/KeysView.package-scope.spec.ts frontend/src/views/user/__tests__/PaymentView.spec.ts
```
Expected:
```text
PASS
```

- [ ] **Step 8: 提交前端冲突解决**

Run:
```bash
git add frontend/src/views/user/PaymentView.vue frontend/src/views/user/PaymentResultView.vue frontend/src/types/index.ts frontend/src/views/user/__tests__/PaymentView.spec.ts frontend/src/views/user/__tests__/PaymentResultView.spec.ts
git commit -m "merge: preserve payment package frontend behavior"
```

### Task 3: 用后端测试锁住订单创建与余额套餐语义

**Files:**
- Modify: `backend/internal/service/payment_balance_package_test.go`
- Modify: `backend/internal/service/payment_order.go`
- Verify: `backend/internal/service/payment_fulfillment.go`
- Verify: `backend/internal/service/payment_service.go`
- Verify: `backend/internal/payment/types.go`

- [ ] **Step 1: 为订单创建冲突补一个“冲突 scope 且未强制切换时返回冲突错误”的测试**

Add to `backend/internal/service/payment_balance_package_test.go`:
```go
func TestCreateBalancePackageOrder_RejectsDifferentScopeWithoutForceSwitch(t *testing.T) {
	client := newPackageScopeEntClient(t)
	createdUser, err := client.User.Create().
		SetEmail("scope-conflict@example.com").
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
	svc := &PaymentService{entClient: client, configService: NewPaymentConfigService(client, &paymentConfigSettingRepoStub{values: map[string]string{}}, nil)}
	_, err = svc.CreateOrder(context.Background(), CreateOrderRequest{
		UserID:           createdUser.ID,
		PaymentType:      payment.TypeAlipay,
		OrderType:        payment.OrderTypeBalancePackage,
		BalancePackageID: pkg.ID,
		ClientIP:         "127.0.0.1",
		SrcHost:          "example.com",
	}, &User{ID: createdUser.ID, Email: createdUser.Email, Username: createdUser.Username, Balance: 15, PackageScope: &codex}, &PaymentConfig{Enabled: true, MaxPendingOrders: 3, OrderTimeoutMin: 30})
	require.Error(t, err)
	require.Contains(t, err.Error(), "PACKAGE_SCOPE_CONFLICT")
}
```

- [ ] **Step 2: 只运行这个新测试，确认它先失败**

Run:
```bash
mkdir -p .cache/go-build .cache/gomod && GOCACHE=$PWD/.cache/go-build GOMODCACHE=$PWD/.cache/gomod go test ./backend/internal/service -run TestCreateBalancePackageOrder_RejectsDifferentScopeWithoutForceSwitch -count=1
```
Expected:
```text
FAIL
```
If blocked by toolchain download, expected blocking evidence is:
```text
go: download go1.26.2 ... i/o timeout
```

- [ ] **Step 3: 确认并保留这 4 条订单语义**

Modify `backend/internal/service/payment_order.go` so that all four paths remain true:
```text
- validateOrderInput 返回 (plan, balancePackage, error)
- balance_package 订单会写入 BalancePackageID 和 PackageScopeSnapshot
- 有余额且 scope 冲突时，未带 ForceSwitchScope 必须拒绝
- invokeProvider/buildPaymentSubject 同时兼容 subscription 与 balance_package
```

- [ ] **Step 4: 保留或补充“订单落库快照”和“强制切换成功”测试**

Ensure these existing tests remain valid and green:
```go
func TestCreateBalancePackageOrderStoresSnapshot(t *testing.T) { ... }
func TestCreateBalancePackageOrder_AllowsDifferentScopeWithForceSwitch(t *testing.T) { ... }
```

- [ ] **Step 5: 运行余额套餐后端定向测试**

Run:
```bash
mkdir -p .cache/go-build .cache/gomod && GOCACHE=$PWD/.cache/go-build GOMODCACHE=$PWD/.cache/gomod go test ./backend/internal/service -run 'TestCreateBalancePackageOrder|TestExecuteBalancePackageFulfillment' -count=1
```
Expected:
```text
PASS
```
Or, if still blocked by toolchain download, capture exact timeout output and do not claim backend green.

- [ ] **Step 6: 提交后端订单冲突解决**

Run:
```bash
git add backend/internal/service/payment_order.go backend/internal/service/payment_balance_package_test.go
git commit -m "merge: preserve balance package order semantics"
```

### Task 4: 让 Ent 生成代码与 schema 真相重新对齐

**Files:**
- Modify: `backend/ent/migrate/schema.go`
- Modify: `backend/ent/mutation.go`
- Modify: `backend/ent/runtime/runtime.go`
- Verify: `backend/ent/schema/group.go`
- Verify: `backend/ent/schema/user.go`
- Verify: `backend/ent/schema/payment_order.go`
- Verify: `backend/ent/schema/balance_package.go`

- [ ] **Step 1: 明确 schema 真相来源**

Run:
```bash
sed -n '1,220p' backend/ent/schema/group.go
sed -n '1,220p' backend/ent/schema/user.go
sed -n '1,260p' backend/ent/schema/payment_order.go
sed -n '1,220p' backend/ent/schema/balance_package.go
```
Expected: 能明确看到以下字段在 schema 中存在：
```text
Group.package_scope
User.package_scope
PaymentOrder.balance_package_id
PaymentOrder.package_scope_snapshot
PaymentOrder.force_switch_scope
BalancePackage entity
```

- [ ] **Step 2: 解决 3 个 Ent 冲突文件，禁止凭感觉删字段**

Modify:
- `backend/ent/migrate/schema.go`
- `backend/ent/mutation.go`
- `backend/ent/runtime/runtime.go`

Keep these exact outcomes:
```text
- BalancePackage 类型和表定义存在
- Group/User 的 package_scope 没丢
- PaymentOrder 的 balance_package_id / package_scope_snapshot / force_switch_scope 没丢
- runtime validator/default/index offset 与 schema 顺序一致
```

- [ ] **Step 3: 搜索冲突标记，确认 7 个文件都已无冲突残留**

Run:
```bash
rg -n '^<<<<<<<|^=======|^>>>>>>>' backend/ent backend/internal/service frontend/src/views/user frontend/src/types
```
Expected:
```text
(no output)
```

- [ ] **Step 4: 运行一组最小后端契约测试或编译级检查**

Run:
```bash
mkdir -p .cache/go-build .cache/gomod && GOCACHE=$PWD/.cache/go-build GOMODCACHE=$PWD/.cache/gomod go test ./backend/internal/service -run 'TestCreateBalancePackageOrderStoresSnapshot|TestCreateBalancePackageOrder_AllowsDifferentScopeWithForceSwitch' -count=1
```
Expected:
```text
PASS
```
If blocked, record the exact toolchain timeout and move to final verification with blocker noted.

- [ ] **Step 5: 提交 Ent 对齐结果**

Run:
```bash
git add backend/ent/migrate/schema.go backend/ent/mutation.go backend/ent/runtime/runtime.go
git commit -m "merge: align ent generated files with package scope schema"
```

### Task 5: 做最终验证并输出候选合并结果

**Files:**
- Verify: 所有 7 个冲突文件
- Verify: `frontend/src/views/user/__tests__/PaymentView.spec.ts`
- Verify: `frontend/src/views/user/__tests__/PaymentResultView.spec.ts`
- Verify: `frontend/src/views/user/__tests__/KeysView.package-scope.spec.ts`
- Verify: `backend/internal/service/payment_balance_package_test.go`
- Verify: `docs/superpowers/specs/2026-05-17-upstream-v126-merge-protection-design.md`

- [ ] **Step 1: 运行前端关键回归套件**

Run:
```bash
pnpm --dir frontend exec vitest run \
  frontend/src/views/user/__tests__/PaymentView.spec.ts \
  frontend/src/views/user/__tests__/PaymentResultView.spec.ts \
  frontend/src/views/user/__tests__/KeysView.package-scope.spec.ts
```
Expected:
```text
PASS
```

- [ ] **Step 2: 尝试运行后端关键回归套件**

Run:
```bash
mkdir -p .cache/go-build .cache/gomod && GOCACHE=$PWD/.cache/go-build GOMODCACHE=$PWD/.cache/gomod go test ./backend/internal/service -run 'TestCreateBalancePackageOrder|TestExecuteBalancePackageFulfillment' -count=1
```
Expected either:
```text
PASS
```
or blocker evidence:
```text
go: download go1.26.2 ... i/o timeout
```

- [ ] **Step 3: 运行合并后完整冲突检查**

Run:
```bash
git diff --name-only --diff-filter=U
rg -n '^<<<<<<<|^=======|^>>>>>>>' .
```
Expected:
```text
(no output)
```

- [ ] **Step 4: 检查最终工作区只包含预期改动**

Run:
```bash
git status --short
```
Expected: 只剩本次合并和测试补充相关文件，不应出现无关文件。

- [ ] **Step 5: 如后端工具链恢复可用，再跑一次完整项目验证**

Run:
```bash
mkdir -p .cache/go-build .cache/gomod && GOCACHE=$PWD/.cache/go-build GOMODCACHE=$PWD/.cache/gomod make test
```
Expected:
```text
PASS
```
If this step still被工具链阻塞，最终报告必须明确写出“前端验证通过，后端完整验证因 Go 1.26.2 下载超时未完成”。

- [ ] **Step 6: 合并提交**

Run:
```bash
git add -A
git commit -m "merge: sync upstream main v0.1.126 with local payment protections"
```

- [ ] **Step 7: 产出最终说明**

Include these exact sections in the final report:
```text
- Outcome: 是否已形成可审阅候选合并结果
- Preserved behaviors: 保住了哪些本地功能
- Upstream absorbed: 吸收了哪些上游变化
- Verification: 前端/后端分别跑了什么，结果如何
- Remaining blocker: 是否还存在 Go 1.26.2 工具链下载超时
```
