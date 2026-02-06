# TODO

Check if the code is still using the old config directory, `~/.config/tera`. If so, replace it with `os.UserConfigDir()`.

## Migration from v1 to v2

### Guide to migrate from v1 to v2

- Move all the files from `~/.config/tera` to OS config directory.
| Operating System | Config Directory |
| --- | --- |
| Linux|~/.config/tera/|
| macOS|~/Library/Application Support/tera/|
| Windows | %APPDATA%\tera\|

- Update the README.md by adding the migration guide.