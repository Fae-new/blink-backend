# Local Agent API for Frontend

This document explains how the frontend application should interact with the local agent to enable localhost API testing.

## Overview

The Local Agent runs on the user's machine at `http://localhost:5555`.

### Flow:
1. Frontend checks if Agent is running (`GET /health`)
2. If user requests a `localhost` or private IP URL:
   - If Agent is connected: Send request to Agent
   - If Agent is NOT connected: Show error/prompt to download Agent

## 1. Agent Detection

Check if the agent is running when the application loads.

**Request:**
```http
GET http://localhost:5555/health
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "message": "Agent is running"
}
```

**Frontend Implementation Example:**
```javascript
async function checkAgentStatus() {
  try {
    const response = await fetch('http://localhost:5555/health');
    if (response.ok) {
      console.log('Local Agent detected!');
      return true;
    }
  } catch (e) {
    console.log('Local Agent not running');
    return false;
  }
}
```

## 2. Execute Request via Agent

Send a request to be executed by the agent locally.

**Request:**
```http
POST http://localhost:5555/execute
Content-Type: application/json
```

**Body:**
```json
{
  "method": "POST",
  "url": "http://localhost:3000/api/users",
  "headers": {
    "Content-Type": "application/json",
    "Authorization": "Bearer token123"
  },
  "body": "{\"name\": \"John Doe\"}"
}
```

**Response (200 OK):**
```json
{
  "status": 201,
  "statusText": "Created",
  "headers": {
    "Content-Type": ["application/json"],
    "Date": ["Mon, 16 Feb 2026 00:00:00 GMT"]
  },
  "body": "{\"id\": 1, \"name\": \"John Doe\"}",
  "timing": {
    "startTime": "2026-02-16T00:00:00Z",
    "endTime": "2026-02-16T00:00:00.123Z",
    "duration": 123
  },
  "error": "" 
}
```

**Error Response:**
If the agent fails to execute the request (e.g., DNS error, connection refused):
```json
{
  "status": 0,
  "error": "dial tcp 127.0.0.1:3000: connect: connection refused"
}
```

## 3. Downloads

The frontend should provide links to download the agent binaries served by the backend:

- **macOS (Intel):** `/downloads/blink-agent-darwin-amd64`
- **macOS (Apple Silicon):** `/downloads/blink-agent-darwin-arm64`
- **Windows:** `/downloads/blink-agent-windows-amd64.exe`
- **Linux:** `/downloads/blink-agent-linux-amd64`

### Auto-Start Installation

Users can install the agent to run automatically on login:

#### macOS & Linux
1.  **Open Terminal**
2.  **Make Executable:** `chmod +x blink-agent-darwin-arm64` (replace with your file name)
3.  **Install:** `./blink-agent-darwin-arm64 --install`
4.  **Result:** Agent runs in background and starts on login.

### Uninstalling

To remove the auto-start configuration:

**macOS/Linux:**
```bash
./blink-agent-darwin-arm64 --uninstall
```

**Windows:**
```powershell
.\blink-agent-windows-amd64.exe --uninstall
```

#### Windows
1.  **Open PowerShell**
2.  **Install:** `.\blink-agent-windows-amd64.exe --install`
3.  **Result:** Agent runs in background and starts on login.
