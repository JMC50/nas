#!/usr/bin/env python3
"""Regenerate every logo / favicon variant from a single source PNG.

Default source: Docs/images/logo-source.png (high-resolution master).
Outputs (all overwritten in place):
  frontend/static/favicon-16.png   16x16
  frontend/static/favicon-32.png   32x32
  frontend/static/favicon-48.png   48x48
  frontend/static/favicon-64.png   64x64
  frontend/static/apple-touch-icon.png   180x180
  frontend/static/favicon.ico      multi-res (16/32/48/64)
  frontend/static/logo.png         256x256
  Docs/images/logo.png             256x256 (mirrors the static one for README)

Re-run any time the master logo changes. Uses LANCZOS resampling and
preserves the source alpha channel so the orange hex on transparent
background survives every downscale.
"""

import sys
from pathlib import Path
from PIL import Image

REPO_ROOT = Path(__file__).resolve().parents[1]
DEFAULT_SOURCE = REPO_ROOT / "Docs" / "images" / "logo-source.png"

TARGETS = [
    ("frontend/static/favicon-16.png", 16),
    ("frontend/static/favicon-32.png", 32),
    ("frontend/static/favicon-48.png", 48),
    ("frontend/static/favicon-64.png", 64),
    ("frontend/static/apple-touch-icon.png", 180),
    ("frontend/static/logo.png", 256),
    ("Docs/images/logo.png", 256),
]
ICO_PATH = REPO_ROOT / "frontend" / "static" / "favicon.ico"
ICO_SIZES = [(16, 16), (32, 32), (48, 48), (64, 64)]


def resize(source: Image.Image, size: int) -> Image.Image:
    if source.mode != "RGBA":
        source = source.convert("RGBA")
    return source.resize((size, size), Image.Resampling.LANCZOS)


def main() -> int:
    src = Path(sys.argv[1]) if len(sys.argv) > 1 else DEFAULT_SOURCE
    if not src.exists():
        print(f"error: source not found: {src}", file=sys.stderr)
        return 2

    master = Image.open(src)
    print(f"source: {src} ({master.size[0]}x{master.size[1]} {master.mode})")

    for rel, size in TARGETS:
        out_path = REPO_ROOT / rel
        out_path.parent.mkdir(parents=True, exist_ok=True)
        resize(master, size).save(out_path, format="PNG", optimize=True)
        print(f"  wrote {rel} ({size}x{size})")

    # Multi-resolution ICO. PIL writes one ICO containing all sizes in one file.
    ICO_PATH.parent.mkdir(parents=True, exist_ok=True)
    base = master if master.mode == "RGBA" else master.convert("RGBA")
    base.save(ICO_PATH, format="ICO", sizes=ICO_SIZES)
    print(f"  wrote {ICO_PATH.relative_to(REPO_ROOT)} (multi-res {ICO_SIZES})")
    return 0


if __name__ == "__main__":
    sys.exit(main())
