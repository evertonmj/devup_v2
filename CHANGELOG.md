# Changelog

All notable changes to DevUp will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **`devup init` command** - Intelligent project initialization ⭐
  - Automatically detects 9 project types and package managers
  - Scans documentation for environment variables and ports
  - Generates complete devup.yaml configuration
  - Interactive and non-interactive modes
- Comprehensive unit tests (54+ tests, all passing)
- Professional documentation suite
- Init command documentation (docs/INIT_COMMAND.md)

### Changed
- Updated all documentation to include init command
- Enhanced TUTORIAL.md with quick start section
- Improved QUICKSTART.md with auto-initialization

### Fixed
- Race condition in health check exec command
- Module path issues

## [1.0.0] - 2024-12-08

### Added
- 🎉 Initial release of DevUp
- Multi-application management from single YAML configuration
- Process-based service runners with lifecycle management
- Multiple running modes per application
- Service dependency resolution and ordering
- Built-in health checks (HTTP, TCP, exec)
- Lifecycle hooks (pre/post start/stop/install/setup)
- Individual log files per service
- CLI commands: list, start, stop, status, env, install, setup, clean, project
- Comprehensive documentation and examples
- 82.1% config coverage, 96.3% health coverage

### Architecture
- Extensible configuration system
- Service abstraction layer for future service types
- Robust process management with graceful shutdown
- YAML-based configuration with validation
