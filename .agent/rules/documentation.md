# Documentation Management

When creating, maintaining, or organizing technical documentation, follow these comprehensive guidelines.

## Core Responsibilities

### 1. Documentation Standards & Implementation Guidelines
Establish and maintain implementation standards including:
- Codebase structure documentation with clear architectural patterns
- Error handling patterns and best practices
- API design guidelines and conventions
- Testing strategies and coverage requirements
- Security protocols and compliance requirements

### 2. Documentation Analysis & Maintenance
Systematically:
- Read and analyze all existing documentation files
- Identify gaps, inconsistencies, or outdated information
- Cross-reference documentation with actual codebase implementation
- Ensure documentation reflects the current state of the system
- Maintain a clear documentation hierarchy and navigation structure

### 3. Code-to-Documentation Synchronization
When codebase changes occur:
- Analyze the nature and scope of changes
- Identify all documentation that requires updates
- Update API documentation, configuration guides, and integration instructions
- Ensure examples and code snippets remain functional and relevant
- Document breaking changes and migration paths

### 4. Product Development Requirements (PDRs)
Create and maintain PDRs that:
- Define clear functional and non-functional requirements
- Specify acceptance criteria and success metrics
- Include technical constraints and dependencies
- Provide implementation guidance and architectural decisions
- Track requirement changes and version history

### 5. Developer Productivity Optimization
Organize documentation to:
- Minimize time-to-understanding for new developers
- Provide quick reference guides for common tasks
- Include troubleshooting guides and FAQ sections
- Maintain up-to-date setup and deployment instructions
- Create clear onboarding documentation

## Documentation Accuracy Protocol

**Principle**: Only document what you can verify exists in the codebase.

### Evidence-Based Writing
Before documenting any code reference:
1. **Functions/Classes**: Verify via code search
2. **API Endpoints**: Confirm routes exist in route files
3. **Config Keys**: Check against `.env.example` or config files
4. **File References**: Confirm file exists before linking

### Conservative Output Strategy
- When uncertain about implementation details → describe high-level intent only
- When code is ambiguous → note "implementation may vary"
- Never invent API signatures, parameter names, or return types
- Don't assume endpoints exist; verify or omit

### Internal Link Hygiene
- Only use `[text](./path.md)` for files that exist in `docs/`
- For code files, verify path before documenting
- Prefer relative links within `docs/`

## Documentation File Standards

- Use clear, descriptive filenames following project conventions
- Maintain consistent Markdown formatting
- Include proper headers, table of contents, and navigation
- Add metadata (last updated, version, author) when relevant
- Use code blocks with appropriate syntax highlighting
- Ensure all variables, function names, class names use correct case (PascalCase, camelCase, snake_case)

## Key Documentation Files

Create or update these essential files:
- `project-overview-pdr.md`: Comprehensive project overview and PDR
- `code-standards.md`: Codebase structure and code standards
- `system-architecture.md`: System architecture documentation
- `api-docs.md`: API documentation (if applicable)

## Best Practices

1. **Clarity Over Completeness**: Write documentation that is immediately useful
2. **Examples First**: Include practical examples before technical details
3. **Progressive Disclosure**: Structure information from basic to advanced
4. **Maintenance Mindset**: Write documentation that is easy to update
5. **User-Centric**: Always consider documentation from reader's perspective
