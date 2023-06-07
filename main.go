// Package plugindemo a demo plugin.
package traefik_ip2region

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	cache "github.com/patrickmn/go-cache"
)

const (
	// RealIPHeader real ip header.
	RealIPHeader       = "X-Real-IP"
	DefaultCacheExpire = 60 * time.Minute
	DefaultCachePurge  = 12 * time.Hour
)

// Headers part of the configuration
type Headers struct {
	Country  string `yaml:"country"`
	Province string `yaml:"province"`
	City     string `yaml:"city"`
	ISP      string `yaml:"isp"`
}

// IpResult Ip result.
type IpResult struct {
	Country  string
	Province string
	City     string
	ISP      string
}

// Config the plugin configuration.
type Config struct {
	DBPath  string   `yaml:"dbPath,omitempty"`
	Headers *Headers `yaml:"headers"`
	Ban     Ban      `yaml:"ban"`
}

// Ban
type Ban struct {
	UserAgent []string `yaml:"userAgent"`
	Country   []string `yaml:"country"`
	City      []string `yaml:"city"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		DBPath:  "ip2region.xdb",
		Headers: &Headers{Country: "X-Ip2Region-Country", Province: "X-Ip2Region-Province", City: "X-Ip2Region-City", ISP: "X-Ip2Region-Isp"},
	}
}

// TraefikIp2Region a Demo plugin.
type TraefikIp2Region struct {
	next     http.Handler
	name     string
	headers  *Headers
	searcher *xdb.Searcher
	cache    *cache.Cache
	ban      Ban
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	// 1、从 dbPath 加载整个 xdb 到内存
	cBuff, err := xdb.LoadContentFromFile(config.DBPath)
	if err != nil {
		fmt.Printf("failed to load content from `%s`: %s\n", config.DBPath, err)
	}

	// 2、用全局的 cBuff 创建完全基于内存的查询对象。
	searcher, err := xdb.NewWithBuffer(cBuff)
	if err != nil {
		fmt.Printf("failed to create searcher with content: %s\n", err)
	}

	return &TraefikIp2Region{
		next:     next,
		name:     name,
		headers:  config.Headers,
		ban:      config.Ban,
		searcher: searcher,
		cache:    cache.New(DefaultCacheExpire, DefaultCachePurge),
	}, nil
}

func (a *TraefikIp2Region) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	// ban UserAgent
	for _, v := range a.ban.UserAgent {
		if v == req.UserAgent() {
			rw.WriteHeader(http.StatusForbidden)
			return
		}
	}

	ipStr := req.Header.Get(RealIPHeader)
	if ipStr == "" {
		ipStr = req.RemoteAddr
		tmp, _, err := net.SplitHostPort(ipStr)
		if err == nil {
			ipStr = tmp
		}
	}

	var (
		result *IpResult
	)

	if c, found := a.cache.Get(ipStr); found {
		result = c.(*IpResult)
	} else {
		// 国家|区域|省份|城市|ISP
		region, err := a.searcher.SearchByStr(ipStr)
		if err == nil {
			data := strings.Split(region, "|")
			result = &IpResult{Country: data[0], Province: data[2], City: data[3], ISP: data[4]}
			a.cache.Set(ipStr, result, cache.DefaultExpiration)
		}
	}

	// ban country
	for _, v := range a.ban.Country {
		if v == result.Country {
			rw.WriteHeader(http.StatusForbidden)
			return
		}
	}

	// ban city
	for _, v := range a.ban.City {
		if v == result.City {
			rw.WriteHeader(http.StatusForbidden)
			return
		}
	}

	a.addHeaders(req, result)

	a.next.ServeHTTP(rw, req)
}

func (a *TraefikIp2Region) addHeaders(req *http.Request, result *IpResult) {
	if result != nil {
		req.Header.Add(a.headers.Country, result.Country)
		req.Header.Add(a.headers.Province, result.Province)
		req.Header.Add(a.headers.City, result.City)
		req.Header.Add(a.headers.ISP, result.ISP)
	} else {
		req.Header.Add(a.headers.Country, "NotFound")
		req.Header.Add(a.headers.Province, "NotFound")
		req.Header.Add(a.headers.City, "NotFound")
		req.Header.Add(a.headers.ISP, "NotFound")
	}

}
