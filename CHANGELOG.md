# Changelog

All notable changes to DevUp will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-01-06

### Added

- 🎉 **Initial public release of DevUp**
- **`devup init` command** - Intelligent project initialization ⭐
  - Automatically detects 9 project types and package managers (npm, go, pip, cargo, maven, gradle, bundle, composer, pipenv)
  - Scans documentation for environment variables and ports
  - Generates complete devup.yaml configuration
  - Interactive and non-interactive modes
- Multi-application management from single YAML configuration
- Process-based service runners with lifecycle management
- Multiple running modes per application
- Service dependency resolution and ordering
- Built-in health checks (HTTP, TCP, exec)
- Lifecycle hooks (pre/post start/stop/install/setup)
- Individual log files per service
- CLI commands: init, list, start, stop, status, env, install, setup, clean, project
- Comprehensive unit tests (54+ tests, all passing)
- Professional documentation suite:
  - README.md with quick start
  - QUICKSTART.md for 5-minute setup
  - TUTORIAL.md comprehensive guide
  - QUICK_REFERENCE.md command cheat sheet
  - docs/INIT_COMMAND.md for auto-initialization
  - ARCHITECTURE.md for technical details
- Working examples for different use cases

### Technical Details

- Test coverage: 82.1% (config), 96.3% (health), 22.6% (service), 17.5% (cmd)
- Dynamic version reading from VERSION file
- Clean codebase ready for public use

### Architecture

- Extensible configuration system
- Service abstraction layer for future service types
- Robust process management with graceful shutdown
- YAML-based configuration with validation
