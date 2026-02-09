# Security & Team Member Updates - Implementation Plan

**Created:** 2026-02-09
**Version:** 1.0
**Status:** Planning
**Estimated Duration:** 12-18 hours (6 phases)

---

## Development Approach

### Methodology

**TDD + Bottom-Up Clean Architecture**

```
Red → Green → Refactor (for each component)
  ↓
Domain → Adapter → Use Case → Infrastructure
  ↓
Integration → Testing → Documentation
```

**Why This Approach?**
- TDD ensures all code is tested and testable
- Bottom-up allows independent development of layers
- Clean Architecture maintains separation of concerns
- Enables parallelization (multiple workers on different layers)

**Key Principles:**
1. Write tests first (Red-Green-Refactor)
2. No layer depends on outer layers
3. Each phase produces working, tested code
4. Commit after each completed task

**Incremental Milestones:**
- Phase 1: Domain models compile and pass tests
- Phase 2: Adapters implement interfaces with passing tests
- Phase 3: Use cases orchestrate adapters with passing tests
- Phase 4: TUI displays new features correctly
- Phase 5: Full app runs with new auth
- Phase 6: All tests pass, documentation complete

---

## Phase Breakdown

### Phase 1: Domain Model Updates (1.5 hours)

**Goal:** Update domain layer with simplified auth and team member aliases

**Deliverables:**
- Simplified `AuthPort` interface
- Enhanced `TeamMember` with alias support
- New `AtlassianConfig` struct
- Updated `Config` struct
- Simplified `User` model
- All domain tests passing

**Tasks:**
- P1.1: Update Domain Ports (30min)
- P1.2: Update Team Member Model (45min)
- P1.3: Update User Model (15min)

**Dependencies:**
- None (can start immediately)

**Acceptance Criteria:**
- [ ] `AuthPort` has 2 methods (ValidateToken, IsTokenValid)
- [ ] `OktaConfig` removed, `AtlassianConfig` added
- [ ] `TeamMember` has Alias field and matching methods
- [ ] `TokenStorePort` has 4 methods (refresh methods removed)
- [ ] `User` has no session management
- [ ] All domain tests pass (`go test ./internal/domain`)
- [ ] Code coverage >90%

**Quality Gates:**
- [ ] All tests passing
- [ ] Code formatted
- [ ] Documentation updated
- [ ] Committed

---

### Phase 2: Adapter Layer (4 hours)

**Goal:** Implement Atlassian adapter, remove Okta, update config validation

**Deliverables:**
- New `AtlassianAdapter` with token validation
- Deleted Okta and callback files
- Updated config loader with alias validation
- Updated JQL builder with alias matching
- All adapter tests passing

**Tasks:**
- P2.1: Create Atlassian Auth Adapter (2h)
- P2.2: Delete Okta Components (10min)
- P2.3: Update Config Loader (1h)
- P2.4: Update JQL Builder (45min)

**Dependencies:**
- Phase 1 complete (domain interfaces defined)

**Acceptance Criteria:**
- [ ] `AtlassianAdapter` validates tokens via `/rest/api/2/myself`
- [ ] Handles 401, 403, network errors gracefully
- [ ] Okta files deleted (okta.go, okta_test.go, callback.go, callback_test.go)
- [ ] Config validation checks alias uniqueness
- [ ] Config validation checks duplicate emails
- [ ] JQL builder supports alias matching
- [ ] All adapter tests pass
- [ ] Code coverage >85%

**Quality Gates:**
- [ ] All tests passing
- [ ] go mod tidy (OAuth deps removed)
- [ ] Code formatted
- [ ] Committed

---

### Phase 3: Use Case Layer (1.5 hours)

**Goal:** Simplify authentication use case to token-only flow

**Deliverables:**
- Simplified `Authenticate` use case
- Updated `SetupToken` use case
- All use case tests passing

**Tasks:**
- P3.1: Simplify Authenticate Use Case (1h)
- P3.2: Update Setup Token Use Case (30min)

