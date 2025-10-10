# quadlet-helper (qh)

Things that should probably be part of a script. Involves a good bit of AI-generated code and meant primarily for personal use with [quad-bucket](https://github.com/mufeedali/quad-bucket).

Part of the experiment here is to explore Go as a functional, object-oriented and AOT-compiled alternative to scripting. A largely successful experiment so far.

## Features

- **Unit Management**: Control quadlet unit files (start, stop, enable, disable, logs, status). Mostly here because I want completions.
- **Cloudflare integration**: Automated Cloudflare IP updater.
- **Example files generation**: Generate example configurations (currently only traefik) and environment files
- **Backup Management**: Create and manage automated backup services using rsync, restic, or rclone. Unnecessarily elaborate, including email notifications. Should have still been just a script.

## Installation

You probably shouldn't be installing this. If this is somehow exactly what you want, you're probably doing something wrong...

But anyway...

```bash
go install github.com/mufeedali/quadlet-helper@latest
```

Or build from source:

```bash
git clone https://github.com/mufeedali/quadlet-helper.git
cd quadlet-helper
go build -o qh
```

## Usage

```bash
# Backup commands
qh backup create <name>      # Create a new backup configuration
qh backup install <name>     # Install backup service and timer
qh backup list               # List all backup configurations
qh backup run <name>         # Run backup immediately
qh backup status <name>      # Check backup status
qh backup logs <name>        # View backup logs

# Unit commands
qh unit list                 # List quadlet units
qh unit start <name>         # Start a unit
qh unit stop <name>          # Stop a unit
qh unit status <name>        # Check unit status
qh unit logs <name>          # View unit logs
qh unit validate <file>      # Validate quadlet file

# Cloudflare commands
qh cloudflare install        # Install Cloudflare IP updater
qh cloudflare run            # Update Cloudflare DNS
qh cloudflare uninstall      # Remove Cloudflare service

# Generate commands
qh generate env       # Generate .env examples
qh generate traefik   # Generate Traefik example
```

## Configuration

By default, quadlet-helper looks for container configurations in:
```
~/.config/containers/systemd
```

Override with the `--containers-path` flag:
```bash
qh --containers-path /custom/path unit list
```

## License

MIT. See [LICENSE](LICENSE).
