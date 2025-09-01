# STIG Viewer X

A free, open-source, audit-friendly STIG/SRG viewer and editor for DoD and commercial cybersecurity engineers. Built to supplement the closed-source DISA tools with full transparency and full cross-platform support (aka we use Macs and are tired of jumping through hoops between STIG Viewers 2 and 3 on different platforms).

## Current Status

**Phase 2 - Editor**: Functional desktop application with comprehensive STIG/CKL editing capabilities. Core architecture complete with production(ish)-ready Go backend and modern React frontend using Wails framework.

### Completed Features
- **Full GUI Application**: Modern desktop app with React/TypeScript frontend
- **STIG Bundle Import**: Import DISA STIG bundles (.zip) with progress tracking
- **STIG Management**: View, select, and manage imported STIG libraries
- **Checklist Creation**: Create new checklists from selected STIGs (defaults to CKLb format)
- **File Operations**: Open, load, and save CKL (XML) and CKLb (JSON) files
- **Advanced Editor**: Full editing capabilities for rule status, finding details, comments, and POA&M dates
- **File Conversion**: Lossless bidirectional conversion between CKL ↔ CKLb formats
- **Validation**: Schema validation against official XSD/JSON-Schema with drift detection
- **Multi-Tab Interface**: View multiple checklists simultaneously with tabbed interface
- **Rule Filtering**: Filter rules by status, severity, and search functionality
- **Summary Dashboard**: Real-time compliance statistics and progress tracking
- **Cross-Platform**: Native binaries for Windows, macOS, and Linux
- **SQLite Database**: Local data persistence with full metadata caching

### In Progress
- **CLI Wrapper**: Command-line interface for batch operations
- **XCCDF Support**: Import XCCDF 1.2 benchmarks on top of checklists
- **Rule Diff**: Compare checklists against new STIG versions

## Features

### Core Functionality (Implemented)
- **Import STIG Bundles**: Load DISA-supplied STIG bundles (.zip) with progress tracking and metadata caching
- **STIG Library Management**: Browse, search, and select from imported STIG collections
- **Checklist Creation**: Generate new checklists from selected STIGs with customizable target data
- **File Support**: Open, render, search, and filter CKL (XML) and CKLb (JSON) checklists
- **Advanced Editor**: Modify rule status, severity, finding details, comments, and POA&M dates
- **Conversion**: Lossless bidirectional conversion between CKL ↔ CKLb formats with size optimization
- **Validation**: Schema validation against official XSD/JSON-Schema with comprehensive error reporting
- **Export/Save**: Generate CKL (for eMASS) and CKLb (for STIG V3 interoperability) files
- **Multi-Tab Interface**: View and edit multiple checklists simultaneously
- **Rule Filtering**: Advanced filtering by status, severity, STIG source, and search terms
- **Summary Dashboard**: Real-time compliance statistics, completion tracking, and findings overview
- **Cross-Platform Distribution**: Native signed binaries for Windows, macOS, and Linux

### Planned Features
- **CLI Wrapper**: Command-line interface for batch file operations and automation
- **XCCDF Support**: Import/export XCCDF 1.2 benchmarks and profiles
- **Rule Diff**: Compare checklists against new STIG versions with change highlighting
- **Accessibility**: WCAG 2.1 AA compliance with high-contrast/dark mode themes
- **Documentation Panel**: Offline STIG rule documentation with direct links

## Tech Stack

- **Backend**: Go 1.24.5 with Wails v2 framework
- **Frontend**: React 18 with TypeScript, Vite build system
- **Database**: SQLite with modernc.org/sqlite driver
- **UI Framework**: Wails v2 (Go + web technologies)
- **Styling**: CSS with modern responsive design
- **State Management**: React hooks with context API
- **File Processing**: Custom parsers for CKL (XML) and CKLb (JSON)
- **Security**: Input validation, path sanitization, and secure file operations
- **Cross-Platform**: Native desktop apps for Windows, macOS, Linux
- **CLI**: Cobra command-line framework (in development)

