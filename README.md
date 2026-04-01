# pm

pm reads its configuration and runs package manager tasks such as
syncing remote package definitions and updating installed packages.

## Usage:

```sh
pm [command]
```

## Available Commands:

- `completion`: Generate the autocompletion script for the specified shell
- `help`: Help about any command
- `sync`: Sync package managers
- `update`: Update package managers

## Flags:

- `--config`: (string) config file (default is $HOME/.pm.yaml)
- `-h`, `--help`: help for pm
- `-q`, `--quiet`: run without command output
- `-v`, `--verbose`: show full command output

Use `pm [command] --help` for more information about a command.

## Configuration

```yaml
managers:
  - manager: apt-get   # Arbitrary name
    depends: [apt-get] # If this binary does not exist in $PATH, then tasks are ignored
    tasks:
      - task: sync-remotes  # Must match available routines, usually sync-|update-|install-
        cmd: apt-get update # Command to run, do not use sudo or doas here
        privileged: true    # Default is false, runs as super user
        interactive: false  # Default is false, allows prompting user in the terminal
                            #   (for example for a sudo password or confirmation)
                            #   If privilege escalation is required but interactive is
                            #   set to false, then use pkexec or pinentry strategies.
      - task: update-remotes
        cmd: apt-get full-upgrade -y
        privileged: true
    remotes:
      - remote: trixie # Arbitrary name
        packages:      # Some example packages
          # media
          - package: mpv
          - package: vlc
          # tools
          - package: curl
          - package: groff
```
