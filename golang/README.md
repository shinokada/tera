# TERA Go Documentation Index

## Status: Ready to Start Coding ✅

All essential documentation is complete. You have everything needed to begin Phase 0.

---

## Core Documentation

### 1. API_SPEC.md
**Purpose:** Complete API specification for all interfaces  
**Contains:**
- Radio Browser API client interface
- Storage interface methods
- Player interface
- Gist client interface
- Token management
- Data models
- Error types
- Constants

**Use for:** Understanding contracts between components

---

### 2. GETTING_STARTED.md
**Purpose:** Step-by-step guide to start development  
**Contains:**
- Prerequisites
- Project setup commands
- Directory structure
- Phase 1 implementation (Data Models)
- Phase 2 implementation (Storage)
- Phase 3 implementation (API Client)
- Phase 4 implementation (Basic UI)
- Example code for each phase
- Makefile
- Common issues

**Use for:** Setting up project and building first version

---

### 3. TESTING.md
**Purpose:** Complete testing strategy and examples  
**Contains:**
- Test organization
- Unit test examples
- Integration test examples
- Table-driven tests
- Mocking patterns
- Coverage goals
- CI/CD setup
- Manual testing checklist
- Performance testing

**Use for:** Writing tests for each component

---

## Supporting Documentation (in spec-documents/)

### 4. core-goals.md
**Purpose:** Why convert to Go  
**Benefits:**
- Native TUI libraries
- Single binary
- Cross-compilation
- Performance
- Built-in JSON
- Type safety

---

### 5. keyboard-shortcuts-guide.md
**Purpose:** Complete keyboard navigation spec  
**Contains:**
- All keyboard shortcuts
- Context-specific keys
- Navigation patterns
- Safety model
- Platform notes

**Use for:** Implementing keyboard handling in UI

---

### 6. technical-approach.md
**Purpose:** Detailed architecture and design patterns  
**Contains:**
- Project structure
- Technology stack details
- Design patterns (MVU, Command, Message-based)
- Component interfaces
- State management
- Error handling
- Performance optimizations
- Testing approach

**Use for:** Understanding overall architecture

---

### 7. implementation-plan.md
**Purpose:** Phased development roadmap  
**Contains:**
- 10 development phases
- Timeline (41-52 days)
- Phase breakdown with tasks
- Milestones
- Risk mitigation
- Success criteria

**Use for:** Planning work and tracking progress

---

### 8. flow-charts.md + flow-charts-part2.md
**Purpose:** Visual flow diagrams for every screen  
**Contains:**
- Mermaid diagrams for all 14 screens
- State transitions
- Data flow
- Error handling patterns
- Navigation flows

**Use for:** Understanding screen behavior and transitions

---

## What Each Document Provides

| Document                    | Answers                                      |
| --------------------------- | -------------------------------------------- |
| API_SPEC.md                 | What interfaces/methods do I implement?      |
| GETTING_STARTED.md          | How do I set up and build the first version? |
| TESTING.md                  | How do I test my code?                       |
| keyboard-shortcuts-guide.md | What keyboard shortcuts do I implement?      |
| technical-approach.md       | What architecture patterns do I use?         |
| implementation-plan.md      | What order do I build features?              |
| flow-charts.md              | How do screens behave?                       |

---

## How to Start Coding

### Step 1: Project Setup (Day 1)
Follow **GETTING_STARTED.md** sections:
1. Install dependencies
2. Create directory structure
3. Initialize go.mod
4. Create basic Makefile

### Step 2: Data Models (Days 2-3)
Implement from **API_SPEC.md** and **GETTING_STARTED.md**:
1. Create `internal/api/models.go`
2. Create `internal/storage/models.go`
3. Write tests (use **TESTING.md** examples)
4. Verify with `go test ./...`

### Step 3: Storage Layer (Days 4-6)
Implement from **API_SPEC.md**:
1. Create `internal/storage/favorites.go`
2. Implement all Storage interface methods
3. Add file locking
4. Write comprehensive tests
5. Test backward compatibility with bash files

### Step 4: API Client (Days 7-9)
Implement from **API_SPEC.md**:
1. Create `internal/api/client.go`
2. Implement all search methods
3. Add retry logic
4. Write unit tests with mock server
5. Test integration with live API

### Step 5: Continue with Phases
Follow **implementation-plan.md** for remaining phases

---

## Quick Reference

### When Implementing a Screen
1. Check **flow-charts.md** for that screen
2. Note all states and transitions
3. Check **keyboard-shortcuts-guide.md** for keys
4. Implement using Bubble Tea MVU pattern
5. Write tests per **TESTING.md**

### When Adding a Feature
1. Define interface in **API_SPEC.md** style
2. Write tests first (TDD)
3. Implement feature
4. Update documentation
5. Run full test suite

### When Stuck
1. **Architecture questions:** See **technical-approach.md**
2. **Implementation order:** See **implementation-plan.md**
3. **Testing questions:** See **TESTING.md**
4. **UI behavior:** See **flow-charts.md**
5. **Setup issues:** See **GETTING_STARTED.md**

---

## Documentation Maintenance

### When Adding New Features
Update these docs:
- [ ] API_SPEC.md (if new interface)
- [ ] keyboard-shortcuts-guide.md (if new shortcuts)
- [ ] flow-charts.md (if new screen)
- [ ] TESTING.md (add test examples)

### When Changing Architecture
Update:
- [ ] technical-approach.md
- [ ] API_SPEC.md
- [ ] Relevant flow charts

---

## Missing Documentation (None! ✅)

All essential documentation is complete:
- ✅ Architecture defined
- ✅ API spec documented
- ✅ Getting started guide ready
- ✅ Testing strategy complete
- ✅ Flow charts for all screens
- ✅ Keyboard shortcuts defined
- ✅ Implementation plan ready

---

## Next Actions

**You can start coding immediately:**

```bash
# 1. Set up project
cd golang
go mod init github.com/shinokada/tera

# 2. Install dependencies
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest

# 3. Create directory structure
mkdir -p cmd/tera internal/{api,storage,player,gist,ui/components} pkg/utils

# 4. Start with Phase 1: Data Models
# Follow GETTING_STARTED.md Step 1
```

**First file to create:**
`internal/api/models.go` (See GETTING_STARTED.md)

**First test to write:**
`internal/api/models_test.go` (See TESTING.md)

---

## Resources for Development

### Bubble Tea
- Tutorial: https://github.com/charmbracelet/bubbletea/tree/master/tutorials
- Examples: https://github.com/charmbracelet/bubbletea/tree/master/examples

### Go Testing
- Official guide: https://go.dev/doc/tutorial/add-a-test
- Table-driven tests: https://go.dev/wiki/TableDrivenTests

### Radio Browser API
- Documentation: https://api.radio-browser.info/
- OpenAPI spec: https://api.radio-browser.info/swagger.json

### GitHub Gist API
- Documentation: https://docs.github.com/en/rest/gists/gists
- Examples: https://github.com/google/go-github

---

## Success Metrics

You'll know the documentation is working when:
- [ ] Can set up project in <30 minutes
- [ ] Can implement Phase 1 without questions
- [ ] Tests pass on first try (using examples)
- [ ] Can find answers to implementation questions
- [ ] Can understand screen behavior from flow charts
- [ ] Can implement keyboard shortcuts correctly

---

## Questions?

If documentation is unclear:
1. Note what's confusing
2. Implement based on best understanding
3. Document assumptions
4. Can iterate on docs later

**Remember:** Docs are guides, not gospel. Adapt as you learn.
