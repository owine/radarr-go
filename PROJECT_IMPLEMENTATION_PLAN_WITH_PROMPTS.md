# 🎯 **Radarr-Go Project Implementation Plan with Agent Prompts**

*Detailed Agent Assignment and Execution Strategy with Actionable Prompts*

## 📋 **Executive Summary**

This implementation plan translates the high-level PROJECT_ROADMAP.md into specific, actionable tasks assigned to specialized Claude Code agents. Each phase includes detailed task assignments, agent responsibilities, dependencies, and **specific prompts for agent invocation**.

**Plan Structure:**

- **Agent-Specific Task Assignments** with clear scope boundaries
- **Actionable Agent Prompts** for each task with complete context
- **Dependency Management** between tasks and agents
- **Quality Gates** and review checkpoints
- **Parallel Execution Opportunities** for faster delivery
- **Risk Mitigation** through agent specialization

---

## 👥 **Agent Roles and Responsibilities**

### **Primary Development Agents**

```yaml
backend-architect:
  - API design and backend architecture
  - Service implementation and database design
  - Performance optimization and scalability
  - Integration patterns and middleware

frontend-developer:
  - React/TypeScript component development
  - UI/UX implementation and responsive design
  - State management and API integration
  - Performance optimization for large datasets

database-admin:
  - Schema design and migration management
  - Performance tuning and optimization
  - Backup and recovery procedures
  - Multi-database compatibility

devops-troubleshooter:
  - CI/CD pipeline management
  - Deployment automation and monitoring
  - Infrastructure as code
  - Performance and security monitoring
```

### **Specialized Support Agents**

```yaml
docs-architect:
  - Documentation strategy and structure
  - User guides and API documentation
  - Developer onboarding materials
  - Community engagement content

test-automator:
  - Test strategy and framework setup
  - Unit, integration, and E2E test implementation
  - Performance and load testing
  - Quality assurance automation

security-auditor:
  - Security architecture review
  - Vulnerability assessment and remediation
  - Authentication and authorization
  - Security best practices implementation

golang-pro:
  - Go code quality and best practices
  - Performance optimization
  - Concurrency and memory management
  - Code architecture review
```

---

## 🚀 **Phase 0: Immediate Critical Fixes**

**Duration**: 2-3 weeks | **Agents**: 3 primary + 2 support

### **Sprint 0.1: Critical Backend Fixes** (Week 1)

#### **Task 0.1.1: Database Migration Fixes** ✅ **COMPLETED**

**Agent**: `database-admin`

**Priority**: CRITICAL

**Duration**: 2-3 days

**STATUS**: ✅ **COMPLETED** - Database migrations are properly structured, migration 007 (wanted_movies) exists with correct foreign key references, cross-database compatibility verified.

**Agent Prompt:**

```text
Please review and fix critical issues in the radarr-go database migrations. The project has 8 migration files that need validation and correction.

Current Issue: The project review identified a critical foreign key reference error in migration 007 (wanted_movies table) that references a non-existent quality_definitions table when it should reference quality_profiles table instead.

Please examine:
- /migrations/postgres/007_wanted_movies.up.sql and .down.sql
- /migrations/mysql/007_wanted_movies.up.sql and .down.sql
- All other migration files for consistency and rollback safety

Tasks to complete:
1. Fix the foreign key reference error in migration 007
2. Validate all up/down migration pairs work correctly without data loss
3. Add missing database constraints and indexes for performance
4. Test migration execution performance on datasets with 10k+ movies
5. Create database operations scripts for backup/restore and monitoring

Focus on cross-database compatibility (PostgreSQL vs MySQL) and provide comprehensive validation that all migrations are production-ready.
```

**Deliverables:**

- Fixed migration 007 with correct foreign key references
- Validated rollback procedures for all 8 migrations
- Performance benchmarks for migration execution
- Database operations scripts (backup, monitoring, disaster recovery)

**Dependencies:** None (can start immediately)

**Quality Gates:**

- All migrations pass without errors on PostgreSQL and MySQL
- Rollback procedures tested and documented
- No data loss scenarios identified

#### **Task 0.1.2: Code Quality Fixes** ✅ **COMPLETED**

**Agent**: `golang-pro`

**Priority**: CRITICAL

**Duration**: 2-3 days

**STATUS**: ✅ **COMPLETED** - `make lint` returns 0 issues, all critical linting issues have been resolved.

**Agent Prompt:**

```text
Please conduct a comprehensive Go code quality review and fix all critical linting issues in the radarr-go project.

The project review identified several critical code quality issues that need immediate attention:

1. **Exhaustive Switch Statements**: 4 locations with missing cases in enum switches (CalendarEventType, HealthStatus, FileOperation switches)
2. **Error Handling**: Critical errcheck linter issues with unchecked error returns
3. **Unused Code**: Remove unused functions and variables flagged by linters
4. **Nil Pointer Safety**: Resolve potential nil pointer dereferences in health checker code

Please examine the entire codebase focusing on:
- `internal/services/health_checkers.go` - nil pointer issues
- `internal/services/calendar_service.go` - exhaustive switches
- `internal/services/file_organization_service.go` - error handling
- All other Go files flagged by golangci-lint

Goals:
- Achieve zero critical linter warnings when running `make lint`
- Maintain or improve current test coverage
- Follow Go best practices and idioms throughout
- Update Go module dependencies to latest compatible versions

Provide a detailed report of all changes made and ensure the code follows modern Go practices.
```

**Deliverables:**

- Zero critical linter warnings
- All exhaustive switch statements completed
- Nil pointer safety in all health check code
- Updated go.mod with latest compatible versions

**Dependencies:** None (can run parallel with database fixes)

**Quality Gates:**

- `make lint` executes with zero critical issues
- All tests pass without warnings
- Security scan shows no critical vulnerabilities

