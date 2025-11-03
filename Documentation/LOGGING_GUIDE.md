# Logging Guide

This guide provides comprehensive information on using the structured logging system in the transaction-split-go project.

## Table of Contents

- [Overview](#overview)
- [Configuration](#configuration)
- [Basic Usage](#basic-usage)
- [Log Levels](#log-levels)
- [Structured Logging](#structured-logging)
- [HTTP Middleware](#http-middleware)
- [Best Practices](#best-practices)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

## Overview

The project uses Go's standard `log/slog` package for structured logging. The logger is configured via environment variables and supports multiple output destinations, log levels, and formats.

### Key Features

- ✅ **Environment-based configuration** - Control via `.env` file
- ✅ **Multiple log levels** - Debug, Info, Warn, Error
- ✅ **Multiple outputs** - Console, file, or both
- ✅ **Structured logging** - Key-value pairs for better parsing
- ✅ **HTTP middleware** - Automatic request/response logging
- ✅ **Format options** - JSON or text format
- ✅ **Source tracking** - File and line numbers for debug level

## Configuration

### Environment Variables

Configure the logger by setting these variables in your `.dev.env` file:

```env
# Log level: debug, info, warn, error
LOG_LEVEL=debug

# Output destination: stdout, file, both
LOG_OUTPUT=both

# Log format: text, json
LOG_FORMAT=text

# File path (used when LOG_OUTPUT is "file" or "both")
LOG_FILE_PATH=logs/app.log
```

### Configuration Options

#### LOG_LEVEL

Controls which messages are logged:

- `debug` - All messages (most verbose)
- `info` - Info, Warn, and Error messages
- `warn` - Warn and Error messages only
- `error` - Error messages only (least verbose)

**Recommendation**: Use `debug` in development, `info` in production.

#### LOG_OUTPUT

Controls where logs are written:

- `stdout` - Console output only
- `file` - File output only (creates file at LOG_FILE_PATH)
- `both` - Both console and file

**Recommendation**: Use `both` in development, `file` in production.

#### LOG_FORMAT

Controls log format:

- `text` - Human-readable format
- `json` - JSON format (better for log aggregation tools)

**Example text format**:
```
2025-10-20 14:30:00.000 INFO  [user.go:136] User created successfully name=John
```

The text format includes:
- Timestamp (YYYY-MM-DD HH:MM:SS.mmm)
- Log level (DEBUG/INFO/WARN/ERROR, padded to 5 chars)
- Source location (file:line) - only when LOG_LEVEL=debug
- Message
- Key-value attributes

**Example JSON format**:
```json
{"time":"2025-10-20T14:30:00.000Z","level":"INFO","msg":"User created successfully","user_id":123,"name":"John"}
```

## Basic Usage

### Package-Level Functions

The simplest way to log is using package-level functions:

```go
import "github.com/MattSharp0/transaction-split-go/internal/logger"

// Debug - detailed information for diagnosing problems
logger.Debug("Processing request", "user_id", userID)

// Info - general informational messages
logger.Info("User created successfully", "user_id", user.ID, "name", user.Name)

// Warn - warning messages for potentially harmful situations
logger.Warn("API rate limit approaching", "current", 950, "limit", 1000)

// Error - error messages that indicate a failure
logger.Error("Failed to create user", "error", err, "name", userName)
```

### Getting the Logger Instance

If you need more control, get the logger instance:

```go
import "github.com/MattSharp0/transaction-split-go/internal/logger"

log := logger.Get()
log.Info("Server starting", "port", 8080)
```

## Log Levels

### Debug

Use for detailed information useful during development and debugging:

```go
logger.Debug("Listing users",
    slog.Int("limit", int(limit)),
    slog.Int("offset", int(offset)),
)
```

**When to use**: 
- Function entry/exit
- Parameter values
- Query details
- Development debugging

### Info

Use for general informational messages about application flow:

```go
logger.Info("User created successfully",
    slog.Int64("user_id", user.ID),
    slog.String("name", user.Name),
)
```

**When to use**:
- Successful operations
- Application state changes
- Important business events
- Server start/stop

### Warn

Use for potentially harmful situations that aren't errors:

```go
logger.Warn("Updating individual split - this may leave transaction in invalid state",
    "split_id", id,
)
```

**When to use**:
- Deprecated API usage
- Invalid input (recoverable)
- Performance concerns
- Configuration issues

### Error

Use for error events that indicate a failure:

```go
logger.Error("Failed to create user",
    "error", err,
    "name", userName,
)
```

**When to use**:
- Database errors
- Network failures
- Failed operations
- Unexpected exceptions

## Structured Logging

### Key-Value Pairs

Always log in key-value pairs for better parsing:

```go
// ❌ Bad - string formatting
logger.Info(fmt.Sprintf("User %d created with name %s", id, name))

// ✅ Good - structured
logger.Info("User created successfully",
    "user_id", id,
    "name", name,
)
```

### Using slog Types

For type safety and consistency, use `slog` types:

```go
import "log/slog"

logger.Info("Transaction created",
    slog.Int64("transaction_id", txID),
    slog.String("name", txName),
    slog.Float64("amount", amount),
    slog.Time("date", txDate),
    slog.Duration("processing_time", duration),
    slog.Bool("is_recurring", isRecurring),
)
```

### Context Logging

Create a logger with common fields for a specific context:

```go
// Create a logger with user context
userLogger := logger.With(
    "user_id", userID,
    "session_id", sessionID,
)

// All logs will include user_id and session_id
userLogger.Info("Processing payment")
userLogger.Error("Payment failed", "error", err)
```

## HTTP Middleware

The HTTP middleware automatically logs all incoming requests and their responses.

### Automatic Logging

The middleware is automatically applied in `server.go`:

```go
// Middleware wraps all routes
Handler: logger.HTTPMiddleware(mux)
```

### Request Logging

For each request, the middleware logs at DEBUG level:

```
level=DEBUG msg="incoming request" method=POST path=/users remote_addr=127.0.0.1:12345 user_agent="curl/7.68.0"
```

### Response Logging

After processing, the middleware logs at appropriate levels:

**Successful requests (2xx):**
```
level=DEBUG msg="request completed" method=POST path=/users status=201 duration=15ms bytes=124
```

**Client errors (4xx):**
```
level=WARN msg="request completed" method=POST path=/users status=400 duration=5ms bytes=45
```

**Server errors (5xx):**
```
level=ERROR msg="request completed" method=POST path=/users status=500 duration=120ms bytes=0
```

### Error Responses

Failed requests are logged at appropriate levels:

- **2xx/3xx responses** → `DEBUG` level (successful requests)
- **4xx errors** (client errors) → `WARN` level
- **5xx errors** (server errors) → `ERROR` level

**Note:** Incoming requests are always logged at `DEBUG` level. The log level for completed requests depends on the status code. This means successful requests will only appear in logs when `LOG_LEVEL=debug`.

## Best Practices

### 1. Use Appropriate Log Levels

```go
// ✅ Debug - detailed diagnostics
logger.Debug("Query parameters", "limit", limit, "offset", offset)

// ✅ Info - business events
logger.Info("Payment processed", "amount", amount)

// ✅ Warn - recoverable issues
logger.Warn("Invalid format, using default", "input", input)

// ✅ Error - actual failures
logger.Error("Database connection failed", "error", err)
```

### 2. Include Context

Always include relevant context:

```go
// ❌ Bad - no context
logger.Error("Failed to create")

// ✅ Good - with context
logger.Error("Failed to create user",
    "error", err,
    "name", userName,
    "email", email,
)
```

### 3. Log Errors Once

Don't log the same error at multiple levels:

```go
// ❌ Bad - double logging
user, err := store.CreateUser(ctx, name)
if err != nil {
    logger.Error("CreateUser failed", "error", err)
    return fmt.Errorf("failed to create user: %w", err) // Also logged by caller
}

// ✅ Good - log once where handled
user, err := store.CreateUser(ctx, name)
if err != nil {
    logger.Error("Failed to create user", "error", err, "name", name)
    http.Error(w, "An error has occurred", http.StatusInternalServerError)
    return
}
```

### 4. Don't Log Sensitive Data

```go
// ❌ Bad - logging passwords
logger.Info("User login", "password", password)

// ✅ Good - omit sensitive data
logger.Info("User login attempt", "username", username)
```

### 5. Use Consistent Key Names

Standardize key names across the application:

```go
// ✅ Good - consistent naming
logger.Info("User created", "user_id", id)
logger.Info("User updated", "user_id", id)
logger.Error("User not found", "user_id", id)

// ❌ Bad - inconsistent naming
logger.Info("User created", "userID", id)
logger.Info("User updated", "user_id", id)
logger.Error("User not found", "id", id)
```

### 6. Log Start and End of Operations

```go
logger.Info("Starting backup process", "type", backupType)
// ... perform backup ...
logger.Info("Backup completed", "duration", time.Since(start), "size_mb", sizeMB)
```

## Examples

### Handler Example

```go
func createUser(s *server.Server, store db.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req models.CreateUserRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            logger.Warn("Invalid JSON in create user request", "error", err)
            http.Error(w, "Bad request", http.StatusBadRequest)
            return
        }

        if req.Name == "" {
            logger.Warn("Create user request missing name")
            http.Error(w, "Name is required", http.StatusBadRequest)
            return
        }

        logger.Info("Creating user", slog.String("name", req.Name))

        user, err := store.CreateUser(context.Background(), req.Name)
        if err != nil {
            logger.Error("Failed to create user",
                "error", err,
                "name", req.Name,
            )
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        logger.Info("User created successfully",
            slog.Int64("user_id", user.ID),
            slog.String("name", user.Name),
        )

        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(user)
    }
}
```

### Database Operation Example

```go
func processTransaction(ctx context.Context, store db.Store, tx models.Transaction) error {
    logger.Debug("Processing transaction",
        "transaction_id", tx.ID,
        "amount", tx.Amount,
        "user_id", tx.UserID,
    )

    result, err := store.CreateTransaction(ctx, tx)
    if err != nil {
        logger.Error("Failed to create transaction",
            "error", err,
            "transaction_id", tx.ID,
        )
        return fmt.Errorf("database error: %w", err)
    }

    logger.Info("Transaction processed successfully",
        slog.Int64("transaction_id", result.ID),
        slog.String("status", result.Status),
    )

    return nil
}
```

### Startup Example

```go
func main() {
    // Load environment and initialize logger
    logCfg := logger.LoadConfigFromEnv()
    log, err := logger.InitLogger(logCfg)
    if err != nil {
        slog.Error("Failed to initialize logger", "error", err)
        os.Exit(1)
    }

    log.Info("Application starting",
        slog.String("environment", os.Getenv("ENVIRONMENT")),
        slog.String("version", os.Getenv("VERSION")),
        slog.String("log_level", string(logCfg.Level)),
    )

    // ... rest of initialization ...

    log.Info("Application ready", "port", port)
}
```

## Troubleshooting

### Logs Not Appearing

**Problem**: No logs are showing up.

**Solutions**:
1. Check `LOG_LEVEL` - if set to `error`, only error messages appear
2. Verify logger is initialized before logging
3. Check log output destination (`LOG_OUTPUT`)

### File Logging Not Working

**Problem**: Logs don't appear in the log file.

**Solutions**:
1. Check `LOG_FILE_PATH` is writable
2. Verify `LOG_OUTPUT` is set to `file` or `both`
3. Check file permissions on the logs directory
4. Ensure logs directory exists (created automatically on startup)

### Too Many Logs

**Problem**: Log files are growing too large.

**Solutions**:
1. Increase `LOG_LEVEL` to `info` or `warn`
2. Review debug log placement - should only be for development
3. Implement log rotation (external tool like `logrotate`)

### Missing Context in Logs

**Problem**: Logs don't have enough information to debug issues.

**Solutions**:
1. Add more context fields to log statements
2. Use `logger.With()` to create contextual loggers
3. Review logged fields - include IDs, states, and relevant data

### Performance Impact

**Problem**: Logging is slowing down the application.

**Solutions**:
1. Use appropriate log levels - avoid excessive debug logging in production
2. Use `LOG_FORMAT=json` for better performance
3. Consider async logging for high-throughput scenarios
4. Review log frequency in tight loops

## Production Recommendations

### Configuration

```env
LOG_LEVEL=info
LOG_OUTPUT=file
LOG_FORMAT=json
LOG_FILE_PATH=/var/log/transaction-split/app.log
```

### Log Rotation

Use a log rotation tool:

```bash
# /etc/logrotate.d/transaction-split
/var/log/transaction-split/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 0644 appuser appuser
}
```

### Monitoring

Consider aggregating logs to a centralized system:

- **ELK Stack** (Elasticsearch, Logstash, Kibana)
- **Grafana Loki**
- **CloudWatch** (AWS)
- **Google Cloud Logging**

Use JSON format for easier parsing by log aggregation tools.

## Additional Resources

- [Go slog Documentation](https://pkg.go.dev/log/slog)
- [Structured Logging Best Practices](https://www.honeycomb.io/blog/structured-logging-and-your-team)
- [The Twelve-Factor App: Logs](https://12factor.net/logs)

## Summary

- Configure via `.env` file (`LOG_LEVEL`, `LOG_OUTPUT`, `LOG_FORMAT`)
- Use structured logging with key-value pairs
- Choose appropriate log levels (Debug/Info/Warn/Error)
- HTTP middleware automatically logs requests
- Include context and avoid sensitive data
- Use JSON format in production for log aggregation

