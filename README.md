# DS1 Remastered Boss & Death Counter

A Windows desktop app for Dark Souls Remastered with two counters:

- **Bosses:** automatically tracks the 26 main bosses, including the Artorias of the Abyss DLC. Mini-bosses such as the Hydra are not included.
- **Deaths:** reads the game's death counter and keeps a separate total in the app. The total can be changed or reset manually.

## Requirements

- Windows 10 or 11, 64-bit
- Dark Souls Remastered
- Microsoft Edge WebView2 Runtime

WebView2 Runtime is preinstalled on most Windows 11 and current Windows 10 systems. If the app window does not open, install the [Evergreen Standalone Installer (x64)](https://developer.microsoft.com/microsoft-edge/webview2/).

## Download and run

1. Download `DS1 Counter.exe` from the [latest GitHub Release](https://github.com/MaksEvs/DS1-Remastered-Boss-Counter/releases/latest).
2. Start Dark Souls Remastered and load a character into the game world.
3. Run `DS1 Counter.exe`.
4. The app checks the game every two seconds. The connection status stays visible above the tabs.

The boss counter reflects the currently loaded character. Its last value is saved and shown when the app starts; after connecting to the game, it refreshes from the current character.

The deaths counter is an app-wide total. It is saved between app and computer restarts. Use **−**, **+1**, **Set**, or **Reset** to adjust it. Automatic tracking reads the game's per-character death value from memory and adds single-step increases to the app-wide total.

Application settings and counter values are stored in the user's configuration directory, in `ds1-death-counter/state.json`.

## OBS setup

Add a **Browser Source** in OBS and use the URL for the counter you want to show:

- Bosses: `http://127.0.0.1:7373/obs.html`
- Deaths: `http://127.0.0.1:7373/deaths.html`

Both overlays have transparent backgrounds and refresh automatically. The app must be running while OBS uses them. Plain-text endpoints are also available at `/obs.txt` and `/deaths.txt`.

## Credits and references

- Boss event flag IDs and memory-reading references: [DS1 Item Tracker](https://github.com/MaiAndr/Ds1ItemTracker) and [DSAP](https://github.com/tathxo/DSAP).
- The `GameDataMan` scan signature and DSR memory-reading approach are based on published community research in [DSRtracker's GameReader](https://github.com/spidy666/DSRtracker/blob/main/GameReader.cs). The death counter field offset (`GameDataMan + 0x98`) is documented in [DS Death Counter's reverse-engineering notes](https://github.com/Theuzin69/ds-death-counter/blob/main/docs/how-it-was-made.md).
- DSRtracker is licensed under GPL-3.0. This project uses the published byte signature as factual reference and does not copy its source implementation.
