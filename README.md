# Tera

## How to stop all the running stations

If multiple stations are running during the development:

```sh
# Kill all mpv processes
killall mpv

# Or if that doesn't work:
pkill -9 mpv

# check if mpv is running
ps aux | grep mpv | grep -v grep
# or
pgrep -a mpv
# No output → not running
# Output with PID → running
```

## Test

Run all the tests:
```sh
go test ./... -v
```