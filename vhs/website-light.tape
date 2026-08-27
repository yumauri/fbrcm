# Website hero recording matching the site's light mode.
Output docs/site/public/fbrcm-light.gif

Require go
Require ttyd
Require ffmpeg

Set Shell zsh
Set Width 1200
Set Height 720
Set FontSize 18
Set FontFamily "Cascadia Code"
Set LineHeight 1.0
Set LetterSpacing 0
Set TypingSpeed 80 ms
Set Theme { "name": "fbrcm light", "black": "#3C3C43", "red": "#B42318", "green": "#18794E", "yellow": "#9A6500", "blue": "#245EA8", "magenta": "#7040A0", "cyan": "#087A83", "white": "#F6F6F7", "brightBlack": "#67676C", "brightRed": "#BD3F16", "brightGreen": "#567A11", "brightYellow": "#A95A25", "brightBlue": "#4F46A5", "brightMagenta": "#A63D6F", "brightCyan": "#147A6E", "brightWhite": "#FFFFFF", "background": "#FFFFFF", "foreground": "#3C3C43", "selection": "#FCE8E1", "cursor": "#E85D2A" }
Set Padding 0
Set Margin 0
Set BorderRadius 0
Set Framerate 24
Set CursorBlink false

Env FBRCM_PROFILE "personal"
Env FBRCM_LOG_LEVEL "silent"
Env FBRCM_VHS_THEME "fbrcm-light"
Env FBRCM_VHS_SESSION "/tmp/fbrcm-vhs-website-light"
Env FBRCM_VHS_BINARY "/tmp/fbrcm-vhs-website-light-bin"
Env NO_COLOR ""
Env TERM "xterm-256color"
Env COLORTERM "truecolor"

Source vhs/website-scenario.tape
