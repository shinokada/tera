# Footer Navigation Plan

Helper menu should be at the center of the terminal.

## Main page

Current:
```text
                                                                                         
                         TERA

  Main Menu & Quick Play 

  Choose an option:

    > 1. Play from Favorites
      2. Search Stations
      3. Manage Lists
      4. I Feel Lucky
      5. Gist Management
      6. Settings

  ─── Quick Play Favorites ───
      10. 101 SMOOTH JAZZ • The United States Of America • MP3 128kbps
    > 11. ▶ 1-NRK Jazz • Norway • MP3 192kbps       
      12. Smooth Radio • The United Kingdom Of Great Britain And Northern Ireland • AAC 48kbps
      13. Classic Vinyl HD • The United States Of America • MP3 320kbps


↑↓/jk: Navigate • Enter: Select • 1-6: Menu • 10+: Quick play • Esc: Stop • Ctrl+C: Quit
```

Update:
```text


  ↑↓/jk: Navigate • Enter: Select • 1-6: Menu • 10+: Quick Play • ?: Help
```

`?: Help` shows:
```text
═══ TERA Help ═══

Navigation
 ↑↓/jk:  Navigate
 Enter:  Select/Play
 1-6:    Main menu
 10+:    Quick play favorites
 Esc:    Stop Playback
 Ctrl+C: Quit
 
Playback
 /*:     Volume 
 m:      Mute

 Press any key to close
```

## Favorites Page

Current:
```text

                         TERA

  🎵 Now Playing

  Name:    Rádio Bossa Nova Brazil 
  Tags:    bossa nova,brazilian music
  Country: Brazil
  Language: portuguese 
  Votes:   12239
  Codec:   MP3 @ 128 kbps

  ▶ Playing... 

  Esc: Back • f: Save to Favorites • s: Save to list • v: Vote • 0: Main Menu • Ctrl+C: Quit
```

Update:/
```text

                         TERA

  🎵 Now Playing

  Name:    Rádio Bossa Nova Brazil 
  Tags:    bossa nova,brazilian music
  Country: Brazil
  Language: portuguese 
  Votes:   12239
  Codec:   MP3 @ 128 kbps

  ▶ Playing... 

  f: Favorites • v: Vote • 0: Main Menu • ?: Help
```

`?: Help` shows:
```text
═══ TERA Help ═══

Navigation
 Esc:    Stop & Back
 0:      Main Menu
 Ctrl+C: Quit
 
Playback
 /*:     Adjust volume 
 m:      Toggle mute

Actions
 f:      Save to Favorites
 v:      Vote

 Press any key to close
```

## Search Results Page

Current:
```tex

                         TERA

  🎵 Now Playing

  Name:    - 0 N - Smooth Jazz on Radio 
  Tags:    chillout,easy listening,jazz,smooth,smoothjazz
  Country: Germany, Bayern
  Language: german
  Votes:   4861  
  Codec:   AAC+ @ 64 kbps
  ▶ Playing...


 Esc: Back • f: Save to Favorites • s: Save to list • v: Vote • 0: Main Menu • Ctrl+C: Quit  
```

Update:
```text

 f: Save to Favorites • s: Save to list • v: Vote • ?: Help  
```

`?: Help` shows:
```text
═══ TERA Help ═══

Navigation
 Esc:    Stop & Back
 0:      Main Menu
 Ctrl+C: Quit
 
Playback
 /*:     Adjust volume 
 m:      Toggle mute

Actions
 f:      Save to Favorites
 s:      Save to List
 v:      Vote

 Press any key to close
```

## I Feel Lucky

Current:
```text

                         TERA

  🎵 Now Playing

  Name:    Qfm
  Tags:    blues,jazz,latin jazz,smooth jazz
  Country: Spain
  Votes:   727
  Codec:   MP3 @ 128 kbps
  ▶ Playing...




 Esc: Stop • f: Save to Favorites • s: Save to list • v: Vote • 0: Main Menu • Ctrl+C: Quit  
```

Update:
```text

 f: Save to Favorites • s: Save to list • v: Vote • ?: Help  
```

`?: Help` shows:
```text
═══ TERA Help ═══

Navigation
 Esc:    Stop & Back
 0:      Main Menu
 Ctrl+C: Quit
 
Playback
 /*:     Adjust volume 
 m:      Toggle Mute

Actions
 f:      Save to Favorites
 s:      Save to List
 v:      Vote

 Press any key to close
```

## Need to remove a page
Remove the following page when you press `Esc`, since the playing page has `f` and `s` keyboard shortcuts.
`Esc` should stop playing the station and back to the Main menu.

```

                         TERA

  💾 Save Station?

  Did you enjoy this station?

  EBS | Lounge

  1) ⭐ Add to Quick Favorites 
  2) Return to Main Menu
  y/1: Yes • n/2/Esc: No  
```