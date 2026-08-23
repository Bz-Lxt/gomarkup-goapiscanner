package scan

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/alkaid/goapiscanner/internal/config"
)

func ResolveTarget(raw string, cfg config.Config) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("base_url 不能为空")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("base_url 非法，需包含 scheme 与 host")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("仅允许 http/https")
	}
	canon := strings.TrimRight(u.Scheme+"://"+u.Host+u.Path, "/")
	pub, _ := url.Parse(cfg.LabPublicURL)
	in, _ := url.Parse(cfg.LabInternalURL)
	if pub != nil && hostEq(u, pub) {
		return strings.TrimRight(cfg.LabInternalURL, "/"), nil
	}
	if in != nil && hostEq(u, in) {
		return strings.TrimRight(cfg.LabInternalURL, "/"), nil
	}
	if cfg.ScanMode != "authorized" {
		return "", fmt.Errorf("当前 SCAN_MODE=lab，仅允许扫描内置靶场 %s", cfg.LabPublicURL)
	}
	host := u.Hostname()
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return canon, nil
	}
	return canon, nil
}

func hostEq(a, b *url.URL) bool {
	if a == nil || b == nil {
		return false
	}
	ha, pa := a.Hostname(), portOf(a)
	hb, pb := b.Hostname(), portOf(b)
	return strings.EqualFold(ha, hb) && pa == pb
}

func portOf(u *url.URL) string {
	if u.Port() != "" {
		return u.Port()
	}
	if u.Scheme == "https" {
		return "443"
	}
	return "80"
}
