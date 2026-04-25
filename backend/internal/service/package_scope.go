package service

import "strings"

const (
	PackageScopeCodex   = "codex"
	PackageScopeGeneral = "general"
)

func NormalizePackageScope(scope string) string {
	switch strings.TrimSpace(scope) {
	case PackageScopeCodex:
		return PackageScopeCodex
	case PackageScopeGeneral:
		return PackageScopeGeneral
	default:
		return ""
	}
}

func PackageScopeMatchesGroup(userScope, groupScope string) bool {
	left := NormalizePackageScope(userScope)
	right := NormalizePackageScope(groupScope)
	if left == "" || right == "" {
		return false
	}
	return left == right
}

func ResolveBackfillPackageScope(user *User) string {
	if user == nil {
		return ""
	}
	if user.Balance > 0 {
		return PackageScopeCodex
	}
	return ""
}

func ensurePackageScopeForPositiveBalanceCredit(user *User, amount float64) bool {
	if user == nil || amount <= 0 || user.PackageScope != nil {
		return false
	}
	scope := PackageScopeCodex
	user.PackageScope = &scope
	return true
}
