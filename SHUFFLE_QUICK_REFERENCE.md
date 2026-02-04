# Shuffle Mode - Quick Reference

## 🚀 Quick Start

```bash
# In TERA:
1. Press '4' → I Feel Lucky
2. Press 't' → Enable shuffle mode  
3. Type 'jazz' → Enter
4. Enjoy! 🎵
```

## ⌨️ Keyboard Controls

| Key   | What It Does                                     |
|-------|--------------------------------------------------|
| `t`   | Toggle shuffle on/off (input screen)             |
| `n`   | Next station →                                   |
| `b`   | Previous station ←                               |
| `p`   | Pause/resume timer ⏸▶                            |
| `h`   | Stop shuffle, keep playing                       |
| `f`   | ⭐ Save to favorites                             |
| `s`   | 💾 Save to list                                  |
| `v`   | 🗳️ Vote                                          |
| `Esc` | Stop & return to input                           |
| `0`   | Stop & return to main menu                       |
| `?`   | Help                                             |

## ⚙️ Settings

**Settings → Shuffle Settings**

1. **Auto-advance**: Automatic station switching
   - Off = Manual control only
   - On = Timer-based switching

2. **Interval**: How long to play each station
   - Options: 1, 3, 5, 10, 15 minutes
   - Default: 5 minutes

3. **History**: Remember previous stations
   - Off = No backward navigation
   - On = Press 'b' to go back

4. **History Size**: How many to remember
   - Options: 3, 5, 7, 10 stations
   - Default: 5 stations

## 📊 What You'll See

### Input Screen (Shuffle On)
```
Shuffle mode: [✓] On  (press 't' to disable)
              Auto-advance in 5 min • History: 5 stations
```

### Playing Station
```
🎵 Now Playing (🔀 Shuffle: jazz)

Station: Smooth Jazz 24/7
▶ Playing...

🔀 Shuffle Active • Next in: 4:23
   Station 3 of session
   
─── Shuffle History ───
  ← Jazz FM London
  ← WBGO Jazz 88.3  
  → Smooth Jazz 24/7  ← Current
```

## 💡 Pro Tips

1. **Manual Control**: Disable auto-advance in settings to control when to skip
2. **Quick Save**: Press 'f' during shuffle to save any station you like
3. **Go Back**: Made a mistake? Press 'b' to return to previous station
4. **Pause Timer**: Need more time? Press 'p' to pause the countdown
5. **Stop Shuffle**: Press 'h' to stop shuffling but keep playing current station

## 📁 Config File

Location: `~/.config/tera/shuffle.yaml`

```yaml
shuffle:
  auto_advance: true       # Timer-based or manual
  interval_minutes: 5      # 1, 3, 5, 10, or 15
  remember_history: true   # Track previous stations
  max_history: 5           # 3, 5, 7, or 10
```

## 🐛 Troubleshooting

**Timer not working?**
- Check Settings → Shuffle Settings → Auto-advance is ON

**Can't go back?**
- Check Settings → Shuffle Settings → Remember History is ON

**No stations found?**
- Try a different keyword (jazz, rock, classical, news)
- Check your internet connection

**Stuck in shuffle?**
- Press `Esc` to stop shuffle and return to input
- Press `0` to stop shuffle and return to main menu

## 🎯 Common Use Cases

**Discovery Mode**
```
Auto-advance: ON
Interval: 3 minutes
History: OFF
→ Quick sampling of many stations
```

**Curated Listening**
```
Auto-advance: OFF
Interval: N/A
History: ON (10 stations)
→ Manual control with full history
```

**Background Music**
```
Auto-advance: ON
Interval: 15 minutes
History: ON (5 stations)
→ Long sessions with some variety
```

## 📝 Testing

Run tests:
```bash
cd /Users/shinichiokada/Terminal-Tools/tera
go test ./internal/ui -v -run "TestLuckyShuffle"
```

Build and try:
```bash
go build -o tera cmd/tera/main.go
./tera
```

---

**Happy Shuffling! 🎵🔀**

For detailed documentation, see:
- README.md → "Shuffle Mode" section
- SHUFFLE_IMPLEMENTATION_COMPLETE.md → Full technical details
- SHUFFLE_TESTING_CHECKLIST.md → Comprehensive testing guide
