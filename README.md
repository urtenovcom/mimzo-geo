# Mimzo Geo

Compact Russia-focused geo data files for Xray-based VPN clients.

## What this is

Standard `geosite.dat` / `geoip.dat` files used by Xray (and apps based on it like Happ, v2rayN, NekoBox) include **all countries** and weigh ~90 MB combined. iOS limits VPN network extensions to **50 MB of RAM** — so loading full geo data on iPhone causes the tunnel to crash with `XrayCore: memory limit exceeded`.

This repo produces **trimmed-down** versions containing only the entries relevant to Russian routing:

- `geosite-mimzo.dat` — only `category-ru`, `private`, `geolocation-ru` (~1–2 MB)
- `geoip-mimzo.dat` — only `ru`, `private` CIDR blocks (~500 KB)

Total RAM footprint: ~3 MB → fits comfortably under iOS limits.

## How to use

Point your Xray client's geo URLs at:

```
https://raw.githubusercontent.com/urtenovcom/mimzo-geo/main/dist/geosite-mimzo.dat
https://raw.githubusercontent.com/urtenovcom/mimzo-geo/main/dist/geoip-mimzo.dat
```

For Happ routing profile JSON:

```json
{
  "Geositeurl": "https://raw.githubusercontent.com/urtenovcom/mimzo-geo/main/dist/geosite-mimzo.dat",
  "Geoipurl":   "https://raw.githubusercontent.com/urtenovcom/mimzo-geo/main/dist/geoip-mimzo.dat",
  "DirectSites": ["geosite:category-ru", "geosite:private"],
  "DirectIp":    ["geoip:ru", "geoip:private"]
}
```

## How it works

GitHub Actions runs weekly (and on push):

1. Downloads the latest full `geosite.dat` and `geoip.dat` from [runetfreedom/russia-v2ray-rules-dat](https://github.com/runetfreedom/russia-v2ray-rules-dat).
2. Decodes the protobuf, filters out everything except categories listed in `categories.json`.
3. Re-encodes into binary `.dat` files.
4. Commits the result to `dist/`.

Source: [`filter/main.go`](filter/main.go).

## Built for

[Mimzo VPN](https://github.com/urtenovcom) — RU-aware VPN service.

## License

MIT.
