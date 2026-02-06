# Debug Logging for TERA

## How to View Debug Logs

When you run TERA, all `log.Println()` statements are written to `tera_debug.log` in the current directory.

### Watch logs in real-time

In a separate terminal window, run:

```bash
tail -f tera_debug.log
```

### View all logs

```bash
cat tera_debug.log
```

### Clear old logs

```bash
rm tera_debug.log
```

## Example Workflow

**Terminal 1** - Run the app:
```bash
make clean-all && make build && ./tera
```

**Terminal 2** - Watch debug output:
```bash
tail -f tera_debug.log
```

Now when you press `v` to vote, you'll see the debug output in Terminal 2!

## Log Format

Logs include:
- Date and time with microseconds
- Source file and line number
- Your debug message

Example output:
```
2025/02/06 10:30:45.123456 app.go:123: DEBUG: Vote button pressed
2025/02/06 10:30:45.234567 vote.go:45: DEBUG: Station ID: 12345
2025/02/06 10:30:45.345678 vote.go:78: DEBUG: Vote successful
```
