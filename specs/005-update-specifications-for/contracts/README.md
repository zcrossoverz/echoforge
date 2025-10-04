# Domain Contracts

**Date**: October 4, 2025 | **Feature**: Clone-and-Extend User Domain

This directory contains the interface contracts for the updated user domain.

## Updated Interfaces

### 1. User Entity (`user_entity.go`)
- Simplified User struct without `SiteID`
- Updated validation rules
- Maintained business logic integrity

### 2. Repository Interfaces (`repository_interfaces.go`)
- Removed site-scoped operations
- Simplified method signatures
- Maintained error handling patterns

### 3. Use Case Interfaces (`usecase_interfaces.go`)
- Updated input DTOs without `SiteID`
- Simplified business logic contracts
- Maintained authentication result format

### 4. Authentication Contracts (`auth_contracts.go`)
- Simplified JWT claims structure
- Updated token generation interface
- Maintained security standards

## Design Principles
- **Single Responsibility**: Each interface has clear, focused purpose
- **Dependency Inversion**: Abstractions don't depend on implementations  
- **Interface Segregation**: Clients depend only on methods they use
- **Consistent Error Handling**: All methods return standardized errors