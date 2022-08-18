package traefik_get_real_ip

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
)

const (
	xRealIP       = "X-Real-Ip"
	xForwardedFor = "X-Forwarded-For"
)

type Proxy struct {
	ProxyHeadername  string `yaml:"proxyHeadername"`
	ProxyHeadervalue string `yaml:"proxyHeadervalue"`
	RealIP           string `yaml:"realIP"`
	OverwriteXFF     bool   `yaml:"overwriteXFF"` // override X-Forwarded-For
}

// Config the plugin configuration.
type Config struct {
	Proxy []Proxy `yaml:"proxy"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// GetRealIP Define plugin
type GetRealIP struct {
	next  http.Handler
	name  string
	proxy []Proxy
}

// New creates and returns a new realip plugin instance.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	fmt.Printf("☃️ All Config：'%v',Proxy Settings len: '%d'\n", config, len(config.Proxy))

	return &GetRealIP{
		next:  next,
		name:  name,
		proxy: config.Proxy,
	}, nil
}

func (g *GetRealIP) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	var realIP string
	for _, proxy := range g.proxy {
		if req.Header.Get(proxy.ProxyHeadername) == "*" || (req.Header.Get(proxy.ProxyHeadername) == proxy.ProxyHeadervalue) {
			fmt.Printf("🐸 Current Proxy：%s\n", proxy.ProxyHeadervalue)

			nIP := req.Header.Get(proxy.RealIP)
			if proxy.RealIP == "RemoteAddr" {
				nIP, _, _ = net.SplitHostPort(req.RemoteAddr)
			}
			forwardedIPs := strings.Split(nIP, ",")
			fmt.Printf("👀 IPs: '%d' detail:'%v'\n", len(forwardedIPs), forwardedIPs)

			for i := 0; i <= len(forwardedIPs)-1; i++ {
				trimmedIP := strings.TrimSpace(forwardedIPs[i])
				excluded := g.excludedIP(trimmedIP)
				fmt.Printf("exluded:%t， currentIP:%s, index:%d\n", excluded, trimmedIP, i)
				if !excluded {
					realIP = trimmedIP
					break
				}
			}
		}
		if realIP != "" {
			if proxy.OverwriteXFF {
				fmt.Println("🐸 Modify XFF to:", realIP)
				req.Header.Del(xForwardedFor)
				req.Header.Set(xForwardedFor, realIP)
				req.Header.Add(xForwardedFor, "127.0.0.0")
			}
			req.Header.Set(xRealIP, realIP)
			break
		}
	}
	g.next.ServeHTTP(rw, req)
}

func (g *GetRealIP) excludedIP(s string) bool {
	ip := net.ParseIP(s)
	return ip == nil
}
