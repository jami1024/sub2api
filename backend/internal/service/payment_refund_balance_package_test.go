package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestRequestRefundRejectsBalancePackageOrders(t *testing.T) {
	ctx := context.Background()
	client := newPackageScopeEntClient(t)

	user, err := client.User.Create().
		SetEmail("refund-bp@example.com").
		SetPasswordHash("hash").
		SetUsername("refund-bp").
		SetBalance(100).
		SetConcurrency(1).
		SetStatus(StatusActive).
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(100).
		SetPayAmount(100).
		SetFeeRate(0).
		SetRechargeCode("PAY-BP-REFUND").
		SetOutTradeNo("sub2_bp_refund").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-bp-refund").
		SetOrderType(payment.OrderTypeBalancePackage).
		SetBalancePackageID(10).
		SetPackageScopeSnapshot(PackageScopeCodex).
		SetStatus(OrderStatusCompleted).
		SetPaidAt(time.Now()).
		SetCompletedAt(time.Now()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("example.com").
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}
	err = svc.RequestRefund(ctx, order.ID, user.ID, "test")
	require.Error(t, err)
	require.ErrorContains(t, err, "REFUND_UNSUPPORTED")
}
