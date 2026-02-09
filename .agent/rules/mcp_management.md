# MCP (Model Context Protocol) Management

When working with MCP server integrations, discovering tools, or executing MCP capabilities, follow these guidelines.

## Purpose

Execute tasks using MCP tools while keeping the main context window clean. Handle MCP discovery and execution in a focused manner.

## Core Responsibilities

1. **Execute via Primary Method**: Attempt task execution using available MCP tools
2. **Fallback Strategies**: Use alternative execution methods if primary fails
3. **Report Results**: Provide concise execution summary
4. **Error Handling**: Report failures with actionable guidance

## Operational Guidelines

- **Context Efficiency**: Keep responses concise
- **Multi-Server**: Handle tools across multiple MCP servers
- **Error Handling**: Report errors clearly with guidance
- **Token Efficiency**: Maintain high quality while being efficient

## Execution Workflow

1. **Receive Task**: Understand what MCP operation is needed
2. **Check Availability**: Verify available MCP tools and servers
3. **Execute**: Run the appropriate MCP tool or operation
4. **Report**: Send concise summary (status, output, artifacts, errors)

## Result Reporting

Concise summaries should include:
- Execution status (success/failure)
- Output/results
- File paths for artifacts (screenshots, etc.)
- Error messages with guidance

## Best Practices

- Keep MCP operations focused and specific
- Report results clearly and concisely
- Handle errors gracefully with helpful messages
- Maintain clean context by avoiding verbose output
- Use appropriate MCP tools for the task at hand

## Quality Standards

- Execute MCP operations efficiently
- Provide clear status reporting
- Handle errors with actionable guidance
- Keep context clean and focused
- Ensure results are properly documented
