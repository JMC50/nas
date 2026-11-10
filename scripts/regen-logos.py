#!/usr/bin/env python3
"""사이트 로고 원본 SVG 하나에서 탭 아이콘, 홈 화면 아이콘, README 로고를 만듭니다.

원본은 앱 화면의 로고(`Logo.svelte`)가 그대로 쓰는 `frontend/src/lib/assets/logo-mark.svg`입니다.
원본을 바꾼 뒤 `python scripts/regen-logos.py`를 실행하면 결과 파일과 `app.html`의 아이콘 주소가 함께 바뀝니다.
원본 path는 M, H, V, h, v, z 명령만 쓰는 직선 도형이어야 하고, Pillow가 필요합니다.
"""

import hashlib
import re
import sys
from pathlib import Path

from PIL import Image, ImageChops, ImageDraw

REPO_ROOT = Path(__file__).resolve().parents[1]
SOURCE = REPO_ROOT / "frontend" / "src" / "lib" / "assets" / "logo-mark.svg"
STATIC = REPO_ROOT / "frontend" / "static"
APP_HTML = REPO_ROOT / "frontend" / "src" / "app.html"
README_LOGO = REPO_ROOT / "Docs" / "images" / "logo.png"

# 레일과 머리 막대의 바탕(Signal s1)과 같은 색 타일 위에 로고를 둠. 밝은 탭 막대에서도 라임 로고가 보이게 함
TILE_COLOR = "#131518"
GRID = 32
MARK_OFFSET = 4
TILE_RADIUS = 7
SUPERSAMPLE = 16


def read_source():
    text = SOURCE.read_text(encoding="utf-8")
    view_box = [float(v) for v in re.search(r'viewBox="([^"]+)"', text).group(1).split()]
    path_element = re.search(r"<path\b[^>]*/>", text).group(0)
    fill = re.search(r'fill="(#[0-9a-fA-F]{6})"', path_element).group(1)
    commands = re.search(r'\sd="([^"]+)"', path_element).group(1)
    return text, view_box[2], path_element, fill, commands


def polygons(commands):
    tokens = re.findall(r"[MHVhvz]|-?\d+(?:\.\d+)?", commands)
    shapes, points = [], []
    x = y = 0.0
    index = 0
    while index < len(tokens):
        op = tokens[index]
        index += 1
        if op == "z":
            shapes.append(points)
            points = []
            continue
        value = float(tokens[index])
        index += 1
        if op == "M":
            y = float(tokens[index])
            index += 1
            x = value
        elif op == "H":
            x = value
        elif op == "V":
            y = value
        elif op == "h":
            x += value
        elif op == "v":
            y += value
        else:
            sys.exit(f"unsupported path command: {op}")
        points.append((x, y))
    return shapes


def render(size, mark_box, fill, shapes, rounded):
    big = size * SUPERSAMPLE
    scale = big / GRID
    image = Image.new("RGBA", (big, big), (0, 0, 0, 0))
    draw = ImageDraw.Draw(image)
    radius = TILE_RADIUS * scale if rounded else 0
    draw.rounded_rectangle([0, 0, big - 1, big - 1], radius=radius, fill=TILE_COLOR)
    unit = scale * (GRID - 2 * MARK_OFFSET) / mark_box
    origin = MARK_OFFSET * scale
    mask = Image.new("1", (big, big), 0)
    for shape in shapes:
        layer = Image.new("1", (big, big), 0)
        ImageDraw.Draw(layer).polygon([(origin + px * unit, origin + py * unit) for px, py in shape], fill=1)
        mask = ImageChops.logical_xor(mask, layer)
    image.paste(Image.new("RGBA", (big, big), fill), (0, 0), mask)
    return image.resize((size, size), Image.Resampling.LANCZOS)


def write_svg(path_element, mark_box):
    scale = (GRID - 2 * MARK_OFFSET) / mark_box
    svg = (
        f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {GRID} {GRID}">'
        f'<rect width="{GRID}" height="{GRID}" rx="{TILE_RADIUS}" fill="{TILE_COLOR}"/>'
        f'<g transform="translate({MARK_OFFSET} {MARK_OFFSET}) scale({scale:g})">{path_element}</g></svg>\n'
    )
    (STATIC / "favicon.svg").write_text(svg, encoding="utf-8")


def write_links(version):
    html = APP_HTML.read_text(encoding="utf-8")
    html = re.sub(r'\n\s*<link rel="(?:icon|apple-touch-icon)"[^>]*/>', "", html)
    links = (
        f'\n    <link rel="icon" type="image/svg+xml" href="%sveltekit.assets%/favicon.svg?v={version}" />'
        f'\n    <link rel="icon" type="image/png" sizes="32x32" href="%sveltekit.assets%/favicon-32.png?v={version}" />'
        f'\n    <link rel="apple-touch-icon" sizes="180x180" href="%sveltekit.assets%/apple-touch-icon.png?v={version}" />'
    )
    html = html.replace('<meta charset="utf-8" />', '<meta charset="utf-8" />' + links, 1)
    APP_HTML.write_text(html, encoding="utf-8")


def main():
    text, mark_box, path_element, fill, commands = read_source()
    shapes = polygons(commands)
    version = hashlib.sha256(text.encode("utf-8")).hexdigest()[:8]

    write_svg(path_element, mark_box)
    render(32, mark_box, fill, shapes, rounded=True).save(STATIC / "favicon-32.png", optimize=True)
    # iOS가 홈 화면 아이콘 모서리를 직접 깎으므로 타일을 꽉 채움
    render(180, mark_box, fill, shapes, rounded=False).save(STATIC / "apple-touch-icon.png", optimize=True)
    render(256, mark_box, fill, shapes, rounded=True).save(STATIC / "favicon.ico", sizes=[(16, 16), (32, 32), (48, 48)])
    render(256, mark_box, fill, shapes, rounded=True).save(README_LOGO, optimize=True)
    write_links(version)
    print(f"source {SOURCE.relative_to(REPO_ROOT)} version {version}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
