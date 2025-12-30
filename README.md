# Next-Board

A modern, high-performance proxy management system.

## Repository Structure

This repository contains:

### `/Xboard`
The original PHP-based Xboard source code. This is the reference implementation for understanding the node protocols.

### `/xboard-go`
**A complete Go-based replacement for Xboard** that maintains 100% wire-compatibility with existing Xboard nodes.

Key features:
- ✅ Xboard node protocol compatible (V1 & V2 APIs)
- ✅ Multi-label node management
- ✅ Advanced traffic multipliers
- ✅ Prometheus metrics
- ✅ Telegram bot integration
- ✅ RESTful API
- ✅ Web UI
- ✅ Production-ready

**[→ See full documentation](xboard-go/README.md)**

**[→ Quick start guide](xboard-go/QUICKSTART.md)**

## Quick Start

```bash
cd xboard-go
docker-compose up -d
```

Access the web interface at `http://localhost:8080`

## Documentation

- [Xboard Go README](xboard-go/README.md) - Complete documentation
- [Quick Start Guide](xboard-go/QUICKSTART.md) - Get running in 5 minutes
- [Xboard Compatibility](xboard-go/docs/xboard-compat.md) - Protocol compatibility details
- [Project Summary](xboard-go/PROJECT_SUMMARY.md) - Architecture overview

## Technology Stack

- **Backend**: Go 1.24, Gin, GORM
- **Database**: MariaDB 11.2
- **Monitoring**: Prometheus, Grafana
- **Bot**: Telegram Bot API
- **Deployment**: Docker, Docker Compose

## License

MIT License
