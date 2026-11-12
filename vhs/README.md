# Record the fbrcm demo

From the repository root, run `FBRCM_VHS_LIVE_PROFILE=personal python3 vhs/record.py`
to regenerate the README GIF, both website GIFs, and the light and dark website
stills from the live demo projects. The tapes need
VHS, ttyd, FFmpeg, Go, Python 3, and the Cascadia Code font. To record only one
GIF, run `vhs vhs/demo.tape`, `vhs vhs/website-dark.tape`, or
`vhs vhs/website-light.tape`.

For a live run, the launcher creates a temporary profile containing only the
five demo projects and uses the source profile's current OAuth access token
and demo quota project. The recording opens published client templates and
ends at a promotion preview. Running without `FBRCM_VHS_LIVE_PROFILE` uses
the checked-in JSON templates with `FBRCM_OFFLINE=1`. The GIFs and stills share
`website-scenario.tape`. Its screenshot markers identify the Northstar overview
and Context explanation, PulseForge parameter details, the workspace with
Projects and Details side by side, and the promotion review. The Logs panel
expands for the `workspace` screenshot and collapses for the other captures.

To regenerate just the two Workspace layout PNGs from the live demo projects,
run `FBRCM_VHS_LIVE_PROFILE=personal python3 vhs/record.py --screenshot workspace`.
Use `--help` to list the other screenshot names.

To explore the same fixtures manually, run `python3 vhs/launch.py fbrcm-dark`.
Add `--empty-streamline` to show the empty template for `streamline-play-514406`.
The default Streamline fixture contains playback, discovery, and other demo
settings.

The promotion preview should show four selectable changes with pruning off.
PulseForge adds `premium_recipe_count`, `premium_badge_copy`, and
`player_controls_layout` to SignalOps, and updates `enable_visual_search`.
Both templates already have the condition and group definitions needed by
those changes.

## Reset a demo project

The JSON files are baselines for the same client projects used in the demo.
Northstar Wallet is version 42, Mercato Mobile is version 12, PulseForge
Fitness is version 5, SignalOps Console is version 7, and Streamline Play is
version 1. Publisher identity is omitted from the exported version metadata.
Project-specific conditions and managed-feature references belong to their
original projects. Restoring a template requires any referenced Firebase
resources to still exist.

Stage the fixture as a local version snapshot, then preview the reset. For
example, to reset Northstar Wallet with the `personal` profile:

```sh
python3 vhs/stage_reset.py northstar-wallet --profile personal
fbrcm --profile personal versions restore client@northstar-wallet 42 --dry-run
```

Staging leaves the current cache pointer and drafts in place. If a snapshot
for that version already exists, staging checks its template content and
reports a conflict if it differs. The dry run shows the diff and asks Firebase
to validate any changes. To publish the reset after reviewing it, repeat the
restore command without `--dry-run` and confirm the project ID at the prompt.
Restore publishes a new Firebase version and preserves unused conditions.
