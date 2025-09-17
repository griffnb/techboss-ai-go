# JSON-Driven MCP Server

This is a completely JSON-driven Model Context Protocol (MCP) server built in bash. All tools are defined in a JSON configuration file with shell command templates, eliminating the need to modify the server script when adding new tools.

## Files

- `make_mcp.sh` - The main MCP server script (no modifications needed for new tools)
- `tools.json` - Configuration file defining all available tools and their commands
- `test_mcp.sh` - Test script to verify the server works
- `README.md` - This documentation file

## Key Features

🎯 **100% JSON-Driven**: Add new tools by only editing `tools.json`  
🔄 **Parameter Substitution**: Use `{{param}}` syntax in command templates  
⚙️ **Default Values**: Support for optional parameters with defaults  
🛡️ **Error Handling**: Automatic parameter validation and error reporting  
📝 **Logging**: Debug information logged to `/tmp/mcp-server-bash.log`

## Usage

### Running the Server

```bash
./make_mcp.sh
```

The server reads JSON-RPC requests from stdin and writes responses to stdout.

### Testing the Server

```bash
./test_mcp.sh
```

This will run a series of test calls to verify the server is working correctly.

## Adding New Tools

To add new tools, simply edit the `tools.json` file. **No modifications to `make_mcp.sh` are needed!**

### Tool Definition Format

```json
{
  "name": "tool_name",
  "description": "Description of what the tool does",
  "inputSchema": {
    "type": "object",
    "properties": {
      "param1": {
        "title": "Parameter 1",
        "type": "string",
        "description": "Description of parameter 1"
      }
    },
    "required": ["param1"]
  },
  "command": "shell command with {{param1}} substitution"
}
```

### Parameter Substitution

The system supports two types of parameter substitution:

1. **Required Parameters**: `{{param_name}}`
   - Must be provided in the tool call
   - Will cause an error if missing

2. **Optional Parameters with Defaults**: `{{param_name:=default_value}}`
   - Uses the provided value if given
   - Falls back to default_value if not provided

### Example Tools

#### Simple Command (No Parameters)
```json
{
  "name": "git_status",
  "description": "Show git status",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "required": []
  },
  "command": "cd /path/to/repo && git status"
}
```

#### Command with Required Parameters
```json
{
  "name": "file_read",
  "description": "Read file contents",
  "inputSchema": {
    "type": "object",
    "properties": {
      "filepath": {
        "title": "File Path",
        "type": "string",
        "description": "Path to the file to read"
      }
    },
    "required": ["filepath"]
  },
  "command": "cat \"{{filepath}}\""
}
```

#### Command with Optional Parameters
```json
{
  "name": "run_tests",
  "description": "Run Go tests",
  "inputSchema": {
    "type": "object",
    "properties": {
      "package": {
        "title": "Package",
        "type": "string",
        "description": "Package to test (optional)"
      }
    },
    "required": []
  },
  "command": "cd /path/to/project && go test {{package:=./...}} -v"
}
```

#### Math Operations
```json
{
  "name": "addition",
  "description": "Add two numbers",
  "inputSchema": {
    "type": "object",
    "properties": {
      "num1": {"title": "Number 1", "type": "string"},
      "num2": {"title": "Number 2", "type": "string"}
    },
    "required": ["num1", "num2"]
  },
  "command": "echo \"$(({{num1}} + {{num2}}))\""
}
```

## Current Available Tools

The server comes pre-configured with these tools:

1. **format** - Format code using `make fmt`
2. **addition** - Add two numbers
3. **subtraction** - Subtract two numbers  
4. **multiplication** - Multiply two numbers
5. **file_read** - Read file contents
6. **directory_list** - List directory contents
7. **git_status** - Show git status
8. **run_tests** - Run Go tests with optional package selection
9. **build_project** - Build the Go project

## Example Tool Calls

### No Parameters
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "format",
    "arguments": {}
  },
  "id": 1
}
```

### With Required Parameters
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "addition",
    "arguments": {
      "num1": "5",
      "num2": "3"
    }
  },
  "id": 2
}
```

### With Optional Parameters
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "run_tests",
    "arguments": {
      "package": "./cmd/server"
    }
  },
  "id": 3
}
```

### Using Default Values
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "run_tests",
    "arguments": {}
  },
  "id": 4
}
```

## Security Considerations

⚠️ **Command Injection**: Be careful with parameter substitution to avoid shell injection attacks. The system doesn't sanitize input - ensure your command templates properly quote parameters.

✅ **Best Practices**:
- Always quote parameters: `"{{filepath}}"` not `{{filepath}}`
- Validate input in your tool descriptions
- Use absolute paths where possible
- Avoid user-controlled command construction

## Error Handling

The server automatically handles:
- Missing required parameters
- Tool not found errors
- Command execution failures
- JSON parsing errors

All errors are returned in standard MCP error format.

## Logging

Debug information is logged to `/tmp/mcp-server-bash.log` including:
- Tool calls and parameters
- Parameter substitution details  
- Command execution and results
- Error conditions

## Dependencies

- `jq` - For JSON parsing and manipulation
- `bash` - Shell interpreter (version 4+ recommended)

## MCP Protocol Compliance

This server implements the Model Context Protocol 2024-11-05 specification and supports:

- `initialize` - Server initialization
- `tools/list` - List available tools dynamically from JSON
- `tools/call` - Execute tool commands with parameter substitution
- `resources/list` - List resources (empty)
- `prompts/list` - List prompts (empty)
