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

## 2026-06-02 - Row Insertion, Navigation Refactor, and Bug Fixes

### New Features
- Row creation in table content view via `[Add Row]` menu option
  - Copy from existing row (pre-fills column values) or manual input
  - Auto-refreshes page after insertion
- Build scripts accept optional target argument for single-platform compilation
- File logging to `postgresql-client.log` in working directory via `Logger.SetLogFile()`

### Improvements
- Collapsed page navigation (First/Prev/Next/Last/Goto) into a unified sub-menu accessed via `[Navigation]`
- Renamed packages: `connector/` to `database/`, `ui/` to `cli/`, `utils/` to `commons/`

### Bug Fixes
- Fixed `showMainMenu` prematurely closing database connection before `selectDatabaseInteractive` could list databases
- Removed unreachable code after `showTableContent` loop

### Files Modified
**Updated:** main.go, internal/cli/interface.go, internal/commons/commons.go, build.sh, build.ps1, CHANGE_HISTORY.md

---

## 2026-05-31 - Interactive UI and Commons Library Addition

### New Features
- YAML configuration file support with environment variable fallback
- Cross-platform build scripts (bash for Linux/macOS, PowerShell for Windows)
- Interactive terminal UI using survey/v2 library for database/table selection
- Database selector with interactive arrow-key navigation
- Table selector with action menu (view structure, content, edit)
- Row editor with column selection and data modification support

### Improvements
- Reorganized codebase into modular packages:
  - `config/` - Configuration management and loading
  - `connector/` - Database connection handling  
  - `database/` - Core database operations
  - `ui/` - Interactive UI components (selectors, menus)
  - `utils/` - Utility functions (formatting, history)
  - `commons/` - Shared utilities, logging, and error types
- Refactored main.go to use new modular structure

### Files Modified
**Created:** commons/commons.go, config/config.go, connector/connector.go, database/database.go, ui/ui.go, utils/utils.go, build.sh, build.ps1, config.example.yaml  
**Updated:** go.mod (added survey/v2 dependency), main.go, README.md

---