#### **Task 0.1.3: Testing Infrastructure Setup** ⚠️ **NEEDS COMPLETION**

**Agent**: `test-automator`

**Priority**: HIGH

**Duration**: 3-4 days

**STATUS**: ⚠️ **NEEDS COMPLETION** - docker-compose test infrastructure missing, `make test` fails due to missing test database containers.

**Agent Prompt:**

```text
Please establish a comprehensive testing infrastructure for radarr-go that enables reliable integration and performance testing.

Current Issue: The project has excellent unit tests but integration tests are currently skipped due to missing database setup. Many tests show "Database tests require PostgreSQL or MariaDB setup - skipping".

Please implement:

1. **Test Database Containers**: Set up Docker Compose configurations for PostgreSQL and MySQL test databases
2. **Integration Test Enablement**: Modify test helpers to use containerized databases instead of skipping
3. **Benchmark Test Fixes**: Enable the currently skipped benchmark tests with proper database setup
4. **Test Data Management**: Create fixtures and helpers for consistent test data

Examine these key files:
- `internal/services/test_helpers.go` - currently skipping database tests
- All `*_test.go` files with skipped database-dependent tests
- Benchmark tests that are currently disabled

Requirements:
- Tests should run in isolated environments
- Support both PostgreSQL and MySQL testing
- Maintain current test coverage while enabling integration tests
- Provide consistent performance benchmarks

Create a testing strategy that works for both local development and CI/CD environments.
```

**Deliverables:**
- Docker Compose setup for test databases
- Integration test suite with >80% pass rate
- Benchmark tests executable with consistent results
- Test data management system

**Dependencies:** Database migration fixes (Task 0.1.1)

**Quality Gates:**
- `make test` runs all tests including integration tests
- Benchmark tests provide consistent performance metrics
- Test coverage maintained above current levels

### **Sprint 0.2: Foundation Preparation** (Weeks 2-3)

#### **Task 0.2.1: Development Environment Enhancement**
**Agent**: `devops-troubleshooter`
**Priority**: MEDIUM
**Duration**: 3-4 days

**Agent Prompt:**
```text
Please enhance the radarr-go development environment to prepare for frontend integration and improve developer experience.

Current State: The project has a solid Go backend with excellent CI/CD, but needs frontend build integration and improved development workflow for the upcoming React frontend implementation.

Please enhance:

1. **Makefile Enhancement**: Add frontend build targets that will integrate with the upcoming React application
   - `make build-frontend` - Build React frontend
   - `make dev-frontend` - Start frontend development server
   - `make build-all` - Build both frontend and backend

2. **Development Docker Composition**: Create docker-compose.dev.yml for full development environment including:
   - PostgreSQL and MySQL databases
   - Go backend with hot reload
   - Frontend development server (placeholder for future React app)
   - Development monitoring and debugging tools

3. **Environment Setup Guide**: Create documentation for new developers to get started quickly

4. **Development Monitoring**: Add development-specific monitoring and debugging capabilities

Focus on creating a seamless development experience that will support both backend and frontend developers working on the project.
```

**Deliverables:**
- Enhanced Makefile with frontend integration targets
- Docker Compose for full development environment
- Developer onboarding documentation
- Development monitoring dashboard

**Dependencies:** Testing infrastructure (Task 0.1.3)

**Quality Gates:**
- New developers can set up environment in <30 minutes
- All build targets work consistently across platforms
- Development environment matches production architecture

#### **Task 0.2.2: Release and Documentation Foundation**
**Agent**: `docs-architect`
**Priority**: MEDIUM
**Duration**: 4-5 days

**Agent Prompt:**
```text
Please create foundational documentation for radarr-go that will support the upcoming major feature development and user adoption.

Current State: The project has excellent implementation (95% feature parity with original Radarr) but documentation significantly lags behind, blocking user adoption.

Please create:

1. **Architecture Documentation Update**: Update CLAUDE.md to reflect the current comprehensive feature set including:
   - Task scheduling system
   - File organization capabilities
   - Notification providers (11 providers)
   - Health monitoring system
   - Calendar and scheduling features
   - Wanted movies management
   - Movie collections support

2. **API Endpoint Inventory**: Create comprehensive catalog of all 150+ implemented API endpoints with:
   - Endpoint descriptions and purposes
   - Request/response examples
   - Authentication requirements
   - Rate limiting information

3. **Configuration Reference**: Document all configuration options with:
   - Default values and examples
   - Environment variable overrides
   - Database-specific configurations
   - Security and performance settings

4. **Release Preparation**: Prepare materials for v0.9.0-alpha release:
   - Comprehensive changelog highlighting new features
   - Release notes explaining current capabilities
   - Known limitations and upcoming features

Focus on creating documentation that showcases the extensive functionality already implemented and guides users toward successful adoption.
```

**Deliverables:**
- Updated architecture documentation
- Complete API endpoint catalog
- Configuration reference guide
- Release notes and changelog for v0.9.0-alpha

**Dependencies:** Code quality fixes (Task 0.1.2)

**Quality Gates:**
- Documentation accurately reflects current codebase
- API inventory includes all implemented endpoints
- Configuration examples are tested and valid

---

## 📚 **Phase 1: Documentation and User Experience**
**Duration**: 4-6 weeks | **Agents**: 2 primary + 3 support

### **Sprint 1.1: Core User Documentation** (Weeks 1-2)

#### **Task 1.1.1: Installation and Setup Documentation**
**Agent**: `docs-architect`
**Priority**: HIGH
**Duration**: 5-6 days

