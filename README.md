# backup-sense

A simple, lightweight HTTP server to receive and store pfSense and OPNsense configuration backups.

> Created by a Brazilian professional who needed a simple way to manage multiple pfSense and OPNsense backups with ❤️ and the Unix philosophy in mind.

## Purpose

This tool provides a dead-simple way to receive firewall configuration backups via HTTP.  
Designed to be **minimal, focused, and reliable** - it does one thing well without unnecessary features.

## Key Features

- Accepts XML backups from pfSense/OPNSense firewalls
- Auto-organizes files by hostname (`hostname/YYYYMMDD-HHMMSS-hostname.xml`)
- Configurable backup location (default: `./backup`)
- Built-in size limit protection
- Client IP logging

## Philosophy

- **Unix-like**: Single-purpose tool that composes well with others
- **No S3/Cloud**: For cloud backups, use dedicated tools like `rclone` or `awscli`
- **No Auth**: Run behind a reverse proxy if security is needed
- **No DB**: Plain filesystem storage for simplicity

## Usage

### Server Options

#### Run Standalone Binary

```bash
./backup-server -p 8008 -m 20 -f /backup/storage
```

Options:

- `-p`: Listening port (default: 80)
- `-m`: Max upload size in MB (default: 10)
- `-f`: Backup directory (default: ./backup)

#### Run via Docker

```bash
docker run -d \
  -p 8008:80 \
  -v ./local-backup:/backup \
  --name backup-sense \
  antun3s/backup-sense:latest
```

### Client Instructions

#### pfSense

```bash
curl -X POST -F "file=@/cf/conf/config.xml" http://yourserver:8008/upload
```

#### OPNsense

```bash
curl -X POST -F "file=@/conf/config.xml" http://yourserver:8008/upload
```

## Why This Exists

Because sometimes you just need:

```bash
curl -F "file=@backup.xml" http://backup.example.com:8008/upload
```

...and nothing more.

> Simplicity is the ultimate sophistication.
