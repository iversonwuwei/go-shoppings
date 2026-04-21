package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/cache"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/jwtx"
	"wechat-mall-saas/internal/pkg/utils"
	"wechat-mall-saas/internal/pkg/wxapp"
	"wechat-mall-saas/internal/repository"
)

type AuthService struct {
	admins  *repository.AdminRepo
	members *repository.MemberRepo
	tenants *repository.TenantRepo
	jwt     *jwtx.Manager
	cache   *cache.Client
	env     string
}

func NewAuthService(admins *repository.AdminRepo, members *repository.MemberRepo, tenants *repository.TenantRepo, j *jwtx.Manager, rdb *cache.Client, env string) *AuthService {
	return &AuthService{admins: admins, members: members, tenants: tenants, jwt: j, cache: rdb, env: env}
}

type AdminLoginResult struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	Admin        *model.Admin `json:"admin"`
}

func (s *AuthService) AdminLogin(ctx context.Context, username, password, ip string) (*AdminLoginResult, error) {
	a, err := s.admins.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if a == nil || a.Status != 1 {
		return nil, apperr.New(10010, "账号不存在或被禁用")
	}
	if !utils.CheckPassword(a.Password, password) {
		return nil, apperr.New(10011, "账号或密码错误")
	}
	tok, err := s.jwt.IssueWithRole(jwtx.SubjectAdmin, a.ID, a.TenantID, a.Role, "")
	if err != nil {
		return nil, err
	}
	rt, err := s.jwt.IssueRefresh(jwtx.SubjectAdmin, a.ID, a.TenantID)
	if err != nil {
		return nil, err
	}
	_ = s.admins.UpdateLogin(ctx, a.ID, ip)
	return &AdminLoginResult{Token: tok, RefreshToken: rt, Admin: a}, nil
}

type MemberLoginResult struct {
	Token  string        `json:"token"`
	Member *model.Member `json:"member"`
}

// MemberLoginByWechat 通过 code 登录/注册。wx 为租户对应的 wxapp.Client。
func (s *AuthService) MemberLoginByWechat(ctx context.Context, wx *wxapp.Client, code string) (*MemberLoginResult, error) {
	sess, err := wx.Code2Session(code)
	if err != nil {
		return nil, apperr.ErrWechatAPI
	}
	m, err := s.members.FindByOpenID(ctx, sess.OpenID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		m = &model.Member{
			OpenID:     sess.OpenID,
			UnionID:    sess.UnionID,
			SessionKey: sess.SessionKey,
			Status:     1,
		}
		if err := s.members.Create(ctx, m); err != nil {
			return nil, err
		}
	} else {
		_ = s.members.UpdateFields(ctx, m.ID, map[string]interface{}{
			"session_key": sess.SessionKey,
		})
	}
	tok, err := s.jwt.Issue(jwtx.SubjectMember, m.ID, m.TenantID, m.OpenID)
	if err != nil {
		return nil, err
	}
	return &MemberLoginResult{Token: tok, Member: m}, nil
}

// BindPhoneByWechat 解密微信手机号并绑定到当前会员
func (s *AuthService) BindPhoneByWechat(ctx context.Context, memberID uint64, encryptedData, iv string) (string, error) {
	m, err := s.members.FindByID(ctx, memberID)
	if err != nil || m == nil {
		return "", apperr.ErrNotFound
	}
	info, err := wxapp.DecryptPhone(m.SessionKey, encryptedData, iv)
	if err != nil {
		return "", apperr.ErrWechatAPI
	}
	_ = s.members.UpdateFields(ctx, memberID, map[string]interface{}{"phone": info.PurePhoneNumber})
	return info.PurePhoneNumber, nil
}

// ==================== 短信验证码 / 手机号登录 / 忘记密码 ====================

// 验证码用途
const (
	VerifyPurposeApply = "apply"          // 入驻申请手机号验证
	VerifyPurposeLogin = "login"          // 手机号+验证码登录
	VerifyPurposeReset = "reset_password" // 忘记密码重置
)

func codeKey(purpose, phone string) string { return fmt.Sprintf("verify:%s:%s", purpose, phone) }
func codeLockKey(purpose, phone string) string {
	return fmt.Sprintf("verify:%s:%s:lock", purpose, phone)
}