**Agent Prompt:**
```text
Please create comprehensive installation and setup documentation that enables users to successfully deploy radarr-go in production environments.

Context: Radarr-go has achieved 95% feature parity with the original Radarr but lacks user-facing documentation. Users need clear guidance to migrate from the C# version and set up the Go version successfully.

Please create:

1. **Multi-Platform Installation Guide** covering:
   - Docker installation and docker-compose examples
   - Binary installation for Linux, macOS, Windows, and FreeBSD
   - Source compilation with Go 1.24+ requirements
   - Systemd service configuration for Linux
   - Windows service setup guide

2. **Database Setup Instructions** with:
   - PostgreSQL setup and configuration (recommended)
   - MariaDB/MySQL setup as alternative
   - Database performance tuning recommendations
   - Connection pooling configuration
   - Migration from SQLite (original Radarr database)

3. **Configuration Guide** including:
   - Complete config.yaml example with all sections
   - Environment variable override examples
   - Security configuration (API keys, CORS, SSL)
   - Performance tuning for different library sizes

4. **Migration Guide from Original Radarr** with:
   - Step-by-step migration process
   - Data backup recommendations
   - Configuration conversion guide
   - Common migration issues and solutions

Focus on creating documentation that enables successful production deployment within 30 minutes for experienced users.
```

**Deliverables:**
- Multi-platform installation guide
- Database setup automation scripts
- Complete configuration reference
- Step-by-step migration guide with tools

**Dependencies:** Development environment setup (Task 0.2.1)

**Quality Gates:**
- Installation guide tested on 3 platforms
- Database setup scripts work on fresh systems
- Migration guide validated with real Radarr data

#### **Task 1.1.2: Feature Documentation**
**Agent**: `docs-architect` + `backend-architect` (review)
**Priority**: HIGH
**Duration**: 7-8 days

**Agent Prompt:**
```text
Please create comprehensive feature documentation for all major radarr-go capabilities, focusing on the advanced features that differentiate it from the original Radarr.

Context: Radarr-go includes sophisticated features like task scheduling, advanced notifications, health monitoring, and file organization that need clear documentation for user adoption.

Please document:

1. **Task Scheduling System**:
   - How to configure automated tasks (movie refresh, import list sync, health checks)
   - Task priority and scheduling options
   - Background job monitoring and cancellation
   - Custom task creation and management
   - Task performance optimization

2. **Notification System** (11 providers):
   - Setup guides for each provider: Discord, Slack, Email, Webhook, Pushover, Telegram, Pushbullet, Gotify, Mailgun, SendGrid, Custom Scripts
   - Template customization and variable substitution
   - Event trigger configuration
   - Notification testing and troubleshooting
   - Advanced notification workflows

3. **File Organization System**:
   - Automated file moving and renaming configuration
   - Naming template system with available tokens
   - Hard link vs copy vs move strategies
   - Import decision workflows
   - File conflict resolution

4. **Health Monitoring**:
   - System health checks and issue detection
   - Performance metrics monitoring
   - Automated issue resolution
   - Custom health check configuration
   - Integration with notification system

Each feature should include practical examples, common use cases, and troubleshooting sections.
```

**Deliverables:**
- Interactive task scheduling tutorial
- Provider-specific notification setup guides
- File organization best practices guide
- Health monitoring dashboard documentation

**Dependencies:** API endpoint inventory (Task 0.2.2)

**Quality Gates:**
- Each feature guide includes working examples
- All notification providers tested and documented
- File organization examples validated

### **Sprint 1.2: API Documentation** (Weeks 3-4)

#### **Task 1.2.1: OpenAPI Specification Generation**
**Agent**: `backend-architect`
**Priority**: HIGH
**Duration**: 6-7 days

**Agent Prompt:**
```text
Please create comprehensive API documentation for radarr-go's 150+ endpoints, focusing on developer integration and third-party client development.

Context: Radarr-go maintains 100% API compatibility with Radarr v3 while adding new endpoints for advanced features. Developers need complete API documentation to build integrations and clients.

Please implement:

1. **OpenAPI 3.0 Specification Generation**:
   - Examine all API handlers in `internal/api/` directory
   - Generate complete OpenAPI spec covering all 150+ endpoints
   - Include request/response schemas with examples
   - Document all authentication mechanisms (API key, session)

2. **Interactive API Documentation**:
   - Integrate Swagger UI for interactive testing
   - Add "Try it out" functionality for all endpoints
   - Include authentication examples and setup
   - Provide realistic request/response examples

3. **API Compatibility Documentation**:
   - Document Radarr v3 API compatibility guarantee
   - Highlight new endpoints and enhanced features
   - Provide migration guide for existing API clients
   - Document rate limiting and pagination patterns

4. **Developer Integration Guide**:
   - Authentication setup examples
   - Common integration patterns
   - Error handling best practices
   - WebSocket real-time updates documentation

Focus on creating documentation that enables rapid integration development and showcases the comprehensive API surface area.
```

**Deliverables:**
- Complete OpenAPI 3.0 specification
- Interactive Swagger UI integration
- Authentication flow documentation
- Error response catalog with examples

**Dependencies:** None (can start with Phase 1)

**Quality Gates:**
- All endpoints documented with request/response examples
- Interactive documentation functional
- Authentication examples tested

#### **Task 1.2.2: Integration Guides and Examples**
**Agent**: `docs-architect` + `backend-architect` (examples)
**Priority**: MEDIUM
**Duration**: 5-6 days

**Agent Prompt:**
```text
Please create practical integration guides and examples that demonstrate how to effectively use the radarr-go API for common use cases.

Context: While the OpenAPI documentation provides complete technical reference, developers need practical examples and integration patterns for common scenarios.

Please create:

1. **Third-Party Client Integration Examples**:
   - Python client example using requests library
   - JavaScript/Node.js client with axios
   - Shell script examples using curl
   - PowerShell examples for Windows automation
   - Go client library example

2. **Common Integration Patterns**:
   - Movie search and addition workflow
   - Queue monitoring and management
   - Bulk operations for large libraries
   - Real-time event handling via WebSocket
   - Custom notification webhook implementation

3. **Automation Examples**:
   - Backup and restore API workflows
   - Library maintenance automation scripts
   - Custom quality management workflows
   - Integration with external tools (Plex, Jellyfin, etc.)

4. **Troubleshooting Guide**:
   - Common API errors and solutions
   - Authentication troubleshooting
   - Rate limiting and performance optimization
   - WebSocket connection issues
   - CORS and cross-origin request handling

Each example should be complete, tested, and include error handling.
```

