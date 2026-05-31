# Change History

This document records all significant changes to the PostgreSQL Client project, organized by date in reverse chronological order.

---

## Update Rules

### Documentation Guidelines
- Each daily entry must not exceed 200 words
- Use concise English descriptions focusing on key changes
- List modified files when relevant
- Highlight new features, bug fixes, and improvements separately
- Maintain reverse chronological order (newest first)
- Use clear section headers with dates
- Avoid redundant details; focus on impact and functionality
- Group related changes under unified headings
- Preserve technical accuracy while ensuring readability

---

## 2026-05-31 - Modular Architecture Refactor

### New Features
- YAML configuration file support with environment variable fallback
- Cross-platform build scripts (bash for Linux/macOS, PowerShell for Windows)

### Improvements
- Reorganized codebase into modular packages:
  - `config/` - Configuration management and loading
  - `connector/` - Database connection handling  
  - `database/` - Core database operations
  - `utils/` - Utility functions (formatting, history)
- Refactored main.go to use new modular structure

### Files Modified
**Created:** config/config.go, connector/connector.go, database/database.go, utils/utils.go, build.sh, build.ps1, config.example.yaml  
**Updated:** main.go, README.md, go.mod

---