#!/bin/bash

# Script para lanzar el servidor MCP Filesystem
# Permite acceso seguro a archivos del sistema

cd "$(dirname "$0")/.."

# Crear directorio para pruebas si no existe
mkdir -p test-workspace

# Lanzar el servidor MCP Filesystem con acceso al workspace actual
# Redirect echo output to stderr to avoid interfering with MCP protocol
echo "Starting Filesystem MCP Server..." >&2
echo "Allowed directory: $(pwd)" >&2

node mcp-servers/src/filesystem/dist/index.js "$(pwd)"