**Deliverables:**
- Client integration examples (Python, JavaScript, etc.)
- Webhook payload examples and testing tools
- Automated backup/restore scripts
- Troubleshooting decision tree

**Dependencies:** OpenAPI specification (Task 1.2.1)

**Quality Gates:**
- Integration examples tested with real clients
- Backup/restore procedures validated
- Troubleshooting guide covers 90% of common issues

### **Sprint 1.3: Developer Documentation** (Weeks 5-6)

#### **Task 1.3.1: Architecture and Development Guides**
**Agent**: `docs-architect` + `golang-pro` (technical review)
**Priority**: MEDIUM
**Duration**: 6-7 days

**Agent Prompt:**
```text
Please create comprehensive developer documentation that enables new contributors to understand the radarr-go architecture and contribute effectively.

Context: Radarr-go demonstrates excellent Go architecture with sophisticated patterns like dependency injection, worker pools, and service containers. New developers need guidance to maintain this quality.

Please create:

1. **Architecture Deep-Dive Documentation**:
   - System architecture diagrams showing service relationships
   - Dependency injection patterns and service container usage
   - Database architecture with GORM optimizations
   - Worker pool and task scheduling system design
   - Real-time update system with WebSocket integration

2. **Code Contribution Guidelines**:
   - Go code style and conventions used in the project
   - Testing requirements and patterns
   - Code review process and criteria
   - Git workflow and commit message standards
   - Performance and security considerations

3. **Testing Strategy Documentation**:
   - Unit testing with mocks and interfaces
   - Integration testing with database containers
   - Benchmark testing for performance regression detection
   - End-to-end testing strategies
   - Test data management and fixtures

4. **Extension Development Guide**:
   - How to add new notification providers
   - Creating custom task handlers
   - Extending the health monitoring system
   - Adding new API endpoints following project patterns
   - Database migration best practices

Focus on maintaining the high code quality standards while enabling rapid contributor onboarding.
```

**Deliverables:**
- System architecture diagrams and explanations
- Code contribution workflow documentation
- Testing best practices guide
- Extension development framework

**Dependencies:** Testing infrastructure (Task 0.1.3)

**Quality Gates:**
- Architecture documentation matches current implementation
- Contributing guidelines enable new developer onboarding
- Testing examples are functional and educational

#### **Task 1.3.2: Operations and Deployment Documentation**
**Agent**: `devops-troubleshooter`
**Priority**: MEDIUM
**Duration**: 5-6 days

**Agent Prompt:**
```text
Please create comprehensive operations and deployment documentation for production radarr-go deployments.

Context: Radarr-go offers significant operational advantages over the original .NET version (single binary, lower resource usage, better performance) but needs proper production deployment guidance.

Please create:

1. **Production Deployment Guide**:
   - Production-ready Docker Compose configurations
   - Kubernetes deployment manifests and best practices
   - Reverse proxy configuration (nginx, Apache, Traefik)
   - SSL/TLS setup and certificate management
   - Environment-specific configuration management

2. **Monitoring and Alerting Setup**:
   - Prometheus metrics collection setup
   - Grafana dashboard templates for system monitoring
   - Log aggregation with structured logging
   - Performance monitoring and alerting thresholds
   - Health check endpoint integration

3. **Performance Tuning Guide**:
   - Database connection pooling optimization
   - Go runtime tuning (GOMAXPROCS, GC settings)
   - Memory usage optimization for large libraries
   - Concurrent download and processing tuning
   - Storage performance considerations

4. **Security Hardening Recommendations**:
   - Network security and firewall configuration
   - Authentication and authorization best practices
   - API key management and rotation
   - Database security hardening
   - Container security best practices

Include automated deployment scripts and monitoring templates that work out of the box.
```

**Deliverables:**
- Production deployment automation scripts
- Monitoring dashboard templates
- Performance optimization playbook
- Security hardening checklist

**Dependencies:** Development environment setup (Task 0.2.1)

**Quality Gates:**
- Deployment scripts tested on production-like environments
- Monitoring templates functional with sample data
- Security recommendations validated

---

## 🎨 **Phase 2: Frontend Foundation**
**Duration**: 8-10 weeks | **Agents**: 3 primary + 2 support

### **Sprint 2.1: Frontend Architecture Setup** (Weeks 1-2)

#### **Task 2.1.1: Project Structure and Build System**
**Agent**: `frontend-developer`
**Priority**: CRITICAL
**Duration**: 5-6 days

**Agent Prompt:**
```text
Please set up the frontend foundation for radarr-go, creating a modern React application that will provide the user interface for all the comprehensive backend functionality.

Context: Radarr-go has excellent backend functionality (95% feature parity with original Radarr) but completely lacks a user interface. We need to create a React/TypeScript frontend that matches the original Radarr's UI capabilities.

Please implement:

1. **React 18 + TypeScript Project Setup**:
   - Create project in `web/frontend/` directory
   - Configure TypeScript with strict settings
   - Set up modern React 18 with concurrent features
   - Configure ESLint and Prettier for code quality

2. **Vite Build System Configuration**:
   - Configure Vite for fast development builds
   - Set up proxy to Go backend at localhost:7878
   - Configure production build optimization
   - Set up environment variable handling

3. **Redux Toolkit + RTK Query Setup**:
   - Configure Redux store with RTK
   - Set up RTK Query for API integration with radarr-go's 150+ endpoints
   - Configure automatic API tag invalidation
   - Set up optimistic updates for user actions

4. **CSS Pipeline with PostCSS**:
   - Set up CSS Modules for component styling
   - Configure PostCSS with modern CSS features
   - Set up design token system for consistent theming
   - Configure responsive design utilities

The goal is to create a solid foundation that can support the comprehensive movie management interface we'll build in subsequent sprints.
```

