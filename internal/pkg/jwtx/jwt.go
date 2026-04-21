package jwtx

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Subject string

const (
	SubjectAdmin  Subject = "admin"
	SubjectMember Subject = "member"
)

type Claims struct {
	Subject  Subject `json:"sub"`
	UserID   uint64  `json:"uid"`
	TenantID uint64  `json:"tid"`
	OpenID   string  `json:"oid,omitempty"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret     []byte
	expire     time.Duration
	refreshExp time.Duration
}

func New(secret string, expireHours, refreshHours int) *Manager {
	return &Manager{
		secret:     []byte(secret),
		expire:     time.Duration(expireHours) * time.Hour,
		refreshExp: time.Duration(refreshHours) * time.Hour,
	}
}

func (m *Manager) Issue(sub Subject, userID, tenantID uint64, openID string) (string, error) {
	claims := Claims{
		Subject:  sub,
		UserID:   userID,
		TenantID: tenantID,
		OpenID:   openID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(m.secret)
}

func (m *Manager) IssueRefresh(sub Subject, userID, tenantID uint64) (string, error) {
	claims := Claims{
		Subject:  sub,
		UserID:   userID,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(m.secret)
}

func (m *Manager) Parse(token string) (*Claims, error) {
	c := &Claims{}
	tok, err := jwt.ParseWithClaims(token, c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return c, nil
}
