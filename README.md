# Redis-inspired in-memory key-value server

A lightweight Redis-inspired in-memory key-value server written in **Go**.

This project was built to understand how systems such as Redis work internally, particularly around TCP networking, request parsing, expiry, command dispatching, concurrent clients, and persistent storage.

## Features

- TCP server running on port `6380`
- RESP-style command parsing
- In-memory key-value storage
- Concurrent client connections using Go goroutines
- Command dispatcher for supported Redis commands
- Persistent key-value storage
- Expiry of fields
- Graceful shutdown using OS signals
- Automatic saving of data before shutdown

## Supported Commands

| Command | Description |
| --- | --- |
| `PING` | Checks whether the server is responding |
| `SET` | Stores a value under a key |
| `GET` | Retrieves a value associated with a key |
| `DEL` | Deletes a key-value pair |
| `EXPIRE` | Adds an expiry for data stored |

## Architecture

The server follows a simple request-processing pipeline:

```text
Client
  │
  │ TCP connection
  ▼
TCP Listener
  │
  ▼
Connection Goroutine
  │
  ▼
RESP Parser
  │
  ▼
Command Dispatcher
  │
  ▼
Key-Value Store
  │
  ▼
RESP Response
  │
  ▼
Client
```
