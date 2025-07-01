
package serverless

const ServerlessWorkSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "Serverless Work Configuration",
  "description": "Configuration schema for serverless function execution",
  "required": ["provider", "function"],
  "properties": {
    "provider": {
      "type": "string",
      "enum": ["aws", "gcp", "azure"],
      "description": "Cloud provider for serverless execution"
    },
    "function": {
      "type": "string",
      "description": "Name of the serverless function to invoke"
    },
    "region": {
      "type": "string",
      "description": "Cloud region where the function is deployed"
    },
    "credentials": {
      "type": "object",
      "properties": {
        "type": {
          "type": "string",
          "enum": ["aws_iam", "aws_role", "service_account", "managed_identity", "service_principal"]
        },
        "access_key": {"type": "string"},
        "secret_key": {"type": "string"},
        "session_token": {"type": "string"},
        "profile": {"type": "string"},
        "role_arn": {"type": "string"},
        "key_file": {"type": "string"},
        "project_id": {"type": "string"},
        "client_id": {"type": "string"},
        "client_secret": {"type": "string"},
        "tenant_id": {"type": "string"},
        "properties": {
          "type": "object",
          "additionalProperties": {"type": "string"}
        }
      }
    },
    "timeout": {
      "type": "string",
      "pattern": "^[0-9]+[smh]$",
      "description": "Function execution timeout"
    },
    "environment": {
      "type": "object",
      "additionalProperties": {"type": "string"},
      "description": "Environment variables for the function"
    },
    "runtime": {
      "type": "string",
      "description": "Runtime environment for the function"
    },
    "memory": {
      "type": "integer",
      "minimum": 128,
      "maximum": 10240,
      "description": "Memory allocation for the function in MB"
    },
    "async": {
      "type": "boolean",
      "default": false,
      "description": "Whether to invoke the function asynchronously"
    },
    "qualifier": {
      "type": "string",
      "description": "Function version or alias to invoke"
    }
  },
  "allOf": [
    {
      "if": {"properties": {"provider": {"const": "aws"}}},
      "then": {
        "required": ["region"],
        "properties": {
          "memory": {"minimum": 128, "maximum": 10240}
        }
      }
    },
    {
      "if": {"properties": {"provider": {"const": "gcp"}}},
      "then": {
        "required": ["region"],
        "properties": {
          "memory": {"minimum": 128, "maximum": 8192}
        }
      }
    },
    {
      "if": {"properties": {"provider": {"const": "azure"}}},
      "then": {
        "properties": {
          "memory": {"minimum": 128, "maximum": 1536}
        }
      }
    }
  ]
}`
