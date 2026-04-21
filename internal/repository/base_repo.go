// Package repository 数据访问层
//
// 所有带 tenant_id 的查询/写入，必须使用 TenantDB 获取已自动追加 tenant_id 条件
// 的 *gorm.DB 实例，避免跨租户数据泄漏。
package repository

import (
	"context"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/pkg/ctxkeys"
)

// TenantDB 返回自动附加 tenant_id 过滤条件的 *gorm.DB。
// 仅限业务表（含 tenant_id 列）使用；平台表（plans/admins/tenants）应直接用 *gorm.DB。
func TenantDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	tx := db.WithContext(ctx)
	if t := ctxkeys.GetTenant(ctx); t != nil && t.ID > 0 {
		return tx.Where("tenant_id = ?", t.ID)
	}
	return tx
}

// EnsureTenant 返回当前租户 ID；未注入时返回 0，由业务层决定是否拒绝。
func EnsureTenant(ctx context.Context) uint64 {
	if t := ctxkeys.GetTenant(ctx); t != nil {
		return t.ID
	}
	return 0
}
