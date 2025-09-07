# Proyecto de Análisis de Acciones con MCP

Una implementación integral del Protocolo de Contexto de Modelo (MCP) en Go, que incluye un sistema de análisis del mercado de valores con integración de datos en tiempo real e interfaz de chatbot potenciada por IA.

## Resumen del Proyecto

Este proyecto demuestra la implementación de MCP (Protocolo de Contexto de Modelo) para un curso universitario de redes (CC3067 Redes - UVG). Consiste en:

- **Host de Chatbot MCP**: Chatbot interactivo con integración de Claude AI
- **Servidor MCP de Análisis de Acciones**: Servidor local que proporciona herramientas de análisis de acciones
- **Integración de API Financiera**: Datos de mercado en tiempo real de Alpha Vantage
- **Motor de Análisis Técnico**: RSI, promedios móviles, MACD, Bandas de Bollinger
- **Recomendaciones de Inversión**: Recomendaciones de comprar/vender/mantener impulsadas por IA

## Arquitectura

```
┌─────────────────┐    JSON-RPC     ┌──────────────────┐    HTTP API    ┌─────────────────┐
│   Host Chatbot  │ ◄────────────► │ Analizador de    │ ◄───────────► │ Alpha Vantage   │
│   (Claude AI)   │                │   Acciones MCP   │                │  API Financiera │
└─────────────────┘                └──────────────────┘                └─────────────────┘
```

## Inicio Rápido

### Prerrequisitos

```bash
# Instalar Go 1.21+
go version

# Obtener claves API (gratis)
# 1. Alpha Vantage: https://www.alphavantage.co/support/#api-key
# 2. Anthropic Claude: https://console.anthropic.com/
```

### Configuración

```bash
# Configurar variables de entorno
export ALPHA_VANTAGE_API_KEY="tu_clave_alpha_vantage"
export ANTHROPIC_API_KEY="tu_clave_anthropic"

# Instalar dependencias
go mod download

# Construir ejecutables
./setup.sh

# Ejecutar el chatbot (se conecta automáticamente al servidor MCP)
./bin/chatbot

# O ejecutar sin conexión automática para control manual
./bin/chatbot -no-auto-connect
```

## Ejemplos de Uso

### Comandos Interactivos

```bash
Chatbot de Análisis de Acciones MCP
===================================

Tú: /analyze AAPL,GOOGL,MSFT
Analizando portafolio: [AAPL GOOGL MSFT]
Análisis completo:

REPORTE DE ANÁLISIS DE PORTAFOLIO
=================================
Portafolio: Portafolio de Análisis
Puntuación General: 72.3/100
Riesgo General: MEDIO

AAPL
  Precio: $185.64 (-1.23%)
  Recomendación: COMPRAR (Puntuación: 75.0/100)
  Nivel de Riesgo: BAJO
```

### Consultas en Lenguaje Natural

```bash
Tú: ¿Debería invertir en acciones de Tesla?
Símbolos de acciones detectados: [TSLA]
ANÁLISIS DE ACCIONES: TSLA
==========================
Precio Actual: $248.42
Recomendación: MANTENER (Puntuación: 58.0/100)
Nivel de Riesgo: ALTO
```

## Herramientas MCP Disponibles

| Herramienta | Descripción | Parámetros |
|-------------|-------------|------------|
| `analyze_portfolio` | Analizar múltiples acciones con recomendaciones | `symbols[]`, `timeframe` |
| `get_stock_price` | Obtener precio actual y análisis técnico | `symbol` |
| `export_analysis` | Exportar resultados a CSV/JSON | `format`, `filename` |

### Comandos de Gestión de Conexión

| Comando | Descripción |
|---------|-------------|
| `/status` | Mostrar estado de conexión y verificación de salud |
| `/connect <ruta>` | Conectar manualmente al servidor MCP |
| `/disconnect <nombre>` | Desconectar del servidor MCP |
| `/list` | Listar herramientas disponibles de servidores conectados |

## Características Técnicas

### Análisis Financiero
- **Indicadores Técnicos**: RSI, SMA, EMA, MACD, Bandas de Bollinger
- **Evaluación de Riesgo**: Análisis de volatilidad y puntuación de riesgo
- **Motor de Recomendaciones**: Sistema de puntuación multifactor
- **Análisis de Portafolio**: Análisis de diversificación

### Implementación MCP
- **JSON-RPC 2.0 Puro**: Sin dependencias externas del SDK MCP
- **Protocolo de Transmisión**: Comunicación bidireccional en tiempo real
- **Manejo de Errores**: Respuestas de error integrales
- **Descubrimiento de Herramientas**: Registro y listado dinámico de herramientas

## Desarrollo

### Estructura del Proyecto

```
proyecto-mcp-bolsa/
├── cmd/chatbot/           # Aplicación host del chatbot
├── internal/
│   ├── mcp/              # Implementación del protocolo MCP
│   ├── stock/            # Lógica de análisis de acciones
│   └── llm/              # Cliente de Claude AI
├── pkg/models/           # Estructuras de datos
├── servers/stock-analyzer/  # Implementación del servidor MCP
├── examples/scenarios/   # Ejemplos de uso
└── config.yaml          # Configuración
```

### Compilación

```bash
# Crear directorio bin
mkdir -p bin

# Compilar chatbot
go build -o bin/chatbot ./cmd/chatbot/

# Compilar servidor analizador de acciones
go build -o bin/stock-analyzer ./servers/stock-analyzer/

# Ejecutar pruebas
go test ./...
```

## Análisis de Red

### Capas de Protocolo (Modelo OSI)

1. **Capa de Aplicación (7)**: Protocolo MCP, JSON-RPC 2.0
2. **Capa de Presentación (6)**: Serialización JSON, codificación UTF-8
3. **Capa de Sesión (5)**: Sesiones HTTP/HTTPS
4. **Capa de Transporte (4)**: TCP para comunicación confiable
5. **Capa de Red (3)**: Enrutamiento IP para llamadas API
6. **Capa de Enlace de Datos (2)**: Tramas Ethernet
7. **Capa Física (1)**: Hardware de red

## Cumplimiento del Protocolo MCP

Esta implementación sigue la especificación oficial de MCP:
- Versión del protocolo: 2024-11-05
- Transporte JSON-RPC 2.0
- Flujo de inicialización estándar
- Descubrimiento y ejecución de herramientas
- Convenciones de manejo de errores

## Referencias

- [Especificación MCP](https://modelcontextprotocol.io/specification/2025-06-18)
- [Especificación JSON-RPC 2.0](https://www.jsonrpc.org/specification)
- [Documentación API Alpha Vantage](https://www.alphavantage.co/documentation/)
- [API Anthropic Claude](https://docs.anthropic.com/)

---

**Curso**: CC3067 Redes - Universidad del Valle de Guatemala