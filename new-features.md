# New Features

## Shuffle mode

- Add a shuffle mode to I feel lucky page.
(- Add a shuffle mode to main page Quick play Favorites.)

## Auto connect
- #4: Using GPRS/4G, I sometimes lose the signal/connection and then have to reconnect manually to the station I was listening to.

### Possible solutions (I'm not sure if any of these works)
1. The "Force Loop" Method
The simplest way to keep mpv from quitting when a stream drops is to use the loop-playlist flag with the force parameter. 

- Command: mpv --loop-playlist=force <URL>
- Why it works: Unlike standard looping, force tells mpv not to skip entries that have failed. If the connection drops and the "file" (stream) ends, mpv will immediately try to start it again. 

2. FFmpeg Reconnect Flags
Since mpv uses FFmpeg for networking, you can pass specific reconnection instructions directly to the underlying stream layer. 
- Command: mpv --stream-lavf-o=reconnect_streamed=1,reconnect_delay_max=5 <URL>
- Why it works: This tells the FFmpeg backend to attempt to reconnect if the TCP/HTTP connection is severed. 

3. Recommended "Driver" Configuration
For the best experience while driving, combine these options to maximize stability and minimize manual interaction:
- --loop-playlist=force: Keeps the player open after a drop.
- --cache=yes: Increases the buffer to handle minor signal dips before the audio actually stops.
- --demuxer-max-bytes=50M: Sets a larger cache size (e.g., 50MB) to bridge longer "dead zones" in 4G coverage. 


## Completed

### (Done v1.4.0) Search updates (added to advanced search)
#### Bitrate/Speed
Select from pre-defined bitrate or input field for a specific bitrate.

#### Search by Popularity/Votes 

### (Done v1.4.0) Advanced Search update 
Combined: Use multiple fields to pinpoint exact vibes

```text
Advanced Search

Tag (optional):        classical
Language (optional):   italian
Country / Country code (optional):
State (optional):
Name contains (optional):

Sort by: votes (default) | relevance:

Bitrate (optional):
1) Low   (≤ 64 kbps)
2) Medium (96–128 kbps)
3) High  (≥ 192 kbps)
```

The definition of “relevance” internally as: Radio Browser default ordering.

[From Radio Browser API docs](https://github.com/ivandotv/radio-browser-api/blob/master/docs/api/README.md)


### (Fixed v1.4.0) Problems
- Country name seems to be required in capital for the first letter
- Can't unselect Bitrate. Once selecting 1/2/3, not able to unselect.
- Country Code is not working. If an input is two letters, it should use country code query. Or should I separate Country and Country Code?
- Language seems to be required in lower case, so code should convert to lower case.
- All other playing station pages use a space keyboard shortcut to pause. In the playing station page after advanced search, the space key should pause the station.
- Add/update README.md for advanced research section.

### (Done v1.4.0) Search page (Added to Advanced Search)
- #42: color code for stream speed (e.g., 64 kbps, 128 kbps, 320 kbps)
- Sort by stream speed

### (Done v1.4.0) Space keyboard shortcut to pause

### (Done v1.3.0) Settings Smart Update commands
Currently Settings > 2. Check for Updates page shows the following:

```
                                               
                         TERA

  🔄 Check for Updates

  Current version: 1.1.

  ─────────────────────────────────────────

  ⬆ New version available!
  Latest version: v1.2.0
  Release page:
    https://github.com/shinokada/tera/releases/latest
  ─────────────────────────────────────────
  Update instructions:
    If installed via Go:
      go install github.com/shinokada/tera/cmd/tera@latest
 
    Or visit the release page to download binaries. 
```

In the above page, if an update is available, add a list item to select updating the version using the following method.

Can Tera run the command when it is selected? Or should we just show the command?

Detect installation method and show appropriate command:

| Method         | Detection                           | Update Command                                       |
| -------------- | ----------------------------------- | ---------------------------------------------------- |
| Homebrew       | Check if brew list tera succeeds    | brew upgrade tera                                    |
| Go install     | Check go env GOPATH contains binary | go install github.com/shinokada/tera/cmd/tera@latest |
| Scoop          | Check if scoop list tera succeeds   | scoop update tera                                    |
| APT/DEB        | Check /var/lib/dpkg/info/tera.list  | sudo apt update && sudo apt upgrade tera             |
| RPM            | Check rpm -q tera                   | sudo dnf upgrade tera                                |
| Manual/Unknown | Fallback                            | Link to releases page                                |


New screen:


```
                                               
                         TERA

  🔄 Check for Updates

  Current version: 1.1.

  ─────────────────────────────────────────

  ⬆ New version available!
  Latest version: v1.2.0

  UPDATE using xxx (this can be brew/go/sudo apt update/...)
  (If Manual/Unknown show the following)
  Release page:
    https://github.com/shinokada/tera/releases/latest
```

### (Done: v1.2.0) Search page enhancements

- Default of 10 last search history without duplicates in Search Radio Stations page
- Add a Settings menu to change the number of search history storage

### (Done: v1.2.0) I feel lucky page
- I feel lucky page: 10 last history without duplicates
- Add a Setting menu to change the number of I feel lucky history storage