package service

import (
	"context"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/balancepackage"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

func validateBalancePackageRequired(name string, price, creditAmount float64, packageScope string) error {
	if strings.TrimSpace(name) == "" {
		return infraerrors.BadRequest("BALANCE_PACKAGE_NAME_REQUIRED", "balance package name is required")
	}
	if price <= 0 {
		return infraerrors.BadRequest("BALANCE_PACKAGE_PRICE_INVALID", "price must be > 0")
	}
	if creditAmount <= 0 {
		return infraerrors.BadRequest("BALANCE_PACKAGE_CREDIT_INVALID", "credit amount must be > 0")
	}
	if NormalizePackageScope(packageScope) == "" {
		return infraerrors.BadRequest("BALANCE_PACKAGE_SCOPE_INVALID", "package scope is invalid")
	}
	return nil
}

func validateBalancePackagePatch(req UpdateBalancePackageRequest) error {
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return infraerrors.BadRequest("BALANCE_PACKAGE_NAME_REQUIRED", "balance package name is required")
	}
	if req.Price != nil && *req.Price <= 0 {
		return infraerrors.BadRequest("BALANCE_PACKAGE_PRICE_INVALID", "price must be > 0")
	}
	if req.CreditAmount != nil && *req.CreditAmount <= 0 {
		return infraerrors.BadRequest("BALANCE_PACKAGE_CREDIT_INVALID", "credit amount must be > 0")
	}
	if req.PackageScope != nil && NormalizePackageScope(*req.PackageScope) == "" {
		return infraerrors.BadRequest("BALANCE_PACKAGE_SCOPE_INVALID", "package scope is invalid")
	}
	return nil
}

func (s *PaymentConfigService) ListBalancePackages(ctx context.Context, forSaleOnly bool) ([]*dbent.BalancePackage, error) {
	q := s.entClient.BalancePackage.Query()
	if forSaleOnly {
		q = q.Where(balancepackage.ForSaleEQ(true))
	}
	return q.Order(balancepackage.BySortOrder()).All(ctx)
}

func (s *PaymentConfigService) GetBalancePackage(ctx context.Context, id int64) (*dbent.BalancePackage, error) {
	pkg, err := s.entClient.BalancePackage.Get(ctx, id)
	if err != nil {
		return nil, infraerrors.NotFound("BALANCE_PACKAGE_NOT_FOUND", "balance package not found")
	}
	return pkg, nil
}

func (s *PaymentConfigService) CreateBalancePackage(ctx context.Context, req CreateBalancePackageRequest) (*dbent.BalancePackage, error) {
	if err := validateBalancePackageRequired(req.Name, req.Price, req.CreditAmount, req.PackageScope); err != nil {
		return nil, err
	}
	scope := NormalizePackageScope(req.PackageScope)
	return s.entClient.BalancePackage.Create().
		SetName(req.Name).
		SetDescription(req.Description).
		SetPrice(req.Price).
		SetCreditAmount(req.CreditAmount).
		SetPackageScope(scope).
		SetProductName(req.ProductName).
		SetForSale(req.ForSale).
		SetSortOrder(req.SortOrder).
		Save(ctx)
}

func (s *PaymentConfigService) UpdateBalancePackage(ctx context.Context, id int64, req UpdateBalancePackageRequest) (*dbent.BalancePackage, error) {
	if err := validateBalancePackagePatch(req); err != nil {
		return nil, err
	}
	u := s.entClient.BalancePackage.UpdateOneID(id)
	if req.Name != nil {
		u.SetName(*req.Name)
	}
	if req.Description != nil {
		u.SetDescription(*req.Description)
	}
	if req.Price != nil {
		u.SetPrice(*req.Price)
	}
	if req.CreditAmount != nil {
		u.SetCreditAmount(*req.CreditAmount)
	}
	if req.PackageScope != nil {
		u.SetPackageScope(NormalizePackageScope(*req.PackageScope))
	}
	if req.ProductName != nil {
		u.SetProductName(*req.ProductName)
	}
	if req.ForSale != nil {
		u.SetForSale(*req.ForSale)
	}
	if req.SortOrder != nil {
		u.SetSortOrder(*req.SortOrder)
	}
	return u.Save(ctx)
}

func (s *PaymentConfigService) DeleteBalancePackage(ctx context.Context, id int64) error {
	err := s.entClient.BalancePackage.DeleteOneID(id).Exec(ctx)
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "foreign key") || strings.Contains(msg, "constraint") {
		return infraerrors.Conflict("BALANCE_PACKAGE_REFERENCED", fmt.Sprintf("balance package %d is referenced by historical orders and cannot be deleted", id))
	}
	return err
}
