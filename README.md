# DS1 Remastered Boss Counter

A small Windows app that automatically counts the 26 main bosses in Dark Souls Remastered, including the Artorias of the Abyss DLC. Mini-bosses such as the Hydra are not included.

## Requirements

- Windows 10 or 11, 64-bit
- Dark Souls Remastered
- Microsoft Edge WebView2 Runtime

WebView2 Runtime is preinstalled on most Windows 11 and current Windows 10 systems. If the app window does not open, install the **Evergreen Standalone Installer (x64)** from the [Microsoft WebView2 download page](https://developer.microsoft.com/microsoft-edge/webview2/).

## Download and run

1. Download `ds1-death-counter.exe` from the [latest GitHub Release](https://github.com/MaksEvs/DS1-Remastered-Boss-Counter/releases/latest).
2. Start Dark Souls Remastered and load a character into the game world.
3. Run `ds1-death-counter.exe`.
4. Check the connection status and boss count in the app window.

The counter reads boss event flags for the currently loaded character from the game's memory. It does not open or modify save files. When you switch characters, the count updates to match the newly loaded save. Keep the app running while using the OBS overlay.

## OBS setup

1. In OBS, add a **Browser Source**.
2. Set its URL to:

   ```text
   http://127.0.0.1:7373/obs.html?v=9
   ```

The overlay uses a transparent background and refreshes automatically once per second. The counter app must be running for the overlay to receive updates.

## References

- [DS1 Item Tracker — boss event flag IDs](https://github.com/MaiAndr/Ds1ItemTracker)
- [DSAP — event flag memory and address research](https://github.com/tathxo/DSAP)
