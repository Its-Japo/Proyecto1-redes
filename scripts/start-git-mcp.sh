#!/bin/bash

# Script para lanzar el servidor MCP Git
# Permite operaciones Git en el repositorio actual

cd "$(dirname "$0")/.."

# Agregar uv al PATH
export PATH="/home/Japo/.local/bin:$PATH"

# Lanzar el servidor MCP Git con acceso al repositorio actual
# Redirect echo output to stderr to avoid interfering with MCP protocol
echo "Starting Git MCP Server..." >&2
echo "Repository: $(pwd)" >&2

cd mcp-servers/src/git && uvx mcp-server-git --repository "$(pwd)/../../.."