## Installation

### Prerequisites
- Go 1.24.5 or later
- Node.js 18+ and npm
- Git

### Build from Source

1. Clone the repository:
    ```bash
    git clone https://github.com/matthowerules/stigx.git
    cd stigx
    ```

2. Install Go dependencies:
    ```bash
    go mod download
    ```

3. Install frontend dependencies:
    ```bash
    cd frontend
    npm install
    cd ..
    ```

4. Build the application:
    ```bash
    wails build
    ```

The build process will create a native executable for your platform in the `build/` directory.

### System Requirements
- **Linux**: `libgtk-3-dev` and `libwebkit2gtk-4.1-dev` packages
- **Windows**: No additional dependencies required
- **macOS**: Xcode command-line tools

## Usage

### Desktop Application
Run the built application to launch the full-featured GUI for STIG management and checklist editing.

#### Key Workflows:
1. **Import STIGs**: Click "Import STIGs" to load DISA STIG bundles (.zip files)
2. **View STIG Library**: Use "View STIGs" to browse and select from imported STIGs
3. **Create Checklists**: Select STIGs and click "Create Checklist from Selected" to generate new checklists
4. **Load Existing**: Use "Load Checklist" to open existing CKL or CKLb files
5. **Edit Rules**: Click on any rule to view details and modify status, findings, and comments
6. **Save Work**: Use "Save Checklist" to export in CKL or CKLb format

#### Interface Features:
- **Multi-Tab Editing**: Work with multiple checklists simultaneously
- **Rule Filtering**: Filter by status (Not Reviewed, Open, Not a Finding, etc.) and severity
- **Search**: Find specific rules by ID, title, or content
- **Summary Dashboard**: Track completion percentage and findings overview
- **File Conversion**: Convert between CKL and CKLb formats with size reporting

### CLI (In Development)
```bash
# Planned CLI commands
stigx-cli convert input.ckl output.cklb    # Convert file formats
stigx-cli validate checklist.cklb         # Validate file schema
stigx-cli batch --input-dir ./checklists  # Batch processing
stigx-cli import bundle.zip               # Import STIG bundles
```

## Project Structure

```
stigx/
├── cmd/
│   ├── cli/          # CLI application entry point (in development)
│   └── stigx/        # GUI application entry point (Wails)
├── internal/
│   ├── cli/          # CLI command implementations
│   ├── converters/   # File format conversion logic
│   ├── database/     # SQLite database operations
│   ├── models/       # Data structures for STIG/CKL/CKLb
│   ├── parsers/      # File parsing and validation
│   └── services/     # Business logic services
├── frontend/         # React/TypeScript web application
│   ├── src/          # Source code
│   ├── public/       # Static assets
│   └── wailsjs/      # Generated Wails bindings
├── docs/             # Documentation
├── pkgconfig/        # Build configuration
```

## Contributing

Please, by all means, help us out - leave some comments, submit some PRs, give us feedback on what you would like. This started as a project for what we wanted a modern STIG Viewer to do and support in our day to day, but we know it's not great.

## Roadmap

See [docs/PROJECT.md](docs/PROJECT.md) for detailed development roadmap and feature backlog.

## Security

We are not there yet, but this is what we are aiming for in the project:
- FIPS-compliant cryptography
- Static analysis
- Dependency integrity verification
- Binary signing

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## Acknowledgments

- Built to support the DoD cybersecurity community
- Inspired by the need for transparent, open-source STIG tools
- Thanks to the Wails framework, React community, and Go ecosystem
- Special thanks to DISA for providing STIG resources, CIS for their benchmarks, Skyward Federal for keeping us employed, Anthropic for Claude Code, GitHub's copilot, 

## Support

- **Email**: matt@matthowe.org
- **Issues**: [GitHub Issues](https://github.com/matthowerules/stigx/issues)
- **Discussions**: [GitHub Discussions](https://github.com/matthowerules/stigx/discussions)
- **Documentation**: See [docs/](docs/) directory for detailed guides

