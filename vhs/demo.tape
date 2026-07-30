# README hero: approximately 100x32 terminal cells.
Output vhs/demo.gif

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
Set Theme "Catppuccin Frappe"
Set Padding 0
Set Margin 0
Set BorderRadius 0
Set Framerate 24
Set CursorBlink false

Env FBRCM_PROFILE "personal"
Env FBRCM_LOG_LEVEL "silent"
Env NO_COLOR ""
Env TERM "xterm-256color"
Env COLORTERM "truecolor"

# Build and launch without spending the viewer's attention on setup.
Hide
Type "go build -o /tmp/fbrcm-vhs-demo ."
Enter
Sleep 5 s
Type "/tmp/fbrcm-vhs-demo project templates set pulseforge-fitness-f60f7 --templates client,server --primary client"
Enter
Sleep 2 s
Type "/tmp/fbrcm-vhs-demo"
Enter
Wait+Screen@30s /Projects/
Down
Enter
Wait+Screen@30s /account_nickname_max_length/
Show

# Start on Northstar Wallet, then navigate to PulseForge Fitness.
Sleep 1 s
Type "1"
Down@350ms 1
Enter
Wait+Screen@30s /enable_visual_search/
Sleep 700 ms
Type "z"
Sleep 700 ms

# Open a parameter with both a default and a conditional value.
Type "~"
Sleep 300 ms
Type@100ms "premium_recipe"
Wait+Screen@10s /premium_recipe_count/
Sleep 700 ms
Down@300ms 2
Enter
Wait+Screen@10s /premium_users/
Sleep 1500 ms
Escape
Sleep 400 ms
Type "~"
Sleep 200 ms
Escape
Wait+Screen@10s /premium_recipe_count/
Sleep 500 ms

# Restore the split layout and choose the promotion target.
Type "z"
Sleep 500 ms
Type "1"
Type "v"
Sleep 500 ms
Type@100ms "signal"
Wait+Screen@10s /Filter: signal/
Sleep 700 ms
Enter

# Select the compact change set and finish on its richest diff.
Wait+Screen@30s /3 to add/
Sleep 1000 ms
Type "a"
Wait+Screen@10s /4 selected/
Sleep 900 ms
Enter
Wait+Screen@10s /Property: app_shell/
Sleep 2500 ms
