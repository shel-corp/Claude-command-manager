# Overview

The command library is a small program that allows users to opt-in to custom claude commands from a library of commands. It is designed to be simple and easy to use, with a focus on providing a seamless user experience. Command library is a CLI application.

# User experience

- The journey starts with running the /dscout/bin/command_library command. 
- The user is presented with a list of available commands, each with a brief description in the following format:
    -  `[X]` `command_name`: A brief description of what the command does.
    - There are three components to the command:
      - The [X] indicates whether the command is enabled or disabled.
        - The `command_name` is the name of the command that can be used in the chat. This is the same as the command file name.
        - The `description` is a brief description of what the command does. This is parsed from the command files yaml header.
- The user can enable or disable commands by highlighting the command and pressing enter.
- The user can also rename commands by highlighting the command and pressing the `r` key.
- The user can disable commands by highlighting the command and pressing the `d` key.
- The user can enable,disable, or rename as many commands as they like in one session.
- The user can exit without saving by pressing the `q` key.
- The user can save their changes by pressing the `s` key.


# Command file format

- Command files are stored in the `/dscout/.claude/command_library/commands` directory.
- Each command is a Markdown file with a `.md` extension. You should reference the `~/.claude/commands/slash_command.md` file for an example.
- On save, the command library will 
    - On add, sym_link the command files to the `~/.claude/commands` directory. Commands that should be renamed will rename the symlink in the commands directory. Update the config file to reflect the new name and enabled status.
    - On remove, delete the symlink from the `~/.claude/commands` directory and update the config to remove the command from the config file.
- You will need to keep track of which commands are enabled and disabled in a separate file, such as `~/.claude/commands_enabled.json`. This file will be a JSON object with the command names as keys and a boolean value indicating whether the command is enabled or disabled, and what the renamed file is, if it exists.
    - This file will live in the `~/.claude/command_library/.config` file.

# File architecture

- The command library script will live in the `/dscout/.claude/command_library` directory.
- The command library will be a CLI application that is run from the `/dscout/bin/command_library` directory.
    - This simply executes the command library script located in `/dscout/.claude/command_library/command_library.sh`.
- The command library will have a `commands` directory where the command files are stored.
- The command library will have a `.config.json` file where the configuration is stored.

