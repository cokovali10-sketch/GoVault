# GoVault

Encrypted local storage for configs and secrets using AES-256-GCM.

## Features

- AES-256-GCM encryption (military grade)
- Password-derived key via SHA-256
- Set, get, delete, list, search operations
- Tag support for organizing entries
- Persistent encrypted file storage
- CLI and programmatic API

## Build & Run

```bash
go build -o govault .

# Store a secret
./govault -pass mypassword set db.password "s123" production

# Retrieve
./govault -pass mypassword get db.password

# List all
./govault -pass mypassword list

# Search by tag
./govault -pass mypassword search production

# Stats
./govault -pass mypassword stats
```

## Test

```bash
go test -v ./vault/
```

## License

MIT
