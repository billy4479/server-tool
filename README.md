# Server Tool

A simple tool to run and maintain different Minecraft servers.

This tool supports Windows and Linux. MacOS is not supported (for now).

## How to use

Place the executable in the folder
where you want all your servers to be in.

When you want to create a new server
or run an existing one
just double click on the program and enjoy!

For a more advanced use see [Configuring](#configuring) below.

## Configuring

This tool uses a file called `server-tool.yml`
to store it's settings.

Depending on your operating system this fila can be found in different places:

- On Windows: `%AppData%\server-tool\server-tool.yml`
- On MacOS: `$HOME/Library/Application Support`
- On *nix platforms: `$XDG_CONFIG_HOME\server-tool\server-tool.yml`

Or it's position can be overridden by setting the `CONFIG_PATH` environment variable.

This is the default configuration (lines starting with `#` are comments):

```yml
# Application related settings
application:

  # Be more concise in the output
  quiet: false

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

  # Passes the `nogui` option to the server disabling the graphical interface
  nogui: true

  # Disable the automatic EULA agreement for new servers
  noeula: false

# Java related options
java:

  # Provide a Java executable yourself.
  #
  # By default the right version for each Minecraft version is downloaded
  # automatically from `adoptium.net`
  executableoverride: ""

  # Java memory
  memory:
    
    # How much?
    amount: 6

    # Gigabytes by default, megabytes otherwise
    gigabytes: true

  # JVM flags (advanced)
  flags:

    # Array of flags passed before `-jar`
    extraflags: []

    # Remove the default flags leaving only `extraflags`
    overridedefault: false

# Git related options
git:
  
  # Completely disable Git integration
  disable: false

  # Disable Github integration.
  #
  # This is used to create a new repository for new servers
  disablegithubintegration: false

  # Overrides for Git (advanced)
  overrides:
    
    # Overrides are disabled by default
    enable: false

    # Array of commands that are run before the server starts.
    #
    # Each command is an array that starts with the Git executable name
    # followed by the arguments.
    # You can then make a list of them, they will run in the order you specified.
    #
    # Example (single command): `["git", "pull", "origin", "master"]`
    # Tip: you may also specify a shell script
    #
    # Default:
    # - `git pull`
    customprecommands: []

    # Same syntax as above, but these commands are run after the server is done.
    #
    # If the server terminates with an error you will be asked if you want to run
    # them anyways.
    #
    # Default:
    # - `git add -A`
    # - `git commit --allow-empty-message -m ""`
    # - `git push`
    custompostcommands: []

# Automated start script related settings.
#
# Allow to specify a script to run to start the server instead of the default procedure.
#
# WARNING: this is an advanced (and unsafe) topic
startscript:

  # Disable this feature completely
  disable: false

  # The name of the script to search for.
  name: "start.sh"
```
