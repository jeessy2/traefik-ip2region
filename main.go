// Package plugindemo a demo plugin.
package traefik_ip2region

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

// searcher cached
var searcher *xdb.Searcher

// Headers part of the configuration
type Headers struct {
	Country  string `yaml:"country"`
	Province string `yaml:"province"`
	City     string `yaml:"city"`
	ISP      string `yaml:"isp"`
}

// Config the plugin configuration.
type Config struct {
	DBPath       string   `yaml:"dbPath,omitempty"`
	Headers      *Headers `yaml:"headers"`
	Ban          Rules    `yaml:"ban"`
	Whitelist    Rules    `yaml:"whitelist"`
	IpFromHeader string   `yaml:"ipFromHeader,omitempty"`
}

// Rules
type Rules struct {
	Enabled   bool     `yaml:"enabled"`
	Country   []string `yaml:"country"`
	Province  []string `yaml:"province"`
	City      []string `yaml:"city"`
	UserAgent []string `yaml:"userAgent"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		DBPath:       "ip2region.xdb",
		Headers:      &Headers{Country: "X-Ip2region-Country", Province: "X-Ip2region-Province", City: "X-Ip2region-City", ISP: "X-Ip2region-Isp"},
		IpFromHeader: "",
	}
}

// TraefikIp2Region a Demo plugin.
type TraefikIp2Region struct {
	next         http.Handler
	name         string
	headers      *Headers
	ban          Rules
	whitelist    Rules
	ipFromHeader string
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	err := loadXdb(config.DBPath)
	if err != nil {
		return nil, err
	}

	return &TraefikIp2Region{
		next:         next,
		name:         name,
		headers:      config.Headers,
		ban:          config.Ban,
		whitelist:    config.Whitelist,
		ipFromHeader: config.IpFromHeader,
	}, nil
}

func (a *TraefikIp2Region) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	ipStr := getClientIP(req, a.ipFromHeader)

	var data []string = make([]string, 5)

	// 国家|区域|省份|城市|ISP
	region, err := searcher.SearchByStr(ipStr)
	if err == nil {
		data = strings.Split(region, "|")
		if len(data) < 5 {
			// If the data is not enough, fill it with empty strings
			data = make([]string, 5)
		}
	}

	// add headers
	// 国家|区域|省份|城市|ISP
	req.Header.Add(a.headers.Country, data[0])
	req.Header.Add(a.headers.Province, data[2])
	req.Header.Add(a.headers.City, data[3])
	req.Header.Add(a.headers.ISP, data[4])

	// Ban
	if a.ban.Enabled {
		// country
		for _, v := range a.ban.Country {
			if v == data[0] {
				rw.WriteHeader(http.StatusForbidden)
				return
			}
		}

		// province
		for _, v := range a.ban.Province {
			if v == data[2] {
				rw.WriteHeader(http.StatusForbidden)
				return
			}
		}

		// city
		for _, v := range a.ban.City {
			if v == data[3] {
				rw.WriteHeader(http.StatusForbidden)
				return
			}
		}

		// UserAgent
		for _, v := range a.ban.UserAgent {
			if v == req.UserAgent() {
				rw.WriteHeader(http.StatusForbidden)
				return
			}
		}
	}

	// Whitelist
	if a.whitelist.Enabled {
		// country
		for _, v := range a.whitelist.Country {
			if v == data[0] {
				a.next.ServeHTTP(rw, req)
				return
			}
		}

		// province
		for _, v := range a.whitelist.Province {
			if v == data[2] {
				a.next.ServeHTTP(rw, req)
				return
			}
		}

		// city
		for _, v := range a.whitelist.City {
			if v == data[3] {
				a.next.ServeHTTP(rw, req)
				return
			}
		}

		// UserAgent
		for _, v := range a.whitelist.UserAgent {
			if v == req.UserAgent() {
				a.next.ServeHTTP(rw, req)
				return
			}
		}
		// if the ip is not in the whitelist, return 403
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	a.next.ServeHTTP(rw, req)
}

func loadXdb(dbPath string) error {
	if searcher == nil {
		// 1、从 dbPath 加载整个 xdb 到内存
		cBuff, err := xdb.LoadContentFromFile(dbPath)
		if err != nil {
			return fmt.Errorf("failed to load content from `%s`: %s", dbPath, err)
		}

		// 2、用全局的 cBuff 创建完全基于内存的查询对象。
		searcher, err = xdb.NewWithBuffer(cBuff)
		if err != nil {
			return fmt.Errorf("failed to create searcher with content: %s", err)
		}
	}
	return nil
}

func getClientIP(req *http.Request, ipFromHeader string) string {
	if ipFromHeader != "" {
		// Check header first
		forwardedFor := req.Header.Get(ipFromHeader)
		if forwardedFor != "" {
			ips := strings.Split(forwardedFor, ",")
			return strings.TrimSpace(ips[0])
		}
	}

	// If ipFromHeader is not present or retrieval is not enabled, fallback to RemoteAddr
	remoteAddr := req.RemoteAddr
	tmp, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		remoteAddr = tmp
	}
	return remoteAddr
}