**Deliverables:**
- Functional React development environment
- Vite configuration with proxy to Go backend
- Redux store with RTK Query integration
- CSS pipeline with design tokens

**Dependencies:** Documentation foundation (Phase 1 completion)

**Quality Gates:**
- Frontend builds successfully
- Development server connects to Go backend
- Redux DevTools integration functional

#### **Task 2.1.2: Design System and Components**
**Agent**: `frontend-developer` + `ui-ux-designer` (design)
**Priority**: HIGH
**Duration**: 6-7 days

**Agent Prompt:**
```text
Please create a comprehensive design system and base component library for radarr-go that matches the visual design of the original Radarr while providing modern React components.

Context: The original Radarr has a clean, functional interface that users are familiar with. We need to recreate this aesthetic while building on modern React patterns and accessibility standards.

Please create:

1. **Design System Foundation**:
   - Color palette matching original Radarr (dark theme primary)
   - Typography system with clear hierarchy
   - Spacing and layout system
   - Icon system using optimized SVGs
   - Animation and transition standards

2. **Base Component Library**:
   - Button variants (primary, secondary, danger, ghost)
   - Form components (Input, Select, Checkbox, Radio, TextArea)
   - Layout components (Container, Grid, Flex)
   - Feedback components (Alert, Toast, Modal, Loading)
   - Navigation components (Menu, Tabs, Breadcrumbs)

3. **Theme System**:
   - Dark theme (default, matching original Radarr)
   - Light theme option for user preference
   - High contrast theme for accessibility
   - CSS custom properties for easy theming

4. **Responsive Design Framework**:
   - Mobile-first responsive utilities
   - Breakpoint system for different screen sizes
   - Touch-friendly interface elements
   - Progressive enhancement patterns

Document all components with Storybook for development reference and include accessibility considerations throughout.
```

**Deliverables:**
- Reusable component library with Storybook
- Theme system with light/dark mode support
- Responsive design utilities and breakpoints
- Icon system with optimized SVGs

**Dependencies:** Project structure setup (Task 2.1.1)

**Quality Gates:**
- Component library documented and tested
- Theme switching functional
- Responsive design tested on multiple devices

#### **Task 2.1.3: Build Integration and Docker**
**Agent**: `devops-troubleshooter`
**Priority**: HIGH
**Duration**: 3-4 days

**Agent Prompt:**
```text
Please integrate the frontend build system with the existing radarr-go infrastructure, enabling seamless development and production deployment.

Context: Radarr-go has excellent CI/CD and deployment infrastructure for the Go backend. We need to extend this to include the new React frontend while maintaining the single-binary deployment advantage.

Please implement:

1. **Makefile Integration**:
   - Enhance existing Makefile with frontend build targets
   - `make build-frontend` - Production frontend build
   - `make dev-frontend` - Start frontend development server
   - `make build-all` - Build both frontend and backend with embedded assets

2. **Docker Multi-Stage Build**:
   - Update Dockerfile with multi-stage build for frontend
   - Node.js stage for building React application
   - Go stage for backend build with embedded frontend assets
   - Final stage with optimized Alpine image

3. **CI/CD Pipeline Enhancement**:
   - Update GitHub Actions to build frontend in CI
   - Add frontend build artifacts to release pipeline
   - Configure caching for Node.js dependencies
   - Add frontend linting and testing to quality gates

4. **Production Asset Embedding**:
   - Configure Go embed for serving frontend assets
   - Set up proper MIME types and caching headers
   - Configure fallback routing for SPA
   - Optimize asset compression and delivery

The result should be a single binary that includes both frontend and backend, maintaining radarr-go's deployment simplicity.
```

**Deliverables:**
- Enhanced Makefile with frontend targets
- Multi-stage Dockerfile with frontend assets
- CI/CD pipeline with frontend build steps
- Production build optimization configuration

**Dependencies:** Frontend project structure (Task 2.1.1)

**Quality Gates:**
- `make build-all` includes frontend assets
- Docker image contains optimized frontend build
- CI/CD successfully builds and deploys frontend

### **Sprint 2.2: Authentication and Core Layout** (Weeks 3-4)

#### **Task 2.2.1: Authentication Implementation**
**Agent**: `frontend-developer` + `security-auditor` (security review)
**Priority**: CRITICAL
**Duration**: 5-6 days

**Agent Prompt:**
```text
Please implement a secure authentication system for the radarr-go frontend that integrates with the existing Go backend API authentication.

Context: Radarr-go uses API key authentication like the original Radarr. The frontend needs to handle this securely while providing a smooth user experience.

Please implement:

1. **Authentication Flow**:
   - Login form with API key input
   - Secure storage of API key (encrypted localStorage)
   - Automatic authentication validation on app start
   - Session management with token refresh
   - Logout functionality with cleanup

2. **Protected Routes System**:
   - Route protection based on authentication status
   - Redirect unauthenticated users to login
   - Preserve intended route after login
   - Handle authentication errors gracefully

3. **API Integration**:
   - Automatic API key inclusion in all requests
   - Handle 401/403 responses appropriately
   - Retry logic for temporary authentication failures
   - Integration with RTK Query authentication

4. **Security Considerations**:
   - Secure API key storage and handling
   - XSS prevention in authentication flow
   - CSRF protection where applicable
   - Proper cleanup on logout/session expiry

Include error handling for common scenarios like invalid API keys, expired sessions, and network failures.
```

**Deliverables:**
- Authentication components with error handling
- Protected route system
- Session persistence and renewal
- Security-compliant authentication flow

**Dependencies:** Component library (Task 2.1.2)