**Dependencies:**
- Phase 2 complete (adapters implemented)

**Acceptance Criteria:**
- [ ] `Authenticate` has 3 methods (ValidateAndSaveToken, HasValidToken, ClearToken)
- [ ] OAuth methods removed (StartLogin, CompleteLogin, etc.)
- [ ] `SetupToken` validates email format
- [ ] All use case tests pass
- [ ] Code coverage >90%

**Quality Gates:**
- [ ] All tests passing
- [ ] Code formatted
- [ ] Committed

---

### Phase 4: Infrastructure Layer (4.5 hours)

**Goal:** Update TUI for token input and alias display

**Deliverables:**
- Redesigned setup screen (token input form)
- Updated/removed login screen
- Updated filter bar (alias display)
- Updated app initialization
- Manual TUI testing complete

**Tasks:**
- P4.1: Update Setup Screen (2h)
- P4.2: Update/Remove Login Screen (1h)
- P4.3: Update Filter Bar (1h)
- P4.4: Update App Initialization (30min)

**Dependencies:**
- Phase 3 complete (use cases available)

**Acceptance Criteria:**
- [ ] Setup screen shows email + token input form
- [ ] Setup screen displays link to token generation
- [ ] Token input is masked
- [ ] Validation errors display clearly
- [ ] Filter bar shows "Name (Alias)" format
- [ ] Filtering by alias works
- [ ] App initialization uses AtlassianAdapter
- [ ] Manual testing checklist complete

**Quality Gates:**
- [ ] App builds successfully
- [ ] Manual testing passes
- [ ] All tests passing
- [ ] Committed

---

### Phase 5: Entry Point & Configuration (1.5 hours)

**Goal:** Update main entry point and documentation

**Deliverables:**
- Updated `main.go` using AtlassianAdapter
- Updated README with new auth
- Migration guide created
- Example config file
- Dependencies cleaned up

**Tasks:**
- P5.1: Update Main Entry Point (45min)
- P5.2: Update Configuration Examples (30min)
- P5.3: Update Dependencies (15min)

**Dependencies:**
- Phase 4 complete (full stack working)

**Acceptance Criteria:**
- [ ] `main.go` uses `NewAtlassianAdapter`
- [ ] `main.go` uses `cfg.Atlassian.Email`
- [ ] README updated (remove Okta, add Atlassian)
- [ ] Migration guide complete with examples
- [ ] `config.example.yaml` created
- [ ] `go.mod` has no OAuth dependencies
- [ ] Full app builds and runs

**Quality Gates:**
- [ ] All tests passing
- [ ] go build succeeds
- [ ] go mod verify succeeds
- [ ] Manual app test succeeds
- [ ] Committed

---

### Phase 6: Testing & Documentation (5 hours)

**Goal:** Comprehensive testing and documentation

**Deliverables:**
- Integration test suite
- Manual testing complete
- Architecture docs updated
- Migration guide polished
- All documentation reviewed

**Tasks:**
- P6.1: Integration Testing (3h)
- P6.2: Update Documentation (2h)

**Dependencies:**
- Phase 5 complete (app fully functional)

**Acceptance Criteria:**
- [ ] Integration tests cover:
  - First-run token setup
  - Token validation (valid/invalid)
  - Alias filtering
  - Backward compatibility (no aliases)
- [ ] Manual testing checklist complete
- [ ] All error scenarios tested
- [ ] Architecture docs updated
- [ ] Migration guide includes screenshots
- [ ] README reflects new auth
- [ ] All tests pass (unit + integration)
- [ ] Code coverage >85% overall

**Quality Gates:**
- [ ] All tests passing
- [ ] Documentation review complete
- [ ] No TODOs in code
- [ ] Committed

---

## Critical Path

**Critical Path Tasks:**
These tasks must complete in sequence and block all other work:

```
Task P1.1 → Task P2.1 → Task P3.1 → Task P4.1 → Task P5.1 → Task P6.1
  [30min]    [2h]       [1h]        [2h]        [45min]     [3h]

Total Critical Path Duration: ~9 hours
```

