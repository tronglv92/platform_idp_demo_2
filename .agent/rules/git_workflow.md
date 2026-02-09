# Git Workflow and Version Control

When managing git operations, commits, and version control, follow these streamlined guidelines.

## Core Principles

- Execute git operations efficiently in 2-4 tool calls
- Use conventional commit messages
- Ensure token efficiency while maintaining quality

## Git Operations

### Staging and Committing
- Stage relevant files using `git add`
- Write clear, conventional commit messages following format:
  - `feat: add new feature`
  - `fix: resolve bug in component`
  - `docs: update documentation`
  - `refactor: improve code structure`
  - `test: add test coverage`
  - `chore: update dependencies`

### Commit Message Format
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Best Practices
- Write descriptive commit messages that explain the "why"
- Keep commits focused on a single logical change
- Review changes before committing
- Ensure code builds and tests pass before committing
- Use meaningful branch names
- Keep commit history clean and readable

## Workflow

1. **Review Changes**: Check what files have been modified
2. **Stage Files**: Add relevant files to staging area
3. **Commit**: Create commit with conventional message
4. **Push**: Push changes to remote repository (if requested)

## Quality Standards

- Commit messages should be clear and descriptive
- Each commit should represent a logical unit of work
- Avoid committing sensitive information (secrets, credentials)
- Ensure code quality before committing
- Follow project-specific git conventions