**Quality Gates:**
- Authentication flow tested with Go backend
- Security review passed
- Session management handles edge cases

#### **Task 2.2.2: Application Shell and Navigation**
**Agent**: `frontend-developer`
**Priority**: HIGH
**Duration**: 6-7 days

**Agent Prompt:**
```text
Please create the main application shell and navigation system that will serve as the foundation for all radarr-go functionality.

Context: The original Radarr has a sidebar navigation with main content area. We need to recreate this layout while supporting all the features implemented in radarr-go (movies, calendar, queue, settings, etc.).

Please implement:

1. **Application Shell Layout**:
   - Header with search bar and user controls
   - Collapsible sidebar navigation
   - Main content area with proper spacing
   - Footer with system status information
   - Responsive layout for mobile devices

2. **Navigation System**:
   - Sidebar navigation with sections for:
     * Movies (library, add movie, collections)
     * Calendar (upcoming releases)
     * Activity (queue, history)
     * Wanted (missing, cutoff unmet)
     * Settings (all configuration sections)
     * System (tasks, health, logs)
   - Active navigation state indicators
   - Navigation counts/badges for queue, wanted movies
   - Quick navigation search

3. **Layout Responsiveness**:
   - Mobile hamburger menu for navigation
   - Tablet-optimized layout
   - Desktop sidebar with expansion/collapse
   - Touch-friendly interface elements

4. **React Router Integration**:
   - Route configuration for all main sections
   - Nested routing for complex sections
   - Route-based navigation state
   - Browser history management

Create a navigation structure that accommodates all existing radarr-go functionality while remaining intuitive and familiar to Radarr users.
```

**Deliverables:**
- Responsive application layout
- Navigation system with state indicators
- Global header with search functionality
- React Router configuration for all routes

**Dependencies:** Authentication implementation (Task 2.2.1)

**Quality Gates:**
- Layout responsive across all screen sizes
- Navigation reflects current application state
- Routing handles all planned sections

#### **Task 2.2.3: Error Handling and Loading States**
**Agent**: `frontend-developer`
**Priority**: MEDIUM
**Duration**: 4-5 days

**Agent Prompt:**
```text
Please implement comprehensive error handling and loading state management that provides excellent user experience across all radarr-go functionality.

Context: Radarr-go has extensive API functionality (150+ endpoints) and real-time features that require robust error handling and user feedback systems.

Please implement:

1. **Global Error Handling**:
   - React Error Boundaries for component errors
   - Global error handler for uncaught exceptions
   - API error handling with user-friendly messages
   - Network error detection and retry mechanisms

2. **Loading State Management**:
   - Global loading indicators for page navigation
   - Component-level loading states for async operations
   - Skeleton screens for content loading
   - Progress indicators for long-running operations

3. **User Feedback System**:
   - Toast notifications for success/error messages
   - Modal dialogs for critical errors or confirmations
   - Inline form validation with helpful error messages
   - Status indicators for system health and connectivity

4. **Offline Detection**:
   - Network connectivity monitoring
   - Offline mode indication
   - Graceful degradation when backend unavailable
   - Auto-recovery when connectivity restored

5. **Performance Considerations**:
   - Lazy loading for non-critical components
   - Image loading optimization
   - Virtual scrolling for large lists
   - Debounced search and filtering

Focus on creating a robust user experience that handles the complexity of radarr-go's extensive functionality gracefully.
```

**Deliverables:**
- Comprehensive error handling system
- User-friendly error messages and recovery
- Loading states for all async operations
- Offline mode with appropriate messaging

**Dependencies:** Application shell (Task 2.2.2)

**Quality Gates:**
- Error boundaries catch and display all error types
- Loading states provide clear user feedback
- Offline detection functional and tested

### **Sprint 2.3: API Integration Layer** (Weeks 5-6)

#### **Task 2.3.1: RTK Query Configuration**
**Agent**: `frontend-developer` + `backend-architect` (API review)
**Priority**: CRITICAL
**Duration**: 7-8 days

**Agent Prompt:**
```text
Please create a comprehensive API integration layer using RTK Query that connects the frontend to all 150+ radarr-go backend endpoints.

Context: Radarr-go has extensive API functionality with full Radarr v3 compatibility plus additional features. The frontend needs type-safe, efficient API integration with caching, optimistic updates, and real-time synchronization.

Please implement:

1. **Complete RTK Query API Slice**:
   - Examine all API handlers in `internal/api/` directory
   - Create RTK Query endpoints for all 150+ API routes
   - Include proper TypeScript typing for all requests/responses
   - Configure automatic tag-based cache invalidation

2. **API Endpoint Categories**:
   - Movies: CRUD, search, metadata, files
   - Queue: monitoring, management, statistics
   - Quality: profiles, definitions, custom formats
   - Download: clients, history, testing
   - Import: lists, sync, exclusions
   - Notifications: providers, testing, history
   - Calendar: events, feeds, configuration
   - Wanted: missing, cutoff unmet, search
   - Collections: management, TMDB sync
   - System: status, tasks, health, configuration

3. **Advanced Features**:
   - Optimistic updates for user actions
   - Automatic background refetching
   - Query deduplication and caching
   - Error retry logic with exponential backoff
   - Request cancellation for component unmount

4. **Type Safety**:
   - Generate TypeScript types from API responses
   - Strict typing for all API parameters
   - Runtime validation for critical data
   - Type-safe query hooks for components

Focus on creating an API layer that efficiently handles radarr-go's comprehensive functionality while maintaining excellent performance.
```

**Deliverables:**
- Complete RTK Query API slice
- Tag-based cache invalidation system
- Optimistic update patterns
- TypeScript types for all API responses

**Dependencies:** OpenAPI specification (Task 1.2.1)

**Quality Gates:**
- All API endpoints accessible through RTK Query
- Cache invalidation working correctly
- Type safety maintained throughout

