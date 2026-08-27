# Website hero recording matching the site's dark mode.
Output docs/site/public/fbrcm-dark.gif

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
Set Theme { "name": "fbrcm dark", "black": "#0D0E13", "red": "#BD3F16", "green": "#4CC38A", "yellow": "#FFC400", "blue": "#6EA8FE", "magenta": "#C58AF9", "cyan": "#56D4DD", "white": "#DFDFD6", "brightBlack": "#6A6A71", "brightRed": "#FF6B6B", "brightGreen": "#A7D95B", "brightYellow": "#FF9F43", "brightBlue": "#8B8CF8", "brightMagenta": "#F27EB1", "brightCyan": "#4CC9B0", "brightWhite": "#FFFFFF", "background": "#0D0E13", "foreground": "#DFDFD6", "selection": "#331E19", "cursor": "#FF8A50" }
Set Padding 0
Set Margin 0
Set BorderRadius 0
Set Framerate 24
Set CursorBlink false

Env FBRCM_PROFILE "personal"
Env FBRCM_LOG_LEVEL "silent"
Env FBRCM_VHS_THEME "fbrcm-dark"
Env FBRCM_VHS_SESSION "/tmp/fbrcm-vhs-website-dark"
Env FBRCM_VHS_BINARY "/tmp/fbrcm-vhs-website-dark-bin"
Env NO_COLOR ""
Env TERM "xterm-256color"
Env COLORTERM "truecolor"

Source vhs/website-scenario.tape