**Parallel Work Opportunities:**
- Phase 1: P1.2 and P1.3 can run parallel to P1.1
- Phase 2: P2.3 and P2.4 can run parallel after P2.1
- Phase 4: P4.3 can run parallel to P4.1 and P4.2
- Phase 6: Documentation work (P6.2) can start during P6.1

**Blocking Dependencies:**
- P2.x blocks on P1.x (need domain interfaces)
- P3.x blocks on P2.x (need adapter implementations)
- P4.x blocks on P3.x (need use cases)
- P5.x blocks on P4.x (need working TUI)
- P6.x blocks on P5.x (need complete app)

---

## Dependency Graph

**Visual Dependency Map:**
```
Phase 1: Domain Layer
    ├─ Task P1.1 (no dependencies) ──┐
    ├─ Task P1.2 (no dependencies) ──┤
    └─ Task P1.3 (no dependencies) ──┘
          ↓
Phase 2: Adapter Layer
    ├─ Task P2.1 (depends on P1.1) ──┐
    ├─ Task P2.2 (depends on P2.1) ──┤
    ├─ Task P2.3 (depends on P1.1, P1.2) ─┤
    └─ Task P2.4 (depends on P1.2) ──┘
          ↓
Phase 3: Use Case Layer
    ├─ Task P3.1 (depends on P2.1, P2.3)
    └─ Task P3.2 (depends on P2.3)
          ↓
Phase 4: Infrastructure
    ├─ Task P4.1 (depends on P3.1, P3.2)
    ├─ Task P4.2 (depends on P3.1)
    ├─ Task P4.3 (depends on P1.2)
    └─ Task P4.4 (depends on P4.1, P4.2)
          ↓
Phase 5: Entry Point
    ├─ Task P5.1 (depends on P4.4)
    ├─ Task P5.2 (no dependencies)
    └─ Task P5.3 (depends on P2.2)
          ↓
Phase 6: Testing
    ├─ Task P6.1 (depends on P5.1)
    └─ Task P6.2 (can start during P6.1)
```

**External Dependencies:**
- None (all dependencies internal to project)

---

## Testing Strategy

### Unit Testing

**Approach:**
- Test-first (TDD): Write test before implementation
- One test file per implementation file
- Aim for >90% code coverage in domain/use case, >85% overall

**Test Organization:**
```
internal/domain/team_member_test.go       # Alias matching tests
internal/adapter/auth/atlassian_test.go   # Token validation tests
internal/adapter/config/config_test.go    # Alias uniqueness tests
internal/usecase/authenticate_test.go     # Use case tests
```

**Key Test Scenarios:**
- **Team Member Alias Matching**:
  - Exact alias match (case-sensitive)
  - Exact name match
  - Case-insensitive alias match
  - Case-insensitive name match
  - No match (returns nil)

- **Token Validation**:
  - Valid token (200 OK)
  - Invalid credentials (401)
  - Insufficient permissions (403)
  - Network errors
  - CAPTCHA lockout
  - Malformed responses

- **Config Validation**:
  - Valid config with aliases
  - Duplicate alias error
  - Invalid alias format (with spaces)
  - Missing Atlassian email
  - Backward compatibility (no aliases)

### Integration Testing

**Approach:**
- Test component interactions
- Use real implementations (not mocks where possible)
- Test with mock HTTP server for Jira API

**Integration Test Suites:**
1. **Authentication Flow**: Setup screen → token input → validation → main screen
2. **Alias Filtering**: Load config → filter by alias → verify JQL → verify results
3. **Migration Flow**: Old config detected → guidance displayed → new config works

### End-to-End Testing

**Approach:**
- Full application lifecycle
- Real user workflows
- Manual testing with checklist

**E2E Test Scenarios:**
1. **First-time setup**: No token → setup screen → enter token → success
2. **Invalid token**: Enter wrong token → error message → retry → success
3. **Alias usage**: Filter by "JohnA" → correct member selected → issues load
4. **Backward compat**: Config without aliases → app works normally

