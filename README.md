# Polling rate switcher

- Only supports xm2 8k
- Checks the current title of the active window and then switches to 1k
- Otherwise switches to 8k
- Only 3MB of RAM c:
- Thanks to https://github.com/niansa/UnofficialEGGMouseConfig

- ![image](https://github.com/user-attachments/assets/4db6b822-8ed2-42f4-a807-56a98fd15f91)


## Internals
- Checks current selected application every 5 seconds when enabled
- Checks every second when disabled
- Only changes polling rate

## Caveats
- Yes its 3.4MB (windows) but it lives in systray (golang was fast to write)
- Should be cross platform compatible. You can try compile urself (and add in the linux/macOS specific stuff to get the active/focused window)
- ONLY tested on windows
- Linux requires udev rules
