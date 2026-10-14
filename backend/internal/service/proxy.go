package service

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	FallbackModeNone   = "none"
	FallbackModeProxy  = "proxy"
	FallbackModeDirect = "direct"
)

type Proxy struct {
	ID             int64
	Name           string
	Protocol       string
	Host           string
	Port           int
	Username       string
	Password       string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ExpiresAt      *time.Time
	FallbackMode   string
	BackupProxyID  *int64
	ExpiryWarnDays int
}

func (p *Proxy) IsActive() bool {
	return p.Status == StatusActive
}

// IsExpired 报告代理是否已过期（基于 expires_at，与 status 无关）。
func (p *Proxy) IsExpired(now time.Time) bool {
	return p.ExpiresAt != nil && !p.ExpiresAt.After(now)
}

func (p *Proxy) URL() string {
	u := &url.URL{
		Scheme: p.Protocol,
		Host:   net.JoinHostPort(p.Host, strconv.Itoa(p.Port)),
	}
	if p.Username != "" && p.Password != "" {
		u.User = url.UserPassword(p.Username, p.Password)
	}
	return u.String()
}

// ResolveAccountProxyURL enforces the proxy invariant for upstream requests.
// A configured ProxyID must never silently degrade to a direct connection when
// the proxy relation was omitted from an account snapshot or the proxy data is
// malformed. Direct connections remain available only for accounts without a
// configured proxy (ProxyID == nil).
func ResolveAccountProxyURL(account *Account) (string, error) {
	if account == nil || account.ProxyID == nil {
		return "", nil
	}
	if account.Proxy == nil {
		return "", fmt.Errorf("account %d has proxy_id %d but proxy details are unavailable", account.ID, *account.ProxyID)
	}
	proxyURL := strings.TrimSpace(account.Proxy.URL())
	parsed, err := url.Parse(proxyURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("account %d proxy %d is invalid", account.ID, *account.ProxyID)
	}
	return proxyURL, nil
}

type ProxyWithAccountCount struct {
	Proxy
	AccountCount   int64
	LatencyMs      *int64
	LatencyStatus  string
	LatencyMessage string
	IPAddress      string
	Country        string
	CountryCode    string
	Region         string
	City           string
	QualityStatus  string
	QualityScore   *int
	QualityGrade   string
	QualitySummary string
	QualityChecked *int64
}

type ProxyAccountSummary struct {
	ID       int64
	Name     string
	Platform string
	Type     string
	Notes    *string
}
