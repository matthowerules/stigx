# STIG Viewer X
## 1. Goals & Guiding Principles  
### 1.1. Purpose
* Provide DoD and commercial cybersecurity engineers with a free, audit-friendly, fully transparent STIG/SRG viewer–editor that supersedes the closed-source DISA tools.
* Support legacy CKL (XML) while fully embracing CKLb (JSON) and SCAP/XCCDF 1.2 content.
* Run identically on Windows, macOS, Linux (desktop first, optional CLI/ library second).
* Remain 100 % open-source, reproducibly built, FIPS-compliant crypto only, and “air-gap friendly” (single binary, no auto-update if disabled).
### 1.2. Non-Goals (for v1)
* Full eMASS API automation (often gated by policy).
* Mobile apps.
* Full SCAP scanner engine (that is a separate discipline).

## 2. Feature Backlog (ranked)

#### Must-Have (M)
* M1 Import DISA-supplied STIG Bundles (.zip) and cache metadata, then list all loaded STIGs where users can select multiples (to generate a CKL).
* M2 Open, render, search, and filter CKL (XML) & CKLb (JSON) checklists. These must save to and be stored on disk.
* M3 Edit status, severity, finding details, comments, POA&M dates, etc.
* M4 Lossless conversion CKL ⇄ CKLb.
* M5 Validate files against official XSD/JSON-Schema; flag schema drift.
* M6 Export CKL (for eMASS) and CKLb (for STIG V3 interoperability).
* M7 Offline documentation panel with direct links to STIG rules.
* M8 Signed portable binaries (.exe, .dmg, .AppImage) with detached SHA-256.
#### Should-Have (S)
* S1 Import / export XCCDF 1.2 (single benchmark or multi-profile) [2].
* S2 Rule-diff view (compare checklist to new STIG version; highlight deltas).
* S3 Built-in “quick-fill” templates for common mitigations.
* S4 CLI wrapper: stig-viewer-x.exe convert foo.ckl bar.cklb.
* S5 JSON-RPC or REST plug-in API (for CI pipelines).
* S6 Accessibility (WCAG 2.1 AA) + high-contrast/dark mode.
#### Could-Have (C)
* C1 Multi-user collaboration via SQLite->PostgreSQL upgrade path.
* C2 Git integration: save checklist as repo, auto-commit on change.
* C3 Binary signing / integrity attestations for exported artifacts.
* C4 Tag rules, create custom checklists (e.g., by system role).
* C5 Mermaid-JS diagram export of compliance posture.
#### Won’t-Have (W) for now
* W1 Native mobile apps.
* W2 Embedded vulnerability scanner engine.

## 3. Technical-Stack Options
#### Evaluation criteria: 
1. cross-platform UI, 
2. license compatibility, 
3. community size, 
4. binary footprint, 
5. learning curve.

## 4. Proposed Architecture

### Core Language: Go (Golang)

### UI Framework:
Wails – A Go-based framework that integrates with modern web technologies (HTML/CSS/JS) for the UI.

### File Parsing:
XML: Use Go's encoding/xml package for CKL parsing.
JSON: Use Go's encoding/json package for CKLb parsing.
XCCDF: Use encoding/xml with custom structs for SCAP/XCCDF.

### Database: SQLite (via go-sqlite3).
* CLI: Native Go CLI tools using flag or cobra.
* Testing: Go's built-in testing package for unit tests and integration tests.
* Build Tools: Use Go's go build for reproducible builds and cross-compilation.

### Architecture (Go-Based)
1. File I/O Layer
    - XML Parsing: Use encoding/xml to parse CKL files into Go structs.
    - JSON Parsing: Use encoding/json to parse CKLb files into Go structs.
    - Validation: Implement schema validation using libraries like gojsonschema for JSON and custom XSD validation for XML.
2. Domain Logic
    - Models: Define Go structs for STIG rules, findings, assets, and history.
    - Converters: Implement functions to convert between CKL ⇄ CKLb ⇄ XCCDF formats.
    - Validation: Add functions to validate checklist files against schemas.
3. Persistence
    - SQLite: Use go-sqlite3 for local storage of checklist data.
    - Change Tracking: Implement basic change tracking using SQLite triggers.
4. UI Layer
    - Wails: Use Wails to create a modern UI with web technologies (React or Svelte) while keeping the backend in Go.
5. CLI Wrapper
    - Build a CLI tool using cobra or flag for tasks like file conversion, validation, and batch processing.
6. Security Hardening
    - Static Analysis: Use tools like golangci-lint for linting and security checks.
    - Supply Chain: Use go mod verify to ensure dependency integrity.
    - Binary Signing: Use cosign or similar tools to sign binaries.

## 5. Phased Roadmap (18-month view)

#### Phase 0 (Month 0-1): Skeleton
* Set up the GitHub repository with Go modules (go mod init).
* Create basic CLI scaffolding (cobra or flag).
* Define initial Go structs for CKL and CKLb parsing.
#### Phase 1 (Month 1-4): “MVP-parse”
* Implement CKL/CKLb read-only parsing and schema validation.
* Build a basic UI (Wails) to open and view files.
* Release v0.1.
#### Phase 2 (Month 4-7): “Editor”
* Add full editing capabilities (status, severity, comments, etc.).
* Implement CKL ⇄ CKLb conversion.
* Release v0.5 (beta).
#### Phase 3 (Month 7-10): “XCCDF & Diff”
* Add XCCDF import/export functionality.
* Implement rule-diff view to compare checklists.
* Release v1.0.
#### Phase 4 (Month 10-14): “Collaboration & API”
* Add SQLite-based multi-user collaboration.
* Implement REST/JSON-RPC API for automation.
* Release v1.3.
#### Phase 5 (Month 14-18): “Ecosystem”
* Add plug-in support for custom rule processing.
* Package binaries for Homebrew, Chocolatey, Snap, etc.
* Release v2.0+.