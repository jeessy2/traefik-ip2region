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

const (
	// RealIPHeader real ip header.
	RealIPHeader = "X-Real-IP"
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
	next    http.Handler
	name    string
	headers *Headers
	ban     Ban
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	err := loadXdb(config.DBPath)
	if err != nil {
		return nil, err
	}

	return &TraefikIp2Region{
		next:    next,
		name:    name,
		headers: config.Headers,
		ban:     config.Ban,
	}, nil
}

func (a *TraefikIp2Region) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	ipStr := req.Header.Get(RealIPHeader)
	if ipStr == "" {
		ipStr = req.RemoteAddr
		tmp, _, err := net.SplitHostPort(ipStr)
		if err == nil {
			ipStr = tmp
		}
	}

	var data []string = make([]string, 5)

	// 国家|区域|省份|城市|ISP
	region, err := searcher.SearchByStr(ipStr)
	if err == nil {
		data = strings.Split(region, "|")
		if len(data) != 5 {
			data = make([]string, 5)
		}
	}

	// add headers
	// 国家|区域|省份|城市|ISP
	req.Header.Add(a.headers.Country, data[0])
	req.Header.Add(a.headers.Province, data[2])
	req.Header.Add(a.headers.City, data[3])
	req.Header.Add(a.headers.ISP, data[4])

	// ban country
	for _, v := range a.ban.Country {
		if v == data[0] {
			rw.WriteHeader(http.StatusForbidden)
			return
		}
	}

	// ban city
	for _, v := range a.ban.City {
		if v == data[3] {
			rw.WriteHeader(http.StatusForbidden)
			return
		}
	}

	// ban UserAgent
	for _, v := range a.ban.UserAgent {
		if v == req.UserAgent() {
			rw.WriteHeader(http.StatusForbidden)
			return
		}
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
