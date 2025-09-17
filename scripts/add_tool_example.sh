#!/bin/bash

# Example: Adding a new tool to the JSON-driven MCP server
# This script demonstrates how easy it is to add new functionality

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_JSON="$SCRIPT_DIR/tools.json"

echo "Adding a new 'echo_message' tool to tools.json..."

# Create a backup
cp "$TOOLS_JSON" "$TOOLS_JSON.backup"

# Add the new tool using jq
jq '.tools += [{
  "name": "echo_message",
  "description": "Echo a message with optional prefix and suffix.\n\nArgs:\n    message: The message to echo\n    prefix: Optional prefix (default: \">>\")\n    suffix: Optional suffix (default: \"<<\")\n\nReturns:\n    The formatted message",
  "inputSchema": {
    "type": "object",
    "properties": {
      "message": {
        "title": "Message",
        "type": "string",
        "description": "The message to echo"
      },
      "prefix": {
        "title": "Prefix",
        "type": "string",
        "description": "Optional prefix for the message"
      },
      "suffix": {
        "title": "Suffix", 
        "type": "string",
        "description": "Optional suffix for the message"
      }
    },
    "required": ["message"]
  },
  "command": "echo \"{{prefix:=>>}} {{message}} {{suffix:=<<}}\""
}]' "$TOOLS_JSON" > "$TOOLS_JSON.tmp" && mv "$TOOLS_JSON.tmp" "$TOOLS_JSON"

echo "New tool added! Testing it..."

# Test the new tool
echo "1. Testing with just message:"
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"echo_message","arguments":{"message":"Hello World"}},"id":1}' | "$SCRIPT_DIR/make_mcp.sh" | jq -r '.result.content[0].text'

echo -e "\n2. Testing with custom prefix and suffix:"
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"echo_message","arguments":{"message":"Hello World","prefix":"[INFO]","suffix":"[END]"}},"id":2}' | "$SCRIPT_DIR/make_mcp.sh" | jq -r '.result.content[0].text'

echo -e "\n3. Testing with only custom prefix:"
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"echo_message","arguments":{"message":"Hello World","prefix":"***"}},"id":3}' | "$SCRIPT_DIR/make_mcp.sh" | jq -r '.result.content[0].text'

echo -e "\nNew tool working perfectly! No changes to make_mcp.sh were needed."
echo "Backup saved as tools.json.backup"
