# Terminal Radio (TERA)

## Overview

Play radio, save/edit favorite, search radio stations.

## Requirement

- [mpv](https://mpv.io/) is a free, open source, and cross-platform media player.
- [jq](https://stedolan.github.io/jq/) is a lightweight and flexible command-line JSON processor.
- [gh](https://cli.github.com/) is the GitHub CLI

## Usage

### Commands

```sh
rob
```

![start](https://raw.githubusercontent.com/shinokada/rob/main/images/radio1.png)


```sh

# uses radio-browser.info bytag and grep
# use country codes
# after search result ask a number to play/save
rob search jazz 

# list a favorite and ask a number to play
rob ls

# open a favorite to edit with EDITOR
rob edit

# stop playing
rob stop

# pause playing
rob pause (or space)
```

### Options

```sh
-h | --help
--version
```

## Features


## Reference

- [Bash menu](https://devdojo.com/bobbyiliev/how-to-create-an-interactive-menu-in-bash)

## Author

Shinichi Okada

## License

Please see LICENSE.
