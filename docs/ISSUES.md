# Security Vulnerabilities

## 1. Zip Bomb/ZIP Slip Vulnerability - **HIGH SEVERITY**
**Location**: `internal/services/stig_service.go:392-471`
**Issue**: The ZIP file extraction in `ImportBundle`, `countXCCDFInNestedZip`, and `processNestedZip` functions is vulnerable to:
- **Zip Bomb attacks**: No limits on decompressed content size or nested ZIP depth
- **ZIP Slip attacks**: No validation of file paths during extraction to prevent directory traversal
- **Resource exhaustion**: Unlimited memory allocation via `io.ReadAll(reader)`

**Impact**: Attackers could:
- Consume all available memory/disk space with crafted ZIP files
- Write files outside intended directories
- Cause denial of service through resource exhaustion

**Mitigation**: 
- Implement size limits for decompressed content
- Validate extracted file paths to prevent directory traversal
- Limit nested ZIP processing depth
- Use streaming readers with size limits instead of `io.ReadAll`

## 2. XML External Entity (XXE) Injection - **HIGH SEVERITY**
**Location**: `internal/models/common.go:264-278`
**Issue**: While basic XXE protections exist in `ValidateXMLContent`, the XML parsers in `internal/parsers/ckl.go` and XCCDF processing may not be properly configured to prevent XXE attacks.

**Impact**: Attackers could:
- Read local system files
- Perform server-side request forgery (SSRF)
- Cause denial of service
- Access sensitive information

**Mitigation**:
- Ensure all XML parsers have external entity processing disabled
- Use secure XML parser configuration with `xml.NewDecoder()` and disable external entity resolution

## 3. Path Traversal in File Operations - **MEDIUM SEVERITY**
**Location**: `main.go:147-151, 265-270`
**Issue**: File path validation in `OpenFileDialog` and `SaveFileDialog` relies on basic checks but may not prevent all path traversal scenarios on different operating systems.

**Impact**: Users could potentially:
- Access files outside intended directories
- Overwrite system files if running with elevated privileges

**Mitigation**:
- Strengthen path validation with platform-specific checks
- Use `filepath.Rel()` to ensure paths stay within allowed directories
- Implement a whitelist of allowed directories

## 4. Debug Information Disclosure - **LOW SEVERITY**
**Location**: `internal/services/stig_service.go:269-287, 412-419`
**Issue**: Debug print statements (`fmt.Printf`) expose internal file structure and processing details to stdout in production.

**Impact**: Information leakage that could aid attackers in understanding system internals.

**Mitigation**:
- Remove or conditionally compile debug statements
- Use proper logging with configurable levels
- Ensure no sensitive information is logged

# Bugs

## 1. Potential SQL Injection via Error Messages - **MEDIUM SEVERITY**
**Location**: `internal/database/database.go:95-121`
**Issue**: Database statistics queries construct SQL dynamically, though currently safe, this pattern could lead to SQL injection if modified.

**Impact**: Potential for SQL injection if code is modified improperly.

**Mitigation**:
- Use parameterized queries consistently
- Add SQL injection prevention code review guidelines

## 2. Missing Input Validation in JSON Parsing - **LOW SEVERITY**  
**Location**: `internal/parsers/cklb.go:44-52`
**Issue**: JSON unmarshaling doesn't validate structure depth or field lengths, potentially allowing memory exhaustion with deeply nested JSON.

**Impact**: Denial of service through memory exhaustion with crafted JSON files.

**Mitigation**:
- Implement maximum JSON nesting depth
- Validate field lengths and structure size
- Use streaming JSON parser for large files

## 3. File Handle Leaks in Error Conditions - **LOW SEVERITY**
**Location**: Multiple locations with `os.Open()` calls
**Issue**: Some error paths may not properly close file handles due to early returns.

**Impact**: Resource leaks could lead to file handle exhaustion.

**Mitigation**:
- Audit all file operations for proper cleanup
- Use defer statements immediately after successful opens
- Consider using context with timeouts for file operations

## 4. Import Progress Bar - **UI BUG**
**Location**: Frontend progress display
**Issue**: Text rendering on Current and Progress is redundant, should eliminate the top one, rename the bottom one, then ensure that the file name wraps within the element and does truncates what is displayed to not blast the whole screen with text.

## 5. Tab Display Names - **UI BUG**
**Location**: Frontend tab interface  
**Issue**: Support opening and viewing multiple CKL files. Feature is complete, now need to ensure that tabs start the beginning of the filename as they are reading some field from the file instead.

# Features

## 1. Users should be able to delete imported STIGs from the View STIGs menu
  x - Complete

# Improvements

## 1. Improve the manner STIGs are displayed in View STIGs. Allow easier multi-selection.
  x - Complete
## 2. Speed improvements on low-spec systems
## 3. Sign binaries on mac and Windows  
## 4. Color/theme improvements for editable fields (need to validate on all platforms as they have rendered differently)

# Dependency Vulnerabilities

## 1. Outdated Frontend Dependencies - **MEDIUM SEVERITY**
**Location**: `frontend/package.json`
**Issue**: Several dependencies are outdated:
- TypeScript 4.6.4 (current stable: 5.x)
- Vite 3.0.7 (current stable: 5.x) 
- React 18.2.0 (current: 18.3.x)

**Impact**: Known security vulnerabilities in older versions.

**Mitigation**:
- Update all dependencies to latest stable versions
- Implement automated dependency scanning
- Regular security updates schedule

## 2. Go Dependencies Security Review Needed - **LOW SEVERITY**
**Location**: `go.mod`
**Issue**: Dependencies should be regularly audited for known vulnerabilities.

**Impact**: Potential security vulnerabilities in third-party code.

**Mitigation**:
- Run `go audit` or similar tools regularly
- Keep dependencies updated
- Monitor security advisories

