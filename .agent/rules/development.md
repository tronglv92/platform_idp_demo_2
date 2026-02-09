# Development and Implementation

When implementing features, follow these execution guidelines to ensure quality and maintainability.

## Core Responsibilities

- Follow YAGNI, KISS, DRY principles
- Ensure token efficiency while maintaining quality
- Respect project-specific coding standards and development rules
- Write clean, maintainable code following project standards
- Add necessary tests for implemented functionality

## Execution Process

### 1. Pre-Implementation Validation
- Read project documentation (codebase summary, code standards, system architecture)
- Verify all dependencies from previous phases are complete
- Check if files exist or need creation
- Understand requirements and acceptance criteria

### 2. Implementation
- Execute implementation steps sequentially as planned
- Follow architecture and requirements exactly as specified
- Write clean, maintainable code
- Handle errors appropriately
- Add comprehensive logging where needed

### 3. Quality Assurance
- Run type checks (e.g., `npm run typecheck`)
- Run tests (e.g., `npm test`)
- Fix any type errors or test failures
- Verify success criteria are met
- Ensure code builds successfully

### 4. Code Standards
- Use consistent naming conventions
- Follow project's file organization structure
- Write self-documenting code with clear variable/function names
- Add comments only where code intent is not obvious
- Ensure proper error handling and validation

## Best Practices

- **File Ownership**: Only modify files within assigned scope
- **Parallel Execution Safety**: Work independently, use well-defined interfaces
- **Testing**: Write tests alongside implementation, not after
- **Documentation**: Update relevant documentation as you implement
- **Performance**: Consider performance implications of implementation choices
- **Security**: Follow security best practices, validate all inputs
