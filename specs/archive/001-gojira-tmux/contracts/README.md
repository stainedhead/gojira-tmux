# Contracts

This directory contains the interface contracts for the gojira-tmux application.

## Purpose

These contracts define the boundaries between architectural layers:
- **Domain Layer**: Core entities (Issue, TeamMember, Project, Comment, User)
- **Port Interfaces**: JiraPort, AuthPort, TokenStorePort, ConfigPort
- **Use Case Interfaces**: ListIssuesUseCase, GetIssueDetailsUseCase, AuthenticateUseCase, SetupTokenUseCase
- **TUI Messages**: Message types for BubbleTea Update function

## Files

| File | Description |
|------|-------------|
| `ports.go` | Complete interface definitions for all ports and use cases |

## Implementation Mapping

| Interface | Implementation Location |
|-----------|------------------------|
| JiraPort | `internal/adapter/jira/client.go` |
| AuthPort | `internal/adapter/auth/okta.go` |
| TokenStorePort | `internal/adapter/auth/token_store.go` |
| ConfigPort | `internal/adapter/config/config.go` |
| ListIssuesUseCase | `internal/usecase/list_issues.go` |
| GetIssueDetailsUseCase | `internal/usecase/get_issue_details.go` |
| AuthenticateUseCase | `internal/usecase/authenticate.go` |
| SetupTokenUseCase | `internal/usecase/setup_token.go` |

## Usage

These contracts are used during implementation to:
1. Define clear boundaries between components
2. Enable unit testing with mock implementations
3. Ensure Clean Architecture dependency rules are followed

The actual Go interfaces should be placed in `internal/domain/ports.go` when implementing.
