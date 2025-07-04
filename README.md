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

## How to use

### Server instructions

<details> <summary> Run standalone binary </summary>

```bash
./backup-server -p 8008 -m 20 -f /backup/storage
```

Options:

- `-p`: Listening port (default: 80)
- `-m`: Max upload size in MB (default: 10)
- `-f`: Backup directory (default: ./backup)

</details>

<details> <summary> Run via docker </summary>

```bash
docker run -d \
  -p 8008:80 \
  -v ./local-backup:/backup \
  --name backup-sense \
  antun3s/backup-sense:latest
```
</details>

### Client instructions

<details> <summary> pfSense </summary>

I recommend Cron package to manage easily cron.

##### Install Cron package

On WebGUI:

- System > Package Manager > Available Packages
- Cron > Install

##### Cron configuration

On WebGUI:

- Services > Cron > Add
  - minute: `0`
  - hour: `1`
  - day of the month: `*`
  - month: `*`
  - day of the week: *`
  - user: `root`
  - command: `curl -X POST -F "file=@/cf/conf/config.xml" http://192.168.8.3:8081/upload` edit backup-sense server name and port
- Click on Apply

</details>

<details> <summary> OPNsense </summary>

##### Configura script

 On SSH: edit backup-sense server name and port on first line

```sh
# create script
printf '#\!/bin/sh\ncurl -X POST -F "file=@/conf/config.xml" http://192.168.8.3:8081/upload\n' > backup-sense.sh && chmod +x /root/backup-sense.sh
chmod +x /root/backup-sense.sh

# create custon action
printf '[run]\ncommand:/root/backup-sense.sh\nparameter:\ntype:script\nmessage:backup-sense\ndescription:backup-sense\n' > /usr/local/opnsense/service/conf/actions.d/actions_backup-sense.conf
service configd restart
```

##### Cron configuration

On WebGUI:

- System > Settings > Cron > Add (+)
  - minutes: `0`
  - hours: `1`
  - day of the week: `*`
  - command: backup-sense
  - description: backup-sense
- Click on Apply
</details>

## Why This Exists

Because sometimes you just need:

```bash
curl -F "file=@backup.xml" http://backup.example.com:8008/upload
```

...and nothing more.

> Simplicity is the ultimate sophistication.
