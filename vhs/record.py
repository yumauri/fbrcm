#!/usr/bin/env python3
"""Regenerate the README and website GIFs and the website screenshots with VHS."""

import argparse
import json
from pathlib import Path
import subprocess
import tempfile


VHS = Path(__file__).resolve().parent
ROOT = VHS.parent
SCENARIO = "vhs/website-scenario.tape"
SCREENSHOT_MARKER = "# Screenshot: "


def with_screenshots(scenario, theme, only=None):
    for line in scenario.splitlines():
        if line.startswith(SCREENSHOT_MARKER):
            scene = line.removeprefix(SCREENSHOT_MARKER)
            if only is not None and only != scene:
                continue
            filename = f"fbrcm-{theme}.png" if scene == "overview" else f"tui-{scene}-{theme}.png"
            yield f"Screenshot docs/site/public/{filename}"
            if only is not None:
                yield "Sleep 500 ms"
                return
        else:
            yield line


def main():
    scenario = (ROOT / SCENARIO).read_text(encoding="utf-8")
    scenes = [line.removeprefix(SCREENSHOT_MARKER) for line in scenario.splitlines()
              if line.startswith(SCREENSHOT_MARKER)]
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--screenshot", choices=scenes, help="capture only this scene in both themes")
    args = parser.parse_args()
    if args.screenshot is None:
        subprocess.run(["vhs", "vhs/demo.tape"], cwd=ROOT, check=True)

    with tempfile.TemporaryDirectory(prefix="fbrcm-vhs-tapes-") as directory:
        for theme in ("dark", "light"):
            wrapper = (VHS / f"website-{theme}.tape").read_text(encoding="utf-8")
            tape = Path(directory) / f"website-{theme}.tape"
            if args.screenshot is not None:
                temporary_gif = tape.with_suffix(".gif")
                wrapper = "\n".join(
                    f"Output {json.dumps(str(temporary_gif))}" if line.startswith("Output ") else line
                    for line in wrapper.splitlines()
                )
            story = "\n".join(with_screenshots(scenario, theme, args.screenshot))
            tape.write_text(wrapper.replace(f"Source {SCENARIO}", story), encoding="utf-8")
            subprocess.run(["vhs", str(tape)], cwd=ROOT, check=True)


if __name__ == "__main__":
    main()
