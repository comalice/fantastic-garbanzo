
package grpc

const GRPCWorkSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "gRPC Work Configuration",
  "description": "Configuration schema for gRPC-based work execution",
  "required": ["endpoint", "method"],
  "properties": {
    "endpoint": {
      "type": "string",
      "description": "gRPC server endpoint",
      "examples": ["localhost:9090", "api.example.com:443"]
    },
    "method": {
      "type": "string",
      "description": "gRPC method to call",
      "examples": ["myservice.MyService/MyMethod"]
    },
    "headers": {
      "type": "object",
      "additionalProperties": {"type": "string"},
      "description": "HTTP headers to include in the request"
    },
    "timeout": {
      "type": "string",
      "pattern": "^[0-9]+[smh]$",
      "description": "Request timeout",
      "default": "30s"
    },
    "retry": {
      "type": "object",
      "properties": {
        "max_attempts": {"type": "integer", "minimum": 1, "default": 3},
        "initial_delay": {"type": "string", "pattern": "^[0-9]+[smh]$", "default": "1s"},
        "max_delay": {"type": "string", "pattern": "^[0-9]+[smh]$", "default": "10s"},
        "multiplier": {"type": "number", "minimum": 1.0, "default": 2.0},
        "jitter": {"type": "boolean", "default": true}
      }
    },
    "tls": {
      "type": "object",
      "properties": {
        "enabled": {"type": "boolean", "default": false},
        "insecure_skip_verify": {"type": "boolean", "default": false},
        "cert_file": {"type": "string"},
        "key_file": {"type": "string"},
        "ca_file": {"type": "string"},
        "server_name": {"type": "string"}
      }
    },
    "compression": {
      "type": "string",
      "enum": ["gzip", "deflate"],
      "description": "Compression algorithm to use"
    },
    "metadata": {
      "type": "object",
      "additionalProperties": {"type": "string"},
      "description": "gRPC metadata to include in the request"
    }
  }
}`
