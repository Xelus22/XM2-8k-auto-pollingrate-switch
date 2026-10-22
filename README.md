# Polling rate switcher

- Supports XM2 8k (PID `0x1966`) and XM2 8k v2 (PID `0x1980`), both VID `0x3367`
- Checks the current title of the active window and then switches to 1k
- Otherwise switches to 8k
- Only 3MB of RAM c:
- Thanks to https://github.com/niansa/UnofficialEGGMouseConfig
- Device IDs and packet layout verified against https://github.com/qsxcv/UnofficialEGGWebConfig

- ![image](https://github.com/user-attachments/assets/4db6b822-8ed2-42f4-a807-56a98fd15f91)


## Internals
- Uses a Windows foreground window event hook instead of polling
- Debounces window changes for 2 seconds before applying a new rate
- Reads the current title of the active window and matches against configured targets
- Only changes polling rate

## Caveats
- Yes its 3.4MB (windows) but it lives in systray (golang was fast to write)
- Should be cross platform compatible. You can try compile urself (and add in the linux/macOS specific stuff to get the active/focused window)
- ONLY tested on windows
- Linux requires udev rules
