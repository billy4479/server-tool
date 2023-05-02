# Server Tool

<u>[**Versione italiana**](README.it.md)</u>

A simple tool to run and maintain different Minecraft servers.

This tool supports Windows and Linux. MacOS is not supported (for now).

## How to use

Place the executable in the folder
where you want all your servers to be in.

When you want to create a new server
or run an existing one
just double click on the program and enjoy!

For a more advanced use see [Configuring](#configuring) below.

## TUI / CLI mode (advanced)

The program runs by default with a GUI but it can be configured to run in TUI or CLI mode too.

Documentation for the CLI version is not yet available.

### TUI Demo

[![asciicast](https://asciinema.org/a/459894.svg)](https://asciinema.org/a/459894)

<center>
<sub>Demo of the TUI version</sub>
</center>

## Where to download

You can get a binary from the [Release page](https://github.com/billy4479/server-tool/releases).
Download the one for your operating system.

## Configuring

This tool uses a file called `server-tool.yml` to store it's settings.

Depending on your operating system this file can be found in different places:

- On Windows: `%AppData%\server-tool\server-tool.yml`
- On MacOS: `$HOME/Library/Application Support`
- On \*nix platforms: `$XDG_CONFIG_HOME\server-tool\server-tool.yml`

Or it's position can be overridden by setting the `CONFIG_PATH` environment variable.

This is the default configuration (lines starting with `#` are comments):

```yml
# Application related settings
application:
  # Enables automatic updates
  autoupdate: true

  # The folder where the server are located relative to the execution directory
  workingdir: "."

  # The folder where Java and version manifests will be downloaded.
  #
  # By default this is set to
  # - On Windows:   `%LocalAppData%\server-tool`
  # - On MacOS:     `$HOME/Library/Caches/server-tool`
  # - On *nix:      `$XDG_CACHE_HOME/server-tool`
  cachedir: ""

# These options are related to the Minecraft server executable
minecraft:
  # Do not print server logs.
  #
  # Stdin is still forwarded so you can still type commands.
  # Logs can be still found as usual in the logs folder of each server
  quiet: false

  # Passes the `nogui` option to the server disabling the graphical interface.
  # This is only used in TUI and CLI mode, in GUI mode this is always
  # considered to be true.
  gui: true

  # Disable the automatic EULA agreement for new servers
  noeula: false

  # Amount of memory to give to the java process in megabytes
  memory: 6144

# Git related options
git:
  # Enable Git integration
  enable: true

  # Creates a lock file that is immediately committed to the repo
  #
  # When the server is started this program checks for the presence of a lock file and immediately aborts if it finds one
  # Note that if config overrides are active this option will be ignored
  uselockfile: true
```