#### **Task 2.3.2: Real-time Updates Implementation**
**Agent**: `backend-architect` (WebSocket) + `frontend-developer` (client)
**Priority**: HIGH
**Duration**: 6-7 days

**Agent Prompt:**
```text
Please implement real-time updates for radarr-go using WebSocket integration to provide live updates for queue progress, system status, and other dynamic content.

Context: The original Radarr uses SignalR for real-time updates. We need to implement similar functionality using WebSocket in Go backend and React frontend for live updates of queue progress, health status, and system activities.

Backend Tasks (backend-architect):
- Implement WebSocket server in Go backend
- Create real-time event broadcasting for:
  * Queue progress updates
  * Download status changes
  * Health check results
  * System task progress
  * Import activity
  * Calendar event updates
- Handle WebSocket connection management and authentication
- Implement proper error handling and connection recovery

Frontend Tasks (frontend-developer):
- Create WebSocket client with automatic reconnection
- Integrate WebSocket events with Redux store
- Create middleware for real-time state updates
- Handle connection state management
- Implement proper cleanup and memory management

Please implement:

1. **WebSocket Server (Go)**:
   - WebSocket endpoint at `/api/v3/signalr/messages`
   - Authentication using existing API key system
   - Event broadcasting for queue, health, activity updates
   - Connection management with cleanup

2. **WebSocket Client (React)**:
   - Automatic connection and reconnection logic
   - Redux middleware for real-time updates
   - Event subscription management
   - Connection status indicators

3. **Real-time Events**:
   - Queue item progress and status changes
   - Download completion notifications
   - Health check status updates
   - Task execution progress
   - Import activity updates

Focus on creating a reliable real-time system that enhances user experience without impacting performance.
```

**Deliverables:**
- WebSocket server implementation in Go
- Frontend WebSocket middleware
- Real-time event subscription system
- Connection resilience and recovery

**Dependencies:** RTK Query configuration (Task 2.3.1)

**Quality Gates:**
- Real-time updates functional across all supported events
- Connection recovery tested under various failure scenarios
- Performance acceptable with high-frequency updates

#### **Task 2.3.3: Data Management and Caching**
**Agent**: `frontend-developer`
**Priority**: MEDIUM
**Duration**: 4-5 days

**Agent Prompt:**
```text
Please implement efficient data management and caching strategies for the radarr-go frontend to handle large movie libraries and complex state management.

Context: Users may have movie libraries with 10,000+ movies, extensive download queues, and complex configuration. The frontend needs efficient data management to maintain performance and user experience.

Please implement:

1. **Normalized State Structure**:
   - Normalize entities (movies, queue items, activities) for efficient lookups
   - Create selectors for complex data relationships
   - Implement entity adapters for CRUD operations
   - Optimize state shape for performance

2. **Client-Side Caching Strategy**:
   - Implement intelligent cache TTL for different data types
   - Cache frequently accessed data (movies, configuration)
   - Invalidate cache based on user actions and real-time updates
   - Persist critical data across browser sessions

3. **Data Persistence**:
   - User preferences and settings in localStorage
   - View state persistence (filters, sort orders, column preferences)
   - Draft state for forms and configuration
   - Offline capability for core functionality

4. **Memory Management**:
   - Efficient handling of large movie libraries
   - Virtual scrolling data management
   - Image caching and lazy loading
   - Cleanup of unused data and subscriptions

5. **Data Transformation**:
   - Utilities for common data transformations
   - Search and filtering optimizations
   - Sorting and grouping utilities
   - Data validation and sanitization

Focus on creating a data layer that scales well with large libraries while maintaining excellent user experience.
```

**Deliverables:**
- Normalized state structure
- Client-side caching with TTL
- User preferences persistence
- Data transformation utilities

**Dependencies:** Real-time updates (Task 2.3.2)

**Quality Gates:**
- Data consistency maintained across real-time updates
- Client-side caching reduces API calls
- User preferences persist across sessions

### **Sprint 2.4: Core Movie Management** (Weeks 7-8)

#### **Task 2.4.1: Movie Library Implementation**
**Agent**: `frontend-developer`
**Priority**: CRITICAL
**Duration**: 8-9 days

**Agent Prompt:**
```text
Please implement the core movie library interface that serves as the primary user interaction point for radarr-go.

Context: This is the main feature that users will interact with daily. The movie library needs to efficiently display and manage large collections (10k+ movies) while providing the familiar Radarr experience.

Please implement:

1. **Movie Grid View**:
   - Poster grid with lazy-loaded movie posters
   - Hover states showing quick movie information
   - Status indicators (monitored, downloaded, wanted)
   - Quality badges and file information
   - Context menu for quick actions

2. **Movie List View**:
   - Tabular view with sortable columns
   - Movie title, year, quality, size, date added
   - Inline editing for monitored status
   - Bulk selection capabilities
   - Custom column configuration

3. **Advanced Filtering and Sorting**:
   - Filter by status (missing, downloaded, wanted, etc.)
   - Filter by quality profile, genre, year range
   - Filter by file presence and quality
   - Text search across title, overview, cast
   - Sort by multiple criteria with save preferences

4. **Virtual Scrolling**:
   - Efficient rendering for 10k+ movie libraries
   - Smooth scrolling performance
   - Dynamic item height support
   - Memory-efficient implementation

5. **Bulk Operations**:
   - Multi-select with shift+click and ctrl+click
   - Bulk edit (quality profile, monitored status)
   - Bulk delete with confirmation
   - Bulk search and download operations

Focus on creating a responsive, performant interface that scales well with large movie libraries while maintaining the familiar Radarr user experience.
```

**Deliverables:**
- Movie grid with poster images and basic info
- Movie list with detailed metadata
- Advanced filtering and sorting system
- Virtual scrolling for 10k+ movies

**Dependencies:** Data management system (Task 2.3.3)

