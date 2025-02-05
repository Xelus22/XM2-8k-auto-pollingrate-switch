# Polling rate switcher

- Only supports xm2 8k
- Checks the current title of the active window and then switches to 1k
- Otherwise switches to 8k
- Thanks to https://github.com/niansa/UnofficialEGGMouseConfig

## Internals
- Checks current selected application every 5 seconds when enabled
- Checks every second when disabled
- Only changes polling rate

## Caveats
- Yes its 3.4MB (windows) but it lives in systray (golang was fast to write)
- Should be cross platform compatible. You can try compile urself
- ONLY tested on windows
- Linux requires udev rules
