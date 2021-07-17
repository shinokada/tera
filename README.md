<p align="center">
<img width="261" src="https://raw.githubusercontent.com/shinokada/tera/main/images/tera.png" />
</p>

<h1  align="center">Terminal Radio (TERA)</h1>

## Overview

Tera is an interactive music radio player. Play your favorite radio station, CRUD your favorite lists, and explore new radio stations from your terminal. 
Tera stores favorite list in the `~/.config/tera/favorite` directory and uses `~/.cache/tera` directory to keep search related results.

## Requirement

Unix-like environment.

- [mpv](https://mpv.io/) is a free, open source, and cross-platform media player.
- [jq](https://stedolan.github.io/jq/) is a lightweight and flexible command-line JSON processor.
- [fzf](https://github.com/junegunn/fzf) is a general-purpose command-line fuzzy finder.
- [gh](https://cli.github.com/) is the GitHub CLI.
- curl: Most UNIX-like OS should have it.

## Features

- 27780+ radio stations powered by [Radio Browser API](https://de1.api.radio-browser.info/).
- [MPV, a free, open source, and cross-platform media player](https://mpv.io/).
- CRUD favorite lists.
- Play from a list
- Search radio station by tag, name, language, country code, state.
- Save a station to a list after playing.
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

After installation please run the following to check `mpv` is 
installed correctly.

```sh
mpv https://live.musopen.org:8085/streamvbr0
```

If it plays music you're ready to go.

## Uninstallation

Remove followind directories.

- tera directory.
- `~/.config/tera/` directory
- `~/.cache/tera` directory

## Usage

### Commands

#### Main Menu

```sh
tera
```

![start](https://raw.githubusercontent.com/shinokada/tera/main/images/radio1.png)

#### Player control

| Keyboard    | Description                          |
| ----------- | ------------------------------------ |
| p and SPACE | Toggle pause/unpause.                |
| [ and ]     | Descrease/increase speed by 10%.     |
| { and }     | Halve/double current playback speed. |
| q           | Stop playing and quit.               |
| / and *     | Descrease/increase volume.           |
| 9 and 0     | Descrease/increase volume.           |
| m           | Mute sound.                          |



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
