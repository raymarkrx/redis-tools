# redis-tools

Redis dump and redis restore tools

## Usage

### redisdump

```bash
Export the content of a running server into .json files.

Usage:
  redisdump [flags]

  Flags:
    -h, --help              help for redisdump
        --host string       server hostname (default "localhost")
        --out string        output file (default "redis.json")
        --password string   password to use when connecting to the server
        --port int          server port (default 6379)
    -v, --version           print version information
```

### redisrestore

```bash
Restore backups generated with redisdump to a running server.

Usage:
  redisrestore <FILENAME> [flags]

Flags:
  -h, --help              help for redisrestore
      --host string       server hostname (default "localhost")
      --password string   password to use when connecting to the server
      --port int          server port (default 6379)
  -v, --version           print version information
```
