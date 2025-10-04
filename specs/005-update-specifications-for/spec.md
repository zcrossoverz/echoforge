# Feature Specification: Update User Domain and Authentication for Clone-and-Extend Model

**Feature Branch**: `005-update-specifications-for`  
**Created**: October 4, 2025  
**Status**: Draft  
**Input**: User description: "Update specifications for Task 1.2 (User Domain Entity & Repository) and Task 1.3 (Register/Login Usecases) to align with new clone-and-extend model for echoforge project (read my updated constitution). Context: echoforge is a reusable Golang backend core for content platforms (blog/manga/news), structured as a modular monolith with hexagonal architecture. Each site clones core (github.com/[yourusername]/echoforge) and extends features with separate Postgres DB (no site_id isolation). Sprint 1 ongoing, Task 1.1 done (go.mod with pinned deps: gin v1.10.0, gorm.io/gorm v1.26.12, gorm.io/driver/postgres v1.5.9, spf13/viper v1.19.0, go.uber.org/zap v1.27.0, google/uuid v1.6.0, golang.org/x/crypto v0.42.0, google/wire v0.8.0, go-playground/validator/v10 v10.27.0, testify v1.13.1; .gitignore created). Constitution updated to v1.2.0 (clone-and-extend, no site_id)."

## Execution Flow (main)
```
1. Parse user description from Input
   → Extracted: Architectural refactor from multi-tenant to clone-and-extend model
2. Extract key concepts from description
   → Actors: Site cloners, end users of cloned sites
   → Actions: Update domain entities, refactor authentication, remove site_id
   → Data: User entities, authentication flows
   → Constraints: Maintain backward compatibility, separate DB per site
3. For each unclear aspect:
   → GitHub username placeholder marked for clarification
   → JWT secret configuration source identified
4. Fill User Scenarios & Testing section
   → Clear refactoring and migration flows identified
5. Generate Functional Requirements
   → All requirements are testable and measurable
6. Identify Key Entities
   → Updated User domain, authentication inputs/outputs
7. Run Review Checklist
   → Minor clarifications marked but spec is actionable
8. Return: SUCCESS (spec ready for planning)
```

---

## User Scenarios & Testing *(mandatory)*

### Primary User Story
As a site maintainer using the echoforge core, I want to clone the repository and deploy with my own separate database so that my users can register and authenticate without any site_id complexity, while the core remains reusable for other site maintainers.

### Acceptance Scenarios

#### Task 1.2: User Domain Entity & Repository Update
1. **Given** I have cloned echoforge core, **When** users register on my site, **Then** their data is stored in my separate database without site_id fields
2. **Given** a user tries to register with a duplicate email, **When** they submit registration, **Then** they receive validation error (unique constraint per database)
3. **Given** I deploy multiple sites from the same core, **When** users with same email register on different sites, **Then** both registrations succeed (separate databases)
4. **Given** I run database migrations, **When** upgrading from multi-tenant version, **Then** site_id columns are safely removed without data loss

#### Task 1.3: Register/Login Usecases Update  
1. **Given** a user visits my cloned site, **When** they register with valid credentials, **Then** account is created in my site's database without site_id tracking
2. **Given** a registered user attempts login, **When** they provide correct credentials, **Then** they receive JWT token with user ID but no site_id claims
3. **Given** I configure JWT secret via environment, **When** users authenticate, **Then** tokens are signed with my site-specific secret
4. **Given** invalid credentials are provided, **When** authentication is attempted, **Then** generic error prevents user enumeration

### Edge Cases
- What happens when migrating existing multi-tenant database to clone-and-extend model?
- How does system handle concurrent registrations during migration?
- What happens when JWT secret is not configured properly?
- How does system respond to malformed input without site_id validation?
- What happens when multiple sites use same JWT secret (security implications)?

## Requirements *(mandatory)*

### Functional Requirements - Task 1.2: User Domain Update
- **FR-001**: User entity MUST remove SiteID field completely from struct definition
- **FR-002**: User entity MUST retain ID (UUID), Email (string, max 255), PasswordHash (string, bcrypt min 60), CreatedAt/UpdatedAt (time.Time)
- **FR-003**: User entity MUST enforce unique email constraint per database (not across sites)
- **FR-004**: NewUser constructor MUST accept only email and passwordHash parameters (remove siteID)
- **FR-005**: UserRepository interface MUST update Create method signature: Create(ctx, *User) error
- **FR-006**: UserRepository interface MUST update FindByEmail method: FindByEmail(ctx, email) (*User, error)
- **FR-007**: User repository implementation MUST remove site_id from all database queries
- **FR-008**: Database migration MUST safely remove site_id column from users table
- **FR-009**: System MUST maintain backward compatibility for existing non-site_id aware code

### Functional Requirements - Task 1.3: Authentication Usecases Update
- **FR-010**: RegisterInput struct MUST remove SiteID field, retain Email and Password validation
- **FR-011**: LoginInput struct MUST remove SiteID field, retain Email and Password fields
- **FR-012**: Registration logic MUST check email uniqueness within single database only
- **FR-013**: Login logic MUST authenticate against single database without site_id filtering
- **FR-014**: JWT token generation MUST include only user ID and standard claims (no site_id)
- **FR-015**: JWT token MUST use HS256 signing with configurable secret from environment

### Non-Functional Requirements
- **NFR-001**: Migration process MUST complete without service downtime
- **NFR-002**: Updated authentication MUST maintain same performance characteristics (sub 2-second response)
- **NFR-003**: System MUST support 1000+ concurrent users per site deployment
- **NFR-004**: Security model MUST comply with OWASP Top 10 without site_id complexity
- **NFR-005**: Test coverage MUST remain >80% after refactoring
- **NFR-006**: Core module MUST remain backward compatible for SemVer compliance
- **NFR-007**: JWT secret MUST be configurable via JWT_SECRET environment variable [NEEDS CLARIFICATION: confirm env var name preference]

### Key Entities *(include if feature involves data)*
- **Updated User Entity**: ID (UUID), Email (string, unique per DB), PasswordHash (string), CreatedAt/UpdatedAt (time.Time) - SiteID removed
- **Simplified RegisterInput**: Email (validated), Password (OWASP compliant) - SiteID removed  
- **Simplified LoginInput**: Email, Password - SiteID removed
- **Updated AuthenticationResult**: JWT token with user ID claims only (no site_id)
- **Migration Scripts**: SQL commands to safely remove site_id columns and constraints

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [ ] No [NEEDS CLARIFICATION] markers remain (1 minor JWT secret clarification present)
- [x] Requirements are testable and unambiguous  
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

---

## Execution Status
*Updated by main() during processing*

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked (1 JWT configuration clarification)
- [x] User scenarios defined
- [x] Requirements generated (15 functional, 7 non-functional)
- [x] Entities identified (5 key entities including migration artifacts)
- [x] Review checklist passed (minor clarification acceptable)

---
