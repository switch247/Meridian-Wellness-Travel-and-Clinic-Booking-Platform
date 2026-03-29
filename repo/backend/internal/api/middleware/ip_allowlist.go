package middleware

import (
	"net"
	"net/http"
	"strings"

	"meridian/backend/internal/api/response"
	"meridian/backend/internal/config"

	"github.com/labstack/echo/v4"
)

type IPAllowlistConfig struct {
	Allow        []string
	TrustProxy   bool
	BypassRoutes map[string]struct{}
}

func IPAllowlist(cfg IPAllowlistConfig) echo.MiddlewareFunc {
	nets := make([]*net.IPNet, 0, len(cfg.Allow))
	for _, rule := range cfg.Allow {
		n, err := config.ParseCIDRorIP(strings.TrimSpace(rule))
		if err == nil && n != nil {
			nets = append(nets, n)
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, bypass := cfg.BypassRoutes[c.Path()]; bypass {
				return next(c)
			}
			ip := requestIP(c, cfg.TrustProxy)
			if ip == nil {
				return response.JSONError(c, http.StatusForbidden, "ip allowlist denied")
			}
			for _, n := range nets {
				if n.Contains(ip) {
					return next(c)
				}
			}
			return response.JSONError(c, http.StatusForbidden, "ip allowlist denied")
		}
	}
}

func requestIP(c echo.Context, trustProxy bool) net.IP {
	if trustProxy {
		xff := strings.TrimSpace(strings.Split(c.Request().Header.Get("X-Forwarded-For"), ",")[0])
		if xff != "" {
			if ip := net.ParseIP(xff); ip != nil {
				return ip
			}
		}
		xri := strings.TrimSpace(c.Request().Header.Get("X-Real-IP"))
		if xri != "" {
			if ip := net.ParseIP(xri); ip != nil {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(c.Request().RemoteAddr)
	if err != nil {
		return net.ParseIP(c.Request().RemoteAddr)
	}
	return net.ParseIP(host)
}
