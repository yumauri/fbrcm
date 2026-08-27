# Shared website recording scenario. The caller provides the theme, temporary
# working directory, and binary path through FBRCM_VHS_* environment variables.

# Build and launch without spending the viewer's attention on setup. A local
# config in a temporary directory selects the recording theme without changing
# the user's saved global or repository configuration.
Hide
Type "go build -o $FBRCM_VHS_BINARY ."
Enter
Sleep 5 s
Type "mkdir -p $FBRCM_VHS_SESSION"
Enter
Type "cd $FBRCM_VHS_SESSION"
Enter
Type "$FBRCM_VHS_BINARY theme switch $FBRCM_VHS_THEME --scope local"
Enter
Type "$FBRCM_VHS_BINARY project templates set pulseforge-fitness-f60f7 --templates client,server --primary client"
Enter
Sleep 2 s
Type "$FBRCM_VHS_BINARY"
Enter
Wait+Screen@30s /Projects/
Down
Enter
Wait+Screen@30s /account_nickname_max_length/

# Compose a useful cover frame before revealing the recording. Expand three
# onboarding parameters, collapse Insights, and leave the Projects cursor on
# Northstar Wallet.
Down 23
Space
Up 4
Space
Up
Space
Up 9
Left
Up 9
Type "1"
Show

# Start with Northstar Wallet's prepared parameter tree, then navigate to
# PulseForge Fitness.
Sleep 1 s
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
