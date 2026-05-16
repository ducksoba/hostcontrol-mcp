# hostcontrol-mcp

Expose host management tools via MCP (Model Context Protocol). Run it on any Linux machine to give AI assistants access to read files, execute commands, search, inspect processes, and manage the host.

## Tools

| Tool       | Description                                               |
| ---------- | --------------------------------------------------------- |
| `read`     | Read file contents with optional line range               |
| `write`    | Write or append to a file                                 |
| `bash`     | Execute shell commands with timeout and working directory |
| `grep`     | Search files using regex patterns, with recursive support |
| `find`     | Find files by name, type, or depth                        |
| `ls`       | List directory contents, with optional long format        |
| `ps`       | List running processes, filterable by user or command     |
| `kill`     | Send signals to processes by PID                          |
| `hostname` | Get the host's hostname                                   |

## Features

- **Stdio transport** — connects via stdin/stdout, works with any MCP client
- **HTTP transport** — streamable HTTP mode for remote access
- **Access control** — path restrictions, command allowlists, timeout limits, process ownership checks
- **Process management** — list and signal processes via `/proc`
- **File operations** — read with line ranges, write with append mode
- **Search** — grep with regex across files and directories

## Build

```bash
make build
```

Binary is output to `dist/hostcontrol-mcp`.

Other targets:

```bash
make clean   # remove dist/
make tidy    # run go mod tidy
make test    # run tests
```

## Run

### Stdio mode (default)

For Claude Desktop / Claude Code, configure your MCP client:

```json
{
  "mcpServers": {
    "hostcontrol": {
      "command": "/path/to/hostcontrol-mcp/dist/hostcontrol-mcp"
    }
  }
}
```

### HTTP mode

Run with HTTP transport:

```bash
./dist/hostcontrol-mcp -transport http -listen 127.0.0.1:3000
```

The MCP endpoint is available at `http://127.0.0.1:3000/mcp`.

Configure your MCP client to connect via HTTP/SSE:

```json
{
  "mcpServers": {
    "hostcontrol": {
      "url": "http://127.0.0.1:3000/mcp"
    }
  }
}
```

### Via SSH tunnel

Run on a remote host and connect via SSH:

```bash
ssh -L 3000:127.0.0.1:3000 user@remote-host
```

Then point your MCP client to `http://127.0.0.1:3000/mcp`.

## Flags

| Flag         | Default          | Description                        |
| ------------ | ---------------- | ---------------------------------- |
| `-transport` | `stdio`          | Transport mode: `stdio` or `http`  |
| `-listen`    | `127.0.0.1:3000` | Listen address for HTTP mode       |
| `-config`    | —                | Path to access control config file |

## Access Control

Load a config file with `-config` to restrict tool access:

```bash
./dist/hostcontrol-mcp -config config.json
```

### Config Format

```json
{
  "allowed_paths": ["/home/user", "/tmp", "/var/log/app"],
  "denied_paths": ["/etc/shadow", "/root"],
  "allow_bash": true,
  "bash_allow_re": ["^ls", "^cat", "^grep", "^ps", "^df", "^uptime"],
  "bash_deny_re": ["^rm", "^sudo", "^chmod", "^chown", "^mkfs", "^dd"],
  "bash_strict": true,
  "max_bash_timeout": 60,
  "kill_restrict_to_owner": true
}
```

### Fields

| Field                    | Applies to                            | Behavior                                                |
| ------------------------ | ------------------------------------- | ------------------------------------------------------- |
| `allowed_paths`          | `read`, `write`, `grep`, `ls`, `find` | Path must start with one of these prefixes              |
| `denied_paths`           | `read`, `write`, `grep`, `ls`, `find` | Checked first — always blocks matching paths            |
| `allow_bash`             | `bash`                                | Must be true to allow bash execution                    |
| `bash_allow_re`          | `bash`                                | Command must match at least one regex                   |
| `bash_deny_re`           | `bash`                                | Checked first — always blocks matching commands         |
| `bash_strict`            | `bash`                                | Blocks `;`, `&&`, `\|\|`, `\|`, `$()`, backticks, and newlines |
| `max_bash_timeout`       | `bash`                                | Caps the timeout parameter (seconds)                    |
| `kill_restrict_to_owner` | `kill`                                | Only allows killing processes owned by the running user |

### Evaluation Order

**Path tools** (`read`, `write`, `grep`, `ls`, `find`):

1. Check `denied_paths` — if matched, block
2. Check `allowed_paths` — if set and no match, block
3. Otherwise, allow

**Bash**:

1. If `allow_bash` is false, block
2. If `bash_strict` is true, block chaining (`;`, `&&`, `||`, `|`), substitution (`$()`, backticks), and newlines
3. Check `bash_deny_re` — if matched, block
4. Check `bash_allow_re` — if set and no match, block
5. Apply `max_bash_timeout` cap
6. Otherwise, allow

**Kill**:

1. If `kill_restrict_to_owner` is true, verify process UID matches running user

### No Config = No Restrictions

If no config file is provided, all tools operate without restrictions (backward compatible).

## Access Control via Encik

For more advanced policy enforcement, use [Encik](https://github.com/encik/encik) as an MCP gateway in front of hostcontrol-mcp:

```
┌─────────────┐     ┌──────────┐     ┌──────────────┐
│ MCP Client  │────▶│  Encik   │────▶│ hostcontrol  │
│             │     │ + Policy │     │ Read/Write/  │
└─────────────┘     │          │     │ Bash/Grep/   │
                    │          │     │ Ls/Ps/Kill   │
                    └──────────┘     └──────────────┘
```
