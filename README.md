# RetroPlayer

TUI to browse and search retro video game soundtracks from [joshw.info](https://joshw.info) archives,
and launch them with [ZXTune](https://zxtune.bitbucket.io/).

## Features

- Terminal user interface (TUI)
- Real-time search for music files
- Smooth keyboard navigation
- Automatic download and playback
- Local database in `~/.cache/retroplayer`
- Support for 25 retro music formats

## Dependencies

- **zxtune-qt** : chiptune player for playback
- **7z** : archive extraction (p7zip-full)

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/elgabo86/retroplayer/main/install.sh | bash
```

## Usage

```bash
retroplayer --update   # Update database (first use)
retroplayer            # Launch the interface
retroplayer --help     # Show help
```

## Navigation

- **Arrow keys ↑↓** : Browse the list
- **Left/Right** : Jump 5 entries
- **Enter** : Play selected file
- **ESC** : Cancel download / Quit
- **Ctrl+C** : Quit
- **Ctrl+U** : Clear search
- **Backspace** : Delete last character
- **Text input** : Real-time search

## Supported formats

SPC (SNES), NSF (NES), USF (N64), SMD (Genesis), PSF (PS1), PSF2 (PS2), PSF3 (PS3),
PSF4 (PS4), GBS (Game Boy), GSF (GBA), 2SF (NDS), 3SF (3DS), DSF (Dreamcast),
Wii, WiiU, GCN, SWITCH, SSF (Saturn), Xbox, X360, PSP, VITA, PC, 3DO

## License

MIT
