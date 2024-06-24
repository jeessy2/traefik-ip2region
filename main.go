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
	DBPath    string   `yaml:"dbPath,omitempty"`
	Headers   *Headers `yaml:"headers"`
	Ban       Rules    `yaml:"ban"`
	Whitelist Rules    `yaml:"whitelist"`
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
		DBPath:  "ip2region.xdb",
		Headers: &Headers{Country: "X-Ip2region-Country", Province: "X-Ip2region-Province", City: "X-Ip2region-City", ISP: "X-Ip2region-Isp"},
	}
}

// TraefikIp2Region a Demo plugin.
type TraefikIp2Region struct {
	Next      http.Handler
	Name      string
	Headers   *Headers
	Ban       Rules
	Whitelist Rules
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	err := loadXdb(config.DBPath)
	if err != nil {
		return nil, err
	}

	return &TraefikIp2Region{
		Next:      next,
		Name:      name,
		Headers:   config.Headers,
		Ban:       config.Ban,
		Whitelist: config.Whitelist,
	}, nil
}

func (a *TraefikIp2Region) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	ipStr := req.RemoteAddr
	tmp, _, err := net.SplitHostPort(ipStr)
	if err == nil {
		ipStr = tmp
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
	req.Header.Add(a.Headers.Country, data[0])
	req.Header.Add(a.Headers.Province, data[2])
	req.Header.Add(a.Headers.City, data[3])
	req.Header.Add(a.Headers.ISP, data[4])

	// Ban
	if a.Ban.Enabled {
		// country
		for _, v := range a.Ban.Country {
			if v == data[0] {
				rw.WriteHeader(http.StatusForbidden)
				return
			}
		}

		// province
		for _, v := range a.Ban.Province {
			if v == data[2] {
				rw.WriteHeader(http.StatusForbidden)
				return
			}
		}

		// city
		for _, v := range a.Ban.City {
			if v == data[3] {
				rw.WriteHeader(http.StatusForbidden)
				return
			}
		}

		// UserAgent
		for _, v := range a.Ban.UserAgent {
			if v == req.UserAgent() {
				rw.WriteHeader(http.StatusForbidden)
				return
			}
		}
	}

	// Whitelist
	if a.Whitelist.Enabled {
		// country
		for _, v := range a.Whitelist.Country {
			if v == data[0] {
				a.Next.ServeHTTP(rw, req)
				return
			}
		}

		// province
		for _, v := range a.Whitelist.Province {
			if v == data[2] {
				a.Next.ServeHTTP(rw, req)
				return
			}
		}

		// city
		for _, v := range a.Whitelist.City {
			if v == data[3] {
				a.Next.ServeHTTP(rw, req)
				return
			}
		}

		// UserAgent
		for _, v := range a.Whitelist.UserAgent {
			if v == req.UserAgent() {
				a.Next.ServeHTTP(rw, req)
				return
			}
		}
		// if the ip is not in the whitelist, return 403
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	a.Next.ServeHTTP(rw, req)
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
