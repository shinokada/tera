# Terminal Radio (TERA)

## Overview

Play radio, CRUD favorites, search radio stations.

## Requirement

- [mpv](https://mpv.io/) is a free, open source, and cross-platform media player.
- [jq](https://stedolan.github.io/jq/) is a lightweight and flexible command-line JSON processor.
- [gh](https://cli.github.com/) is the GitHub CLI. Only for Gist functions.

## Features

- Powered by [Radio Browser API](https://de1.api.radio-browser.info/) and [MPV, a free, open source, and cross-platform media player](https://mpv.io/).
- 27780+ radio stations.
- Play from a list
- Search radio station by tag, name, language, country code, state.
- After play, you can save to a list.
- CRUD favorite lists.
- Delete a radio station from a list
- I feel lucky menu.
- Gist upload.

## Installation

Using [Awesome package manager](https://github.com/shinokada/awesome):

```sh
awesome install shinokada/tera
```

HomeBrew/LinuxBrew

```sh
brew tap shinokada/tera
brew install tera
```

## Usage

While playing music:

- Pause: `q` or `space`.
- Forward: Right arrow.
- Backward: Left arrow.
- [More MPV control](https://mpv.io/manual/master/)

### Commands

#### Main Menu

```sh
tera
```

![start](https://raw.githubusercontent.com/shinokada/tera/main/images/radio1.png)

#### Search Menu

You can search by tag, name, language, country code, state, and advanced(todo).

![start](https://raw.githubusercontent.com/shinokada/tera/main/images/searchmenu.png)

### Options

```sh
-h | --help
--version
```

## Reference

- [Bash menu](https://devdojo.com/bobbyiliev/how-to-create-an-interactive-menu-in-bash)

## Author

Shinichi Okada

## License

Please see LICENSE.
