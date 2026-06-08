// Filters upstream Xray geo .dat files down to RU-only entries.
// Reads filter/categories.json for the keep-list, writes results to dist/.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

type Config struct {
	GeositeCategories  []string `json:"geosite_categories"`
	GeoipCountries     []string `json:"geoip_countries"`
	UpstreamGeositeURL string   `json:"upstream_geosite_url"`
	UpstreamGeoipURL   string   `json:"upstream_geoip_url"`
}

func main() {
	cfg := loadConfig("filter/categories.json")
	if err := os.MkdirAll("dist", 0o755); err != nil {
		die("mkdir dist: %v", err)
	}

	fmt.Println("== Geosite ==")
	processGeosite(cfg.UpstreamGeositeURL, cfg.GeositeCategories, "dist/geosite-mimzo.dat")

	fmt.Println("== GeoIP ==")
	processGeoip(cfg.UpstreamGeoipURL, cfg.GeoipCountries, "dist/geoip-mimzo.dat")

	fmt.Println("Done.")
}

func loadConfig(path string) *Config {
	b, err := os.ReadFile(path)
	if err != nil {
		die("read config: %v", err)
	}
	c := &Config{}
	if err := json.Unmarshal(b, c); err != nil {
		die("parse config: %v", err)
	}
	return c
}

func processGeosite(url string, keepList []string, outPath string) {
	raw := download(url)
	fmt.Printf("downloaded %d bytes from %s\n", len(raw), url)

	list := &pb.GeoSiteList{}
	if err := proto.Unmarshal(raw, list); err != nil {
		die("decode geosite: %v", err)
	}
	fmt.Printf("upstream has %d categories\n", len(list.Entry))

	keep := lowerSet(keepList)
	filtered := &pb.GeoSiteList{}
	totalDomains := 0
	for _, e := range list.Entry {
		if _, ok := keep[strings.ToLower(e.CountryCode)]; !ok {
			continue
		}
		filtered.Entry = append(filtered.Entry, e)
		totalDomains += len(e.Domain)
		fmt.Printf("  kept: %s (%d domains)\n", e.CountryCode, len(e.Domain))
	}

	if len(filtered.Entry) == 0 {
		die("no matching categories — check categories.json against upstream")
	}

	out, err := proto.Marshal(filtered)
	if err != nil {
		die("encode geosite: %v", err)
	}
	if err := os.WriteFile(outPath, out, 0o644); err != nil {
		die("write geosite: %v", err)
	}
	fmt.Printf("wrote %s: %d categories, %d domains, %d bytes\n",
		filepath.Base(outPath), len(filtered.Entry), totalDomains, len(out))
}

func processGeoip(url string, keepList []string, outPath string) {
	raw := download(url)
	fmt.Printf("downloaded %d bytes from %s\n", len(raw), url)

	list := &pb.GeoIPList{}
	if err := proto.Unmarshal(raw, list); err != nil {
		die("decode geoip: %v", err)
	}
	fmt.Printf("upstream has %d country blocks\n", len(list.Entry))

	keep := lowerSet(keepList)
	filtered := &pb.GeoIPList{}
	totalCidrs := 0
	for _, e := range list.Entry {
		if _, ok := keep[strings.ToLower(e.CountryCode)]; !ok {
			continue
		}
		filtered.Entry = append(filtered.Entry, e)
		totalCidrs += len(e.Cidr)
		fmt.Printf("  kept: %s (%d cidrs)\n", e.CountryCode, len(e.Cidr))
	}

	if len(filtered.Entry) == 0 {
		die("no matching countries — check categories.json against upstream")
	}

	out, err := proto.Marshal(filtered)
	if err != nil {
		die("encode geoip: %v", err)
	}
	if err := os.WriteFile(outPath, out, 0o644); err != nil {
		die("write geoip: %v", err)
	}
	fmt.Printf("wrote %s: %d countries, %d cidrs, %d bytes\n",
		filepath.Base(outPath), len(filtered.Entry), totalCidrs, len(out))
}

func download(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		die("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		die("GET %s: status %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		die("read body %s: %v", url, err)
	}
	return body
}

func lowerSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[strings.ToLower(s)] = struct{}{}
	}
	return m
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fatal: "+format+"\n", args...)
	os.Exit(1)
}
