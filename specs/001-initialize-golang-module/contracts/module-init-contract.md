# Module Initialization Contract

## Overview
Contract definition for the Go module initialization process. This defines the expected inputs, outputs, and behavior for setting up the echoforge project.

## Contract: Initialize Go Module

### Endpoint/Function
**Name**: `InitializeModule`  
**Type**: Setup Script/Function  
**Purpose**: Initialize a new Go module with required dependencies

### Input Schema
```yaml
ModuleInitRequest:
  type: object
  required:
    - modulePath
    - goVersion
    - dependencies
  properties:
    modulePath:
      type: string
      pattern: '^github\.com/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$'
      example: "github.com/zcrossoverz/echoforge"
    goVersion:
      type: string
      pattern: '^1\.(2[5-9]|[3-9][0-9])(\.[0-9]+)?$'
      example: "1.25"
    dependencies:
      type: array
      items:
        $ref: '#/components/schemas/Dependency'
    createGitignore:
      type: boolean
      default: true
    createDirectories:
      type: boolean  
      default: true
```

### Output Schema
```yaml
ModuleInitResponse:
  type: object
  required:
    - success
    - moduleFiles
  properties:
    success:
      type: boolean
    moduleFiles:
      type: array
      items:
        $ref: '#/components/schemas/CreatedFile'
    binarySize:
      type: integer
      description: "Estimated binary size in bytes"
      maximum: 20971520  # 20MB
    warnings:
      type: array
      items:
        type: string
    errors:
      type: array
      items:
        type: string
```

### Component Schemas
```yaml
Dependency:
  type: object
  required:
    - name
    - version
  properties:
    name:
      type: string
      example: "github.com/gin-gonic/gin"
    version:
      type: string
      pattern: '^v[0-9]+\.[0-9]+(\.[0-9]+)?$'
      example: "v1.10.0"
    isDirect:
      type: boolean
      default: true

CreatedFile:
  type: object
  required:
    - path
    - size
  properties:
    path:
      type: string
      example: "go.mod"
    size:
      type: integer
    checksum:
      type: string
      pattern: '^[a-f0-9]{64}$'
```

### Success Response Example
```json
{
  "success": true,
  "moduleFiles": [
    {
      "path": "go.mod",
      "size": 450,
      "checksum": "a1b2c3d4e5f6..."
    },
    {
      "path": "go.sum", 
      "size": 2100,
      "checksum": "f6e5d4c3b2a1..."
    },
    {
      "path": ".gitignore",
      "size": 200,
      "checksum": "1a2b3c4d5e6f..."
    }
  ],
  "binarySize": 15728640,
  "warnings": [],
  "errors": []
}
```

### Error Response Example
```json
{
  "success": false,
  "moduleFiles": [],
  "binarySize": 0,
  "warnings": [
    "Go version 1.25 is newer than tested version 1.21"
  ],
  "errors": [
    "Invalid module path: must start with github.com/",
    "Dependency version conflict: gin v1.10.0 requires go >= 1.20"
  ]
}
```

## Contract: Validate Module Setup

### Endpoint/Function
**Name**: `ValidateModule`  
**Type**: Validation Script/Function  
**Purpose**: Verify that the module was initialized correctly

### Input Schema
```yaml
ValidationRequest:
  type: object  
  required:
    - moduleRoot
  properties:
    moduleRoot:
      type: string
      description: "Path to module root directory"
    checkBuildable:
      type: boolean
      default: true
    checkDependencies:
      type: boolean
      default: true
    checkStructure:
      type: boolean
      default: true
```

### Output Schema
```yaml
ValidationResponse:
  type: object
  required:
    - valid
    - checks
  properties:
    valid:
      type: boolean
    checks:
      type: array
      items:
        $ref: '#/components/schemas/ValidationCheck'

ValidationCheck:
  type: object
  required:
    - name
    - passed
  properties:
    name:
      type: string
      enum: ["module_file", "dependencies", "buildable", "structure", "binary_size"]
    passed:
      type: boolean
    message:
      type: string
    details:
      type: object
```

### Success Response Example
```json
{
  "valid": true,
  "checks": [
    {
      "name": "module_file",
      "passed": true,
      "message": "go.mod is valid",
      "details": {
        "modulePath": "github.com/zcrossoverz/echoforge",
        "goVersion": "1.25"
      }
    },
    {
      "name": "dependencies", 
      "passed": true,
      "message": "All dependencies resolved",
      "details": {
        "directDeps": 10,
        "indirectDeps": 45
      }
    },
    {
      "name": "buildable",
      "passed": true, 
      "message": "Module builds successfully",
      "details": {
        "buildTime": "2.3s"
      }
    },
    {
      "name": "structure",
      "passed": true,
      "message": "Directory structure is correct",
      "details": {
        "requiredDirs": ["internal", "cmd", "pkg", "configs", "tests"],
        "foundDirs": ["internal", "cmd", "pkg", "configs", "tests", "docs"]
      }
    },
    {
      "name": "binary_size",
      "passed": true,
      "message": "Binary size within limits",
      "details": {
        "sizeBytes": 15728640,
        "limitBytes": 20971520
      }
    }
  ]
}
```

## Pre-conditions
- Target directory exists and is writable
- Go toolchain is installed (version 1.25+)
- Network access for dependency download
- Git is installed for module management

## Post-conditions
- Valid go.mod file created with correct module path
- All specified dependencies added with pinned versions
- go.sum file generated with dependency checksums
- Required directory structure created
- .gitignore file created with appropriate exclusions
- Module builds successfully without errors
- Binary size is under 20MB limit

## Error Conditions
- Invalid module path format
- Go version compatibility issues
- Dependency version conflicts
- Network connectivity problems
- File system permission errors
- Binary size exceeds limit

## Performance Requirements
- Module initialization completes within 30 seconds
- Dependency resolution completes within 60 seconds
- Binary build completes within 10 seconds
- Total setup time under 2 minutes

## Security Considerations
- All dependencies must pass security vulnerability scanning
- Module path must be owned by authorized entity
- No dependencies from untrusted sources
- Checksums must be verified for integrity