**Quality Gates:**
- Grid and list views performant with large datasets
- Filtering and sorting responsive and intuitive
- Virtual scrolling maintains performance

#### **Task 2.4.2: Movie Detail and Management**
**Agent**: `frontend-developer`
**Priority**: HIGH
**Duration**: 6-7 days

**Agent Prompt:**
```text
Please implement comprehensive movie detail pages and management functionality that provides access to all radarr-go movie-related features.

Context: Each movie in radarr-go has extensive metadata, file information, download history, and management options. The detail page needs to present this information clearly while providing access to all available actions.

Please implement:

1. **Movie Detail View**:
   - Movie poster and backdrop images
   - Complete metadata (title, year, overview, cast, crew, ratings)
   - File information (quality, size, media info, languages)
   - Download history and activity timeline
   - Collection information and related movies

2. **Movie Actions Interface**:
   - Edit movie information (quality profile, monitored status)
   - Manual search with release selection
   - File management (rename, delete, quality upgrade)
   - Metadata refresh and re-identification
   - Movie deletion with file handling options

3. **File Management**:
   - Display all movie files with detailed information
   - File quality comparison and upgrade options
   - Manual import of additional files
   - File organization preview and execution
   - Media info display (codecs, resolution, etc.)

4. **Activity and History**:
   - Download history with detailed information
   - Search attempts and results
   - Import activity and decisions
   - Error tracking and resolution

5. **Integration Features**:
   - Links to external services (TMDB, IMDb)
   - Collection membership and navigation
   - Related movies and recommendations
   - Trailer and extra content links

Create a comprehensive movie management interface that provides access to all radarr-go functionality while maintaining clarity and ease of use.
```

**Deliverables:**
- Comprehensive movie detail view
- Movie editing interface
- File management system
- Action confirmation dialogs

**Dependencies:** Movie library implementation (Task 2.4.1)

**Quality Gates:**
- Movie detail page displays all relevant information
- Edit functionality maintains data integrity
- File management operations work correctly

#### **Task 2.4.3: Movie Search and Import**
**Agent**: `frontend-developer` + `backend-architect` (TMDB integration)
**Priority**: HIGH
**Duration**: 5-6 days

**Agent Prompt:**
```text
Please implement the movie search and import functionality that allows users to discover and add new movies to their radarr-go library.

Context: Radarr-go integrates with TMDB for movie discovery and supports various import methods. Users need an intuitive interface to search, preview, and add movies while avoiding duplicates and ensuring proper configuration.

Please implement:

1. **Movie Search Interface**:
   - TMDB search integration with real-time results
   - Search result cards with poster, title, year, overview
   - Advanced search filters (year, genre, language)
   - Search result pagination and infinite scroll
   - Quick add buttons with duplicate detection

2. **Add Movie Workflow**:
   - Movie selection with complete metadata preview
   - Quality profile selection with explanation
   - Root folder selection with available space
   - Monitored status configuration
   - Duplicate detection with merge options

3. **Bulk Import Operations**:
   - Bulk add from search results
   - Import from lists with preview
   - CSV/file import with validation
   - Duplicate handling strategies
   - Progress tracking for large imports

4. **Import Validation**:
   - Duplicate movie detection across different criteria
   - Quality profile validation
   - Root folder space validation
   - Movie availability checking
   - Import conflict resolution

5. **Advanced Features**:
   - Movie recommendations based on library
   - Trending and popular movie discovery
   - Collection-based adding
   - Import list integration
   - Quick search from navigation bar

Focus on creating an efficient movie discovery and addition process that prevents common user errors while supporting power user workflows.
```

**Deliverables:**
- TMDB search integration
- Add movie workflow with validation
- Bulk operation interface
- Import progress tracking

**Dependencies:** Movie detail implementation (Task 2.4.2)

**Quality Gates:**
- TMDB search returns accurate results
- Add movie workflow prevents duplicates
- Bulk operations provide clear feedback

---

## ⚡ **Phase 3: Core Feature Implementation**
*[Continue with similar detailed agent prompts for remaining phases...]*

---

## 📊 **Agent Coordination and Execution Guide**

### **How to Use These Agent Prompts**

Each task includes a complete prompt that can be used directly with the specialized Claude Code agents:

1. **Copy the Agent Prompt** from the task description
2. **Invoke the specified agent** (e.g., `database-admin`, `frontend-developer`)
3. **Provide the complete context** including dependencies and current project state
4. **Review deliverables** against the specified quality gates
5. **Coordinate with dependent tasks** before proceeding

### **Parallel Execution Strategy**

The prompts are designed to enable parallel execution where dependencies allow:

- **Phase 0**: Tasks 0.1.1 and 0.1.2 can run in parallel
- **Phase 1**: Documentation tasks can run parallel to code development
- **Phase 2**: Frontend architecture can be developed while backend enhancements are made
- **Cross-phase**: Security reviews can happen continuously throughout development

### **Quality Assurance Integration**

Each agent prompt includes:
- **Clear success criteria** that can be tested
- **Integration points** with other tasks
- **Performance requirements** where applicable
- **Security considerations** for sensitive components

---

## 🎯 **Conclusion**

This implementation plan provides actionable, complete prompts for specialized Claude Code agents to execute the radarr-go project roadmap. Each prompt includes sufficient context and specific deliverables to enable successful task completion while maintaining project coherence and quality standards.

**Ready for Immediate Execution:**
- All agent prompts are complete and actionable
- Dependencies are clearly mapped
- Quality gates provide measurable success criteria
- Parallel execution opportunities maximize delivery speed

With these detailed agent prompts, the radarr-go project can achieve full feature parity with the original Radarr while maintaining the performance and operational advantages of the Go implementation.

---

**Document Version**: 1.0

**Last Updated**: December 2024

**Next Review**: Phase 0 Sprint Planning

**Maintained By**: Radarr-Go Implementation Team
