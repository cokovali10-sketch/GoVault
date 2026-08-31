# GoVault

Encrypted local storage for configs and secrets with AES-256-GCM encryption.

**Version:** 1.0.0 | **Language:** Go | **License:** MIT

## Features

### Security
- **AES-256-GCM encryption** — military-grade authenticated encryption
- **Password-derived key** — SHA-256 hashing of password
- **Random nonce** — unique encryption per save operation
- **File permissions** — vault files created with 0600 (owner-only)

### Storage Operations
- **Set** — store key-value pairs with optional tags
- **Get** — retrieve values by key
- **Delete** — remove entries
- **Rename** — rename keys without changing value
- **Exists** — check if key exists
- **Count** — total number of entries

### Organization
- **Tags** — categorize entries (e.g., "production", "staging")
- **Search by key** — find entries by key name
- **Search by tag** — find all entries with a specific tag
- **List** — sorted list of all keys

### Data Management
- **Export JSON** — unencrypted export for migration
- **Import JSON** — import from unencrypted JSON
- **Backup** — encrypted backup to separate file
- **Change password** — re-encrypt with new password (in-memory)

### Statistics
- Entry count
- Total size
- Unique tags
- Tag distribution

## Build & Run

```bash
go build -o govault .

# Store a secret
./govault -pass mypassword set db.password "s123" production

# Store with multiple tags
./govault -pass mypassword set api.key "abc123" api production critical

# Retrieve
./govault -pass mypassword get db.password

# List all entries
./govault -pass mypassword list

# Search by tag
./govault -pass mypassword search production

# Rename a key
./govault -pass mypassword rename old.key new.key

# Export to JSON (unencrypted)
./govault -pass mypassword export > backup.json

# Import from JSON
./govault -pass mypassword import backup.json

# Backup vault
./govault -pass mypassword backup vault-backup.db

# Stats
./govault -pass mypassword stats
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `set <key> <value> [tags...]` | Store a value |
| `get <key>` | Retrieve a value |
| `rm <key>` | Delete a key |
| `list` | List all keys |
| `search <query>` | Search by key or tag |
| `rename <old> <new>` | Rename a key |
| `export` | Export as JSON |
| `import <file>` | Import from JSON |
| `backup <path>` | Create encrypted backup |
| `stats` | Show vault statistics |

## Programmatic API

```go
package main

import (
    "fmt"
    "github.com/cokovali10-sketch/GoVault/vault"
)

func main() {
    v := vault.NewVault("secrets.db", "mypassword")
    
    // Load existing vault
    v.Load()
    
    // Store secrets
    v.Set("db.password", []byte("s3cret"), "database", "production")
    v.Set("api.key", []byte("abc123"), "api", "production")
    
    // Retrieve
    val, _ := v.Get("db.password")
    fmt.Println(string(val))
    
    // Search by tag
    keys := v.SearchByTag("production")
    fmt.Println("Production keys:", keys)
    
    // Stats
    fmt.Println(v.Stats())
    
    // Save
    v.Save()
}
```

## Test

```bash
go test -v ./vault/
```

## Project Structure

```
GoVault/
├── main.go              # CLI entry point
├── vault/
│   ├── vault.go         # Core vault with encryption
│   └── vault_test.go    # 14 unit tests
├── go.mod               # Go module definition
├── .gitignore
└── README.md
```

## Security Notes

- Always use strong passwords (12+ characters)
- Vault files are encrypted at rest
- No network communication — fully offline
- Backup encrypted vault files regularly

## License

MIT
