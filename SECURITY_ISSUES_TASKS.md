# Security Issues Fix Tasks

## Overview
This document lists all security issues found by GitHub Code Scanning and provides detailed tasks to fix each one.

## Critical Security Issues (9 issues)

### 1. NoSQL Injection Vulnerabilities (6 issues)

#### Alert #15: NoSQL Injection in log_repository.go (line 196)
**File:** `backend/internal/repository/mongodb/log_repository.go`
**Fix:** Replace direct query construction with parameterized queries using MongoDB's built-in filtering methods

#### Alert #14: NoSQL Injection in log_repository.go (line 95)
**File:** `backend/internal/repository/mongodb/log_repository.go`
**Fix:** Use proper BSON filtering instead of string concatenation

#### Alert #13: NoSQL Injection in log_repository.go (line 78)
**File:** `backend/internal/repository/mongodb/log_repository.go`
**Fix:** Sanitize input and use MongoDB's structured query format

#### Alert #12: NoSQL Injection in user_repository.go (line 64)
**File:** `backend/internal/repository/mongodb/user_repository.go`
**Fix:** Use parameterized queries with BSON filters

#### Alert #11: NoSQL Injection in environment.go (line 96)
**File:** `backend/internal/repository/mongodb/environment.go`
**Fix:** Replace dynamic query building with safe query methods

#### Alert #10: NoSQL Injection in environment.go (line 68)
**File:** `backend/internal/repository/mongodb/environment.go`
**Fix:** Use structured queries instead of string concatenation

### 2. Command Injection (1 issue)

#### Alert #9: Command Injection in SSH manager
**File:** `backend/internal/service/ssh/manager.go` (line 81)
**Fix:** Validate and sanitize user input, use command arguments array instead of shell string

### 3. Server-Side Request Forgery (SSRF) (1 issue)

#### Alert #8: Uncontrolled data in network request
**File:** `backend/internal/service/environment/service.go` (line 939)
**Fix:** Validate URLs against an allowlist, ensure proper URL parsing and validation

### 4. Insecure SSH Configuration (1 issue)

#### Alert #7: Insecure HostKeyCallback
**File:** `backend/internal/service/ssh/manager.go` (line 198)
**Fix:** Implement proper host key verification instead of InsecureIgnoreHostKey

## GitHub Actions Security Issues (6 issues)

### 5. Expression/Code Injection in Actions (2 issues)

#### Alert #6: Expression injection in deploy-notification.yml
**File:** `.github/workflows/deploy-notification.yml` (lines 16-24)
**Fix:** Use environment variables instead of direct expression interpolation

#### Alert #1: Code injection in deploy-notification.yml
**File:** `.github/workflows/deploy-notification.yml` (line 19)
**Fix:** Set untrusted input to environment variable and use shell syntax

### 6. Missing Workflow Permissions (4 issues)

#### Alert #5: Missing permissions in ci-cd.yml (deploy job)
**File:** `.github/workflows/ci-cd.yml` (lines 146-221)
**Fix:** Add explicit permissions block with minimal required permissions

#### Alert #4: Missing permissions in ci-cd.yml (build-frontend job)
**File:** `.github/workflows/ci-cd.yml` (lines 55-84)
**Fix:** Add explicit permissions block

#### Alert #3: Missing permissions in ci-cd.yml (build-backend job)
**File:** `.github/workflows/ci-cd.yml` (lines 26-54)
**Fix:** Add explicit permissions block

#### Alert #2: Missing permissions in deploy-notification.yml
**File:** `.github/workflows/deploy-notification.yml` (lines 11-24)
**Fix:** Add explicit permissions block

## Implementation Plan

1. Create a new branch for security fixes
2. Fix NoSQL injection vulnerabilities (highest priority)
3. Fix command injection vulnerability
4. Fix SSRF vulnerability
5. Fix insecure SSH configuration
6. Fix GitHub Actions security issues
7. Run tests after each fix
8. Build and verify the application
9. Commit changes with appropriate message
