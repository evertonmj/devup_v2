# Changelog

All notable changes to DevUp will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-12-08

### Added
- 🎉 Initial release of DevUp
- Multi-application management from single YAML configuration
- Process-based service runners with lifecycle management
- Multiple running modes per application
- Service dependency resolution and ordering
- Built-in health checks (HTTP, TCP, exec)
- Lifecycle hooks (pre/post start/stop)
- Individual log files per service
- CLI commands:
  - `devup list` - List all applications
  - `devup start` - Start services
  - `devup stop` - Stop services
  - `devup status` - Check service status
- Comprehensive documentation and examples
- Example configurations for migration from Makefile/shell scripts

### Architecture
- Extensible configuration system
- Service abstraction layer for future service types
- Robust process management with graceful shutdown
- YAML-based configuration with validation

### Coming Soon
- Docker service type
- tmux session management
- Install command for dependency management
- Health monitoring dashboard
- Service restart command
- Log tailing command
- Homebrew distribution