func gen6Digit() string {
	// 000000-999999
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

// SendVerifyCode 生成并下发短信验证码。非 production 环境会把验证码回传，便于联调。
// purpose: apply / login / reset_password
func (s *AuthService) SendVerifyCode(ctx context.Context, phone, purpose string) (string, error) {
	if phone == "" {
		return "", apperr.ErrParamInvalid
	}
	switch purpose {
	case VerifyPurposeApply, VerifyPurposeLogin, VerifyPurposeReset:
	default:
		return "", apperr.ErrParamInvalid
	}
	if s.cache == nil {
		return "", apperr.ErrInternal
	}
	// 60 秒发送节流
	if ok, _ := s.cache.SetNX(ctx, codeLockKey(purpose, phone), "1", 60*time.Second).Result(); !ok {
		return "", apperr.New(30011, "验证码发送过于频繁，请稍后再试")
	}
	code := gen6Digit()
	if err := s.cache.Set(ctx, codeKey(purpose, phone), code, 5*time.Minute).Err(); err != nil {
		return "", err
	}
	// TODO: 真实短信网关对接。当前仅在非生产环境回传验证码。
	if s.env != "" && s.env != "production" {
		return code, nil
	}
	return "", nil
}

// VerifyAndConsumeCode 校验验证码并消费（成功后删除）
func (s *AuthService) VerifyAndConsumeCode(ctx context.Context, phone, purpose, code string) error {
	if phone == "" || code == "" {
		return apperr.ErrParamInvalid
	}
	if s.cache == nil {
		return apperr.ErrInternal
	}
	v, err := s.cache.Get(ctx, codeKey(purpose, phone)).Result()
	if err != nil || v == "" {
		return apperr.New(30012, "验证码已失效，请重新获取")
	}
	if v != code {
		return apperr.New(30013, "验证码错误")
	}
	_ = s.cache.Del(ctx, codeKey(purpose, phone)).Err()
	return nil
}

// AdminLoginBySMS 根据租户ID + 手机号 + 验证码登录。tenantID=0 表示平台管理员。
func (s *AuthService) AdminLoginBySMS(ctx context.Context, tenantID uint64, phone, code, ip string) (*AdminLoginResult, error) {
	if err := s.VerifyAndConsumeCode(ctx, phone, VerifyPurposeLogin, code); err != nil {
		return nil, err
	}
	a, err := s.admins.FindByPhone(ctx, tenantID, phone)
	if err != nil {
		return nil, err
	}
	if a == nil || a.Status != 1 {
		return nil, apperr.New(10010, "手机号未绑定管理员账号或账号被禁用")
	}
	tok, err := s.jwt.IssueWithRole(jwtx.SubjectAdmin, a.ID, a.TenantID, a.Role, "")
	if err != nil {
		return nil, err
	}
	rt, err := s.jwt.IssueRefresh(jwtx.SubjectAdmin, a.ID, a.TenantID)
	if err != nil {
		return nil, err
	}
	_ = s.admins.UpdateLogin(ctx, a.ID, ip)
	return &AdminLoginResult{Token: tok, RefreshToken: rt, Admin: a}, nil
}

// ResetAdminPassword 通过手机号 + 验证码重置管理员密码
func (s *AuthService) ResetAdminPassword(ctx context.Context, tenantID uint64, phone, code, newPassword string) error {
	if len(newPassword) < 6 {
		return apperr.New(20001, "新密码长度至少 6 位")
	}
	if err := s.VerifyAndConsumeCode(ctx, phone, VerifyPurposeReset, code); err != nil {
		return err
	}
	a, err := s.admins.FindByPhone(ctx, tenantID, phone)
	if err != nil {
		return err
	}
	if a == nil {
		return apperr.New(10010, "手机号未绑定管理员账号")
	}
	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.admins.UpdatePassword(ctx, a.ID, hash)
}

// RegisterTenantWithAdmin 提交入驻申请，并为该租户创建初始管理员账号。
// 若 verifyCode 非空，会按 apply 用途校验验证码；username/password 用于创建管理员。
func (s *AuthService) RegisterTenantWithAdmin(ctx context.Context, tenantSvc *TenantService, in *model.Tenant, username, password, verifyCode string) (*model.Tenant, *model.Admin, error) {
	if username == "" || password == "" {
		return nil, nil, apperr.New(20001, "请填写管理员用户名和密码")
	}
	if len(password) < 6 {
		return nil, nil, apperr.New(20001, "管理员密码长度至少 6 位")
	}
	// 校验手机号验证码
	if err := s.VerifyAndConsumeCode(ctx, in.ContactPhone, VerifyPurposeApply, verifyCode); err != nil {
		return nil, nil, err
	}
	// 用户名重复校验
	if exist, _ := s.admins.FindByUsername(ctx, username); exist != nil {
		return nil, nil, apperr.New(30006, "管理员用户名已被占用")
	}
	t, err := tenantSvc.Register(ctx, in)
	if err != nil {
		return nil, nil, err
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		return t, nil, err
	}
	admin := &model.Admin{
		Username: username,
		Password: hash,
		RealName: in.ContactName,
		Phone:    in.ContactPhone,
		Email:    in.ContactEmail,
		Role:     "admin",
		TenantID: t.ID,
		Status:   0, // 待租户审核通过后由平台激活
	}
	if err := s.admins.Create(ctx, admin); err != nil {
		return t, nil, err
	}
	return t, admin, nil
}
