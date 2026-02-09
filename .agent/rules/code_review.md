# Code Review Standards

When reviewing code for quality, security, and maintainability, follow these comprehensive guidelines.

## Core Responsibilities

1. **Code Quality Assessment**
   - Review code for adherence to coding standards and best practices
   - Evaluate readability, maintainability, and documentation quality
   - Identify code smells, anti-patterns, and technical debt
   - Assess error handling, validation, and edge case coverage
   - Run compile/typecheck/build scripts to check for issues

2. **Type Safety and Linting**
   - Perform thorough TypeScript type checking
   - Identify type safety issues and suggest stronger typing
   - Run appropriate linters and analyze results
   - Balance strict type safety with developer productivity

3. **Build and Deployment Validation**
   - Verify build processes execute successfully
   - Check for dependency issues or version conflicts
   - Validate deployment configurations
   - Ensure proper environment variable handling without exposing secrets
   - Confirm test coverage meets project standards

4. **Performance Analysis**
   - Identify performance bottlenecks and inefficient algorithms
   - Review database queries for optimization opportunities
   - Analyze memory usage patterns and potential leaks
   - Evaluate async/await usage and promise handling
   - Suggest caching strategies where appropriate

5. **Security Audit**
   - Identify common security vulnerabilities (OWASP Top 10)
   - Review authentication and authorization implementations
   - Check for SQL injection, XSS, and other injection vulnerabilities
   - Verify proper input validation and sanitization
   - Ensure sensitive data is properly protected
   - Validate CORS, CSP, and other security headers

## Review Process

1. **Initial Analysis**: Focus on recently changed files unless asked to review entire codebase
2. **Systematic Review**: Work through each concern area methodically
3. **Prioritization**: Categorize findings by severity:
   - **Critical**: Security vulnerabilities, data loss risks, breaking changes
   - **High**: Performance issues, type safety problems, missing error handling
   - **Medium**: Code smells, maintainability concerns, documentation gaps
   - **Low**: Style inconsistencies, minor optimizations

4. **Actionable Recommendations**: For each issue:
   - Clearly explain the problem and its potential impact
   - Provide specific code examples of how to fix it
   - Suggest alternative approaches when applicable
   - Reference relevant best practices or documentation

## Quality Standards

- Be constructive and educational in feedback
- Acknowledge good practices and well-written code
- Provide context for why certain practices are recommended
- Balance ideal practices with pragmatic solutions
- Focus on issues that truly matter for code quality, security, and maintainability
