# Code Simplification Guidelines

When simplifying and refining code, follow these principles to enhance clarity, consistency, and maintainability while preserving all functionality.

## Core Principles

1. **Preserve Functionality**: Never change what the code does—only how it does it. All original features, outputs, and behaviors must remain intact.

2. **Apply Project Standards**: Follow established coding standards from project documentation. Adapt to the project's language, framework, and conventions.

3. **Enhance Clarity**: Simplify code structure by:
   - Reducing unnecessary complexity and nesting
   - Eliminating redundant code and abstractions
   - Improving readability through clear variable and function names
   - Consolidating related logic
   - Removing unnecessary comments that describe obvious code
   - Avoiding deeply nested conditionals—prefer early returns or guard clauses
   - Choosing clarity over brevity—explicit code is better than compact code

4. **Maintain Balance**: Avoid over-simplification that could:
   - Reduce code clarity or maintainability
   - Create overly clever solutions hard to understand
   - Combine too many concerns into single functions/components
   - Remove helpful abstractions that improve organization
   - Prioritize "fewer lines" over readability
   - Make the code harder to debug or extend

5. **Focus Scope**: Only refine recently modified code unless explicitly instructed to review a broader scope.

## Refinement Process

1. Identify the recently modified code sections
2. Analyze for opportunities to improve elegance and consistency
3. Apply project-specific best practices and coding standards
4. Ensure all functionality remains unchanged
5. Verify the refined code is simpler and more maintainable
6. Run appropriate verification (typecheck, linter, tests) if available

## Quality Standards

- Prioritize readable, explicit code over overly compact solutions
- Ensure all refactoring preserves exact functionality
- Run verification tools to confirm no regressions
- Focus on meaningful improvements, not cosmetic changes
