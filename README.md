# go-myip

A lightweight, self-hosted "what is my IP address" service written in Go. Returns your public IP, user agent, and optional geolocation data in HTML, JSON, or plain text.

---

## Features

- Detect client IP with proxy header support (`X-Real-Ip`, `X-Forwarded-For`)
- Optional geolocation via MaxMind GeoLite2 (city, country, continent)
- Content negotiation: HTML, JSON (`application/json`), and plain text (`text/plain`)
- Per-IP rate limiting with configurable thresholds
- Health check endpoint
- Ready to deploy on [Railway](https://railway.app)

---

## Endpoints

| Method | Path      | Description                              |
|--------|-----------|------------------------------------------|
| GET    | `/`       | Returns IP, user agent, and geo info     |
| GET    | `/health` | Health check, returns status and date    |

### Response formats

The `/` endpoint respects the `Accept` header:

**HTML** (default browser view)

**JSON**
```bash
curl -H "Accept: application/json" https://your-domain.com/
```
```json
{
  "ip": "1.2.3.4",
  "ip_public": true,
  "user_agent": "curl/8.0",
  "city": "Amsterdam",
  "country": "Netherlands",
  "continent": "Europe",
  "country_code": "NL",
  "continent_code": "EU"
}
```

**Plain text**
```bash
curl -H "Accept: text/plain" https://your-domain.com/
```
```
IP: 1.2.3.4
User-Agent: curl/8.0
City: Amsterdam
Country: Netherlands
Continent: Europe
Country Code: NL
Continent Code: EU
```

---

## Getting Started

### Prerequisites

- Go 1.21+
- (Optional) MaxMind license key for GeoIP

### Run locally

```bash
git clone https://github.com/ToneSploit/go-myip.git
cd go-myip
cp .env.example .env
# Edit .env with your configuration
go run main.go
```

The server starts on port `8080`.

---

## Configuration

Copy `.env.example` to `.env` and adjust the values:

| Variable                              | Default | Description                                      |
|---------------------------------------|---------|--------------------------------------------------|
| `GEOIP_ENABLED`                       | `false` | Enable geolocation lookups via GeoLite2          |
| `MAXMIND_ACCOUNT_ID`                  |         | MaxMind account ID (required if GeoIP enabled)   |
| `MAXMIND_LICENSE_KEY`                 |         | MaxMind license key (required if GeoIP enabled)  |
| `APP_RATE_LIMIT_REQUESTS_PER_MINUTE`  | `20`    | Max requests per minute per IP                   |
| `APP_RATE_LIMIT_BURST`                | `5`     | Burst size for rate limiter                      |
| `APP_RATE_LIMIT_EXPIRES_SECONDS`      | `360`   | Rate limiter memory expiry in seconds            |
| `TINYLYTICS`                          |         | Optional Tinylytics site ID for analytics        |

---

## Deploy on Railway

This project includes `railway.toml` and `nixpacks.toml` for zero-config Railway deployments.

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/new/template)

1. Fork this repository
2. Create a new project on Railway and connect your fork
3. Add your environment variables in the Railway dashboard
4. Railway will build and deploy automatically

The `RAILWAY_PUBLIC_DOMAIN` environment variable is picked up automatically when running on Railway.

---

## Project Structure

```
go-myip/
├── main.go           # Application entry point, routing, handlers
├── logger/           # Structured logging (zap)
├── shared/           # GeoIP lookup and environment setup
├── templates/        # HTML templates
├── .env.example      # Example configuration
├── nixpacks.toml     # Nixpacks build config
└── railway.toml      # Railway deployment config
```

---

## Tech Stack

- [Echo](https://echo.labstack.com/) - HTTP framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Zap](https://github.com/uber-go/zap) - Structured logging
- [MaxMind GeoLite2](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) - IP geolocation (optional)

---

## License

MIT