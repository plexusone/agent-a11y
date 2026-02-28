# Production Readiness Tasks

## Overview

This document tracks the work needed to make `a11y-audit-service` production ready.

**Current State**: ~55% Production Ready

---

## Phase 1: Critical (Blocking Production)

### P0-1: Fix go.mod Replace Directives
- [ ] Remove local replace directives for `vibium-go` and `omnillm`
- [ ] Publish dependencies or use proper versioning
- **Effort**: 1 hour
- **Status**: Not Started

### P0-2: Unit Tests
- [x] Add tests for `types` package
- [x] Add tests for `config` package
- [x] Add tests for `wcag` package (rules)
- [x] Add tests for `llm` package (judge)
- [x] Add tests for `audit` package (types)
- [x] Add tests for `auth` package
- [x] Add tests for `crawler` package
- [x] Add tests for `journey` package
- [x] Add tests for `report` package
- [ ] Add tests for `api` package
- **Effort**: 2-3 days
- **Status**: ~90% Complete

### P0-3: Complete WCAG Rules
- [x] Implement `ColorContrastRule` (1.4.3) - luminance calculation with WCAG ratios
- [x] Implement `LinkDistinguishableRule` (1.4.1) - checks for non-color indicators
- [x] Implement `KeyboardAccessRule` (2.1.1) - checks for click handlers and tabindex
- [x] Implement `FocusVisibleRule` (2.4.7) - detects outline suppression
- [x] Implement `DescriptiveHeadingsRule` (2.4.6) - detects vague/empty headings
- [x] Implement `ARIALabelsRule` (4.1.2) - validates ARIA attributes and roles
- [x] Implement `FormInputNamesRule` (4.1.2) - validates accessible names
- **Effort**: 2 days
- **Status**: Complete

### P0-4: Database Layer for Job Persistence
- [ ] Define storage interface
- [ ] Implement SQLite storage backend
- [ ] Implement PostgreSQL storage backend (optional)
- [ ] Add job expiration/cleanup
- [ ] Migrate API server to use storage layer
- **Effort**: 1-2 days
- **Status**: Not Started

### P0-5: API Authentication
- [ ] Add API key authentication middleware
- [ ] Add JWT token support (optional)
- [ ] Secure all API endpoints
- [ ] Add authentication configuration
- **Effort**: 1 day
- **Status**: Not Started

### P0-6: Fix HTML Report Generation
- [ ] Fix `api/server.go` line 252 - returns JSON instead of HTML
- [ ] Ensure report.Writer is used for HTML format
- **Effort**: 2 hours
- **Status**: Not Started

### P0-7: Input Validation
- [ ] Add URL validation for audit targets
- [ ] Add config validation on load
- [ ] Add API request body validation
- [ ] Validate WCAG level values (A/AA/AAA)
- [ ] Add bounds checking for crawler depth/pages
- **Effort**: 4 hours
- **Status**: Not Started

---

## Phase 2: Important (Should Have)

### P1-1: CI/CD Pipeline
- [ ] Add GitHub Actions workflow for build
- [ ] Add GitHub Actions workflow for tests
- [ ] Add GitHub Actions workflow for linting (golangci-lint)
- [ ] Add release workflow with goreleaser
- **Effort**: 4 hours
- **Status**: Not Started

### P1-2: Docker Support
- [ ] Create Dockerfile
- [ ] Create docker-compose.yml for local development
- [ ] Add multi-stage build for smaller images
- [ ] Document container usage
- **Effort**: 4 hours
- **Status**: Not Started

### P1-3: Rate Limiting
- [ ] Add rate limiting middleware to API
- [ ] Configure limits per endpoint
- [ ] Add rate limit headers to responses
- **Effort**: 2 hours
- **Status**: Not Started

### P1-4: Graceful Shutdown
- [ ] Handle SIGINT/SIGTERM properly
- [ ] Wait for in-flight audits to complete
- [ ] Add configurable shutdown timeout
- [ ] Clean up browser resources on shutdown
- **Effort**: 2 hours
- **Status**: Not Started

### P1-5: Structured Logging
- [ ] Add request correlation IDs
- [ ] Add structured fields to all log calls
- [ ] Sanitize sensitive data from logs
- [ ] Add log level configuration
- **Effort**: 4 hours
- **Status**: Not Started

### P1-6: Integration Tests
- [ ] Add API endpoint integration tests
- [ ] Add end-to-end audit tests
- [ ] Add browser automation tests
- [ ] Set up test fixtures and mock servers
- **Effort**: 1-2 days
- **Status**: Not Started

---

## Phase 3: Enhancement (Nice to Have)

### P2-1: Observability
- [ ] Add Prometheus metrics endpoint
- [ ] Track audit duration, success/failure rates
- [ ] Track API request latency
- [ ] Add OpenTelemetry tracing support
- **Effort**: 1 day
- **Status**: Not Started

### P2-2: Webhook Notifications
- [ ] Add webhook configuration
- [ ] Send notifications on audit completion
- [ ] Support retry logic for failed webhooks
- **Effort**: 4 hours
- **Status**: Not Started

### P2-3: Job Queue System
- [ ] Add Redis-based job queue
- [ ] Support job priorities
- [ ] Enable horizontal scaling
- **Effort**: 1 day
- **Status**: Not Started

### P2-4: PDF Report Export
- [ ] Add PDF report format
- [ ] Include charts and visualizations
- [ ] Support custom branding
- **Effort**: 4 hours
- **Status**: Not Started

### P2-5: Browser Pool
- [ ] Implement browser instance pool
- [ ] Support concurrent audits
- [ ] Add resource limits and cleanup
- **Effort**: 1 day
- **Status**: Not Started

### P2-6: Performance Benchmarks
- [ ] Add benchmark tests
- [ ] Profile memory usage
- [ ] Optimize hot paths
- [ ] Document performance characteristics
- **Effort**: 4 hours
- **Status**: Not Started

### P2-7: Enhanced Security
- [ ] Add HTTPS/TLS configuration
- [ ] Implement secrets encryption
- [ ] Tighten CORS configuration
- [ ] Add security headers
- **Effort**: 4 hours
- **Status**: Not Started

---

## Completed Tasks

### 2026-02-27: Unit Tests Added
- Added comprehensive unit tests for 9 packages
- Packages covered: types, config, wcag, llm, audit, auth, crawler, journey, report
- All tests pass with `go test ./...`

### 2026-02-27: WCAG Rules Completed
- Implemented all 7 stub WCAG rules with full DOM inspection
- Rules now perform actual accessibility checks via JavaScript evaluation
- All 18 WCAG rules are now complete

### 2026-02-27: Lint Issues Fixed
- Fixed errcheck issues (unchecked error returns in defer statements)
- Fixed staticcheck issue (unused variable)
- All lint checks pass with `golangci-lint run`

---

## Notes

- **Current WCAG Rules**: 18 total (all complete)
- **Test Coverage**: ~50% (9/11 packages have unit tests)
- **Report Formats**: 5 (JSON, HTML, Markdown, VPAT, WCAG)
- **API Endpoints**: 6 (all functional but unsecured)
