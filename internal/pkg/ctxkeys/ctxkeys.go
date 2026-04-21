package ctxkeys

import "context"

type ctxKey string

const (
	TenantCtxKey ctxKey = "tenant_ctx"
	AdminCtxKey  ctxKey = "admin_ctx"
	MemberCtxKey ctxKey = "member_ctx"
)

type TenantInfo struct {
	ID       uint64
	Code     string
	PlanID   uint64
	Features []string
	Status   int8
	Expired  bool
}

type AdminInfo struct {
	ID       uint64
	Username string
	Role     string
	TenantID uint64 // 0 = 平台管理员
}

type MemberInfo struct {
	ID       uint64
	OpenID   string
	TenantID uint64
}

func WithTenant(ctx context.Context, t *TenantInfo) context.Context {
	return context.WithValue(ctx, TenantCtxKey, t)
}

func GetTenant(ctx context.Context) *TenantInfo {
	if v, ok := ctx.Value(TenantCtxKey).(*TenantInfo); ok {
		return v
	}
	return nil
}

func WithAdmin(ctx context.Context, a *AdminInfo) context.Context {
	return context.WithValue(ctx, AdminCtxKey, a)
}

func GetAdmin(ctx context.Context) *AdminInfo {
	if v, ok := ctx.Value(AdminCtxKey).(*AdminInfo); ok {
		return v
	}
	return nil
}

func WithMember(ctx context.Context, m *MemberInfo) context.Context {
	return context.WithValue(ctx, MemberCtxKey, m)
}

func GetMember(ctx context.Context) *MemberInfo {
	if v, ok := ctx.Value(MemberCtxKey).(*MemberInfo); ok {
		return v
	}
	return nil
}

func HasFeature(t *TenantInfo, feat string) bool {
	if t == nil {
		return false
	}
	for _, f := range t.Features {
		if f == feat {
			return true
		}
	}
	return false
}
