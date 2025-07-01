
package docker

const DockerWorkSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "Docker Work Configuration",
  "description": "Configuration schema for Docker-based work execution",
  "required": ["image"],
  "properties": {
    "image": {
      "type": "string",
      "description": "Docker image to run",
      "examples": ["ubuntu:20.04", "python:3.9", "node:16-alpine"]
    },
    "command": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Command to execute in the container"
    },
    "args": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Arguments to pass to the command"
    },
    "environment": {
      "type": "object",
      "additionalProperties": {"type": "string"},
      "description": "Environment variables to set in the container"
    },
    "volumes": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["source", "target"],
        "properties": {
          "source": {"type": "string"},
          "target": {"type": "string"},
          "read_only": {"type": "boolean", "default": false},
          "type": {"type": "string", "enum": ["bind", "volume", "tmpfs"], "default": "bind"}
        }
      }
    },
    "resources": {
      "type": "object",
      "properties": {
        "cpu_limit": {"type": "string", "pattern": "^[0-9]+(\\.[0-9]+)?$"},
        "memory_limit": {"type": "string", "pattern": "^[0-9]+[kmgtKMGT]?[bB]?$"},
        "cpu_request": {"type": "string", "pattern": "^[0-9]+(\\.[0-9]+)?$"},
        "memory_request": {"type": "string", "pattern": "^[0-9]+[kmgtKMGT]?[bB]?$"}
      }
    },
    "network": {
      "type": "object",
      "properties": {
        "mode": {"type": "string", "enum": ["bridge", "host", "none", "container"]},
        "ports": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["host_port", "container_port"],
            "properties": {
              "host_port": {"type": "integer", "minimum": 1, "maximum": 65535},
              "container_port": {"type": "integer", "minimum": 1, "maximum": 65535},
              "protocol": {"type": "string", "enum": ["tcp", "udp"], "default": "tcp"}
            }
          }
        }
      }
    },
    "working_dir": {"type": "string"},
    "user": {"type": "string"},
    "privileged": {"type": "boolean", "default": false},
    "timeout": {"type": "string", "pattern": "^[0-9]+[smh]$"}
  }
}`
