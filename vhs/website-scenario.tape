# Shared story for the README and both documentation themes.
Hide
Type "python3 vhs/launch.py $FBRCM_VHS_THEME"
Enter
Wait+Screen@60s /Select project in Projects panel/
Wait+Screen@10s /server@pulseforge-fitness-f60f7/
Down
Enter
Wait+Screen@45s /account_nickname_max_length/

# Prepare the Northstar cover with the banking summary and onboarding values.
PageDown 2
Left
PageDown
Down
Space
Down 5
Space
Up@30ms 32
Type "0"
Type@20ms "------"
Type "1"
Wait+Screen@10s /android_beta_10/
Show
# Screenshot: overview
Sleep 1800 ms

# Evaluate Northstar's passkey prompt for the Android beta cohort.
Type "2"
Type "~"
Type@65ms "enable_passkey_signup"
Wait+Screen@10s /enable_passkey_signup/
Down 2
Type "9"
Type "a"
Wait+Screen@10s /Add Context Input/
Enter
Wait+Screen@10s /device.os=/
Type "android"
Enter
Wait+Screen@10s /Context on/
Type "a"
Wait+Screen@10s /Add Context Input/
Down 9
Enter
Wait+Screen@10s /Field=value/
Type "percent=5"
Enter
Type "2"
Ctrl+e
Wait+Screen@10s /Explain enable_passkey_signup with Context/
# Screenshot: context
Sleep 1700 ms
Escape
Type "9"
Type "X"
Type "2"
Type "~"
Escape

# Inspect a PulseForge parameter with default and conditional values.
Type "0"
Type "[[[["
Type "1"
Down
Enter
Wait+Screen@45s /enable_visual_search/
Type "z"
Type "~"
Type@80ms "premium_recipe"
Wait+Screen@10s /premium_recipe_count/
Down 2
Enter
Wait+Screen@10s /premium_users/
# Screenshot: details
Sleep 1800 ms
Escape
Type "~"
Escape
Type "z"
Type "0"
Type@20ms "------"
Type@20ms "====="
Type "2"
Enter
Wait+Screen@10s /premium_users/
Type "0"
Up
Type "5"
# Screenshot: workspace
Sleep 500 ms
Escape
Type "0"
Down
Type "]]]]"
Type "c"

# Preview selected changes without staging or publishing them.
Type "1"
Type "v"
Wait+Screen@10s /Promote to/
Type@80ms "signal"
Wait+Screen@10s /Filter: signal/
Enter
Wait+Screen@45s /3 to add/
Type "z"
Type "a"
Wait+Screen@10s /4 selected/
# Screenshot: promotion
Sleep 1500 ms
Enter
Wait+Screen@10s /(Parameter|Property): app_shell/
Sleep 2500 ms
