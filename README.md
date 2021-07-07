<h1  align="center">Terminal Radio (TERA)</h1>
<p align="center">
<img width="261" src="https://raw.githubusercontent.com/shinokada/tera/main/images/tera.png" />
</p>

## Overview

Tera is an interactive music radio player. Play your favorite radio station, CRUD your favorite lists, and explore new radio stations from your terminal.

## Requirement

- [mpv](https://mpv.io/) is a free, open source, and cross-platform media player.
- [jq](https://stedolan.github.io/jq/) is a lightweight and flexible command-line JSON processor.
- [gh](https://cli.github.com/) is the GitHub CLI.

## Features

- 27780+ radio stations powered by [Radio Browser API](https://de1.api.radio-browser.info/).
- [MPV, a free, open source, and cross-platform media player](https://mpv.io/).
- Play from a list
- Search radio station by tag, name, language, country code, state.
- Save a station to a list after playing.
- CRUD favorite lists.
- Delete a radio station from a list.
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

### Commands

#### Main Menu

```sh
tera
```

![start](https://raw.githubusercontent.com/shinokada/tera/main/images/radio1.png)

#### Search Menu

You can search by tag, name, language, country code, state, and advanced(todo).

![start](https://raw.githubusercontent.com/shinokada/tera/main/images/searchmenu.png)

#### Music player

- Pause: `q` or `space`.
- Forward: Right arrow.
- Backward: Left arrow.
- [More MPV control](https://mpv.io/manual/master/)

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
