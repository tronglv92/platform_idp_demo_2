# Debugging and Issue Investigation

When investigating issues, analyzing system behavior, or diagnosing problems, follow these systematic approaches.

## Core Competencies

- **Issue Investigation**: Systematically diagnosing and resolving incidents using methodical debugging
- **System Behavior Analysis**: Understanding complex system interactions, identifying anomalies, tracing execution flows
- **Database Diagnostics**: Querying databases for insights, examining table structures, analyzing query performance
- **Log Analysis**: Collecting and analyzing logs from servers, CI/CD pipelines, and application layers
- **Performance Optimization**: Identifying bottlenecks, developing optimization strategies
- **Test Execution & Analysis**: Running tests for debugging, analyzing failures, identifying root causes

## Investigation Methodology

### 1. Initial Assessment
- Gather symptoms and error messages
- Identify affected components and timeframes
- Determine severity and impact scope
- Check for recent changes or deployments

### 2. Data Collection
- Query relevant databases using appropriate tools
- Collect server logs from affected time periods
- Retrieve CI/CD pipeline logs
- Examine application logs and error traces
- Capture system metrics and performance data

### 3. Analysis Process
- Correlate events across different log sources
- Identify patterns and anomalies
- Trace execution paths through the system
- Analyze database query performance and table structures
- Review test results and failure patterns

### 4. Root Cause Identification
- Use systematic elimination to narrow down causes
- Validate hypotheses with evidence from logs and metrics
- Consider environmental factors and dependencies
- Document the chain of events leading to the issue

### 5. Solution Development
- Design targeted fixes for identified problems
- Develop performance optimization strategies
- Create preventive measures to avoid recurrence
- Propose monitoring improvements for early detection

## Tools and Techniques

- **Database Tools**: Query analyzers for performance insights
- **Log Analysis**: grep, awk, sed for log parsing
- **Performance Tools**: Profilers, APM tools, system monitoring utilities
- **Testing Frameworks**: Run unit tests, integration tests, diagnostic scripts
- **CI/CD Tools**: Pipeline debugging and log analysis

## Best Practices

- Always verify assumptions with concrete evidence
- Consider the broader system context when analyzing issues
- Document investigation process for knowledge sharing
- Prioritize solutions based on impact and implementation effort
- Ensure recommendations are specific, measurable, and actionable
- Test proposed fixes in appropriate environments before deployment
- Consider security implications of both issues and solutions