**Location:** Manual testing checklist in `status.md`

### Performance Testing

**Benchmarks:**
- Token validation: <500ms
- Team member matching: <1ms
- App startup: <2 seconds

### Security Testing

**Security Checks:**
- [ ] Tokens never logged
- [ ] Tokens masked in UI
- [ ] Tokens stored in keychain only
- [ ] HTTPS enforced for Jira URL
- [ ] Error messages don't expose token fragments

---

## Rollout Strategy

### Development Environment

**Phase:** Initial development and testing

**Criteria:**
- All unit tests passing
- Integration tests passing
- Manual testing checklist complete

**Rollback:** Not applicable (dev only)

---

### Production Environment

**Phase:** Production release

**Deployment Strategy:**
- [ ] Tag release (v2.0.0 - breaking change)
- [ ] Create GitHub release with migration guide
- [ ] Update README on GitHub
- [ ] Notify users via release notes

**Go-Live Checklist:**
- [ ] All tests passing
- [ ] Migration guide complete
- [ ] Example config file included
- [ ] README updated
- [ ] Release notes written
- [ ] Breaking change clearly documented

**Rollback Plan:**
1. Users can keep using previous version
2. Old config files still work with previous version
3. No database migrations (config file only)
4. Tokens stored separately (no conflict)

---

## Success Metrics

### Development Metrics

**Code Quality:**
- Test coverage: Target >85% overall, >90% domain/use case
- Linter warnings: Target 0
- Code review approvals: At least 1

**Velocity:**
- Estimated total: 12-18 hours
- Target: Complete in 2-3 days full-time

### Release Metrics

**Adoption:**
- Migration guide views
- Issue tickets (target <5 migration issues)

**Performance:**
- App startup <2 seconds
- Token validation <500ms
- No performance regressions

**Reliability:**
- No critical bugs in first week
- Migration success rate >95%

---

## Timeline Summary

**Total Estimated Duration:** 12-18 hours (2-3 days full-time)

**Phase Breakdown:**
- Phase 1: 1.5 hours - Domain model updates
- Phase 2: 4 hours - Adapter layer (Atlassian adapter, config, JQL)
- Phase 3: 1.5 hours - Use case simplification
- Phase 4: 4.5 hours - TUI updates (setup, filter bar, app)
- Phase 5: 1.5 hours - Main entry point and docs
- Phase 6: 5 hours - Testing and documentation

**Key Milestones:**
- Day 1 AM: Phases 1-2 complete (domain + adapters)
- Day 1 PM: Phase 3 complete (use cases)
- Day 2 AM: Phase 4 complete (TUI working)
- Day 2 PM: Phase 5 complete (full app runs)
- Day 3: Phase 6 complete (all tests pass, docs ready)

**Contingency:**
- Buffer: 50% (6 hours for unknown unknowns)
- Total with buffer: 18-24 hours (3-4 days full-time)

---

## Next Steps

1. ✅ Review and approve this plan
2. ✅ Create detailed task breakdown (tasks.md)
3. **Next**: Begin Phase 1 implementation
   - Start with P1.1 (Update Domain Ports)
   - Follow TDD workflow (Red-Green-Refactor)
   - Update status.md after each task
4. Daily progress tracking
5. Update status.md as phases complete

---

## Notes

**Assumptions:**
- User has access to Atlassian account for token generation
- Jira instance uses Atlassian Cloud (not server/data center)
- Team size <100 members (for O(n) matching performance)
- Users accept breaking change for simpler auth

**Open Questions:**
- Should we support `JIRA_EMAIL` environment variable? **Proposed: No, keep in config for consistency**
- Maximum alias length? **Proposed: 20 characters**
- Should we validate aliases against Jira usernames? **Proposed: No, keep aliases as UI-only**

**Constraints:**
- Must maintain backward compatibility for team members without aliases
- Must not break existing Jira API queries
- Must support macOS, Linux, Windows keychain storage
- Must follow Clean Architecture principles
