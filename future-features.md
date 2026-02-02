# Future Features

## Search page

- Default of 10 last search history without duplicates in Search Radio Staions page
- Add a Setting menu to change the number of search history storage

## I feel lucky page
- I feel lucky page: 10 last history without duplicates
- Add a Setting menu to change the number of I feel lucky history storage

## Search page
- #42: color code for stream speed (e.g., 64 kbps, 128 kbps, 320 kbps)
- Sort by stream speed

## Auto connect
- #4: Using GPRS/4G, I sometimes lose the signal/connection and then have to reconnect manually to the station I was listening to.

### Possilbe solutions (I'm not sure if any of these works)
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

## Settings
- Smart Update commands
In Check for Updates page, if an update is available, add a list to update using the following method.
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
