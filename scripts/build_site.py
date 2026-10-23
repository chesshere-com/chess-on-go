#!/usr/bin/env python3
"""Build the dependency-free documentation site and check every local link."""
from __future__ import annotations

import html
from html.parser import HTMLParser
import json
from pathlib import Path
import re
import shutil
from statistics import median
from urllib.parse import unquote, urlsplit

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "site"
OUTPUT = ROOT / "_site"
REPO = "https://github.com/chesshere-com/chess-on-go"
RAW = "2026-10-24-corentings.txt"

CASES = [
    ("Perft/Initial", "Perft · initial position, depth 4", "197,281 leaf nodes. Root setup and warm-up excluded; both traversals bulk-count depth one."),
    ("Perft/Kiwipete", "Perft · Kiwipete, depth 3", "97,862 leaf nodes. Mutable make/unmake versus child Position.Update; no PGN or draw bookkeeping."),
    ("LoadFENAndLegalMoves/Initial", "Load FEN + legal moves · initial", "FEN parsing, game construction, and legal-move access are inside the timer. Constructor work differs."),
    ("LoadFENAndLegalMoves/Kiwipete", "Load FEN + legal moves · Kiwipete", "FEN parsing, game construction, and legal-move access are inside the timer. Constructor work differs."),
    ("FENExport/Initial", "Export FEN · initial position", "Similar timings: the sample ranges overlap. The lower allocation count does not establish a meaningful timing win."),
    ("CachedLegalMovesCopy/Initial", "Copy cached legal moves · initial", "Both APIs return newly allocated defensive copies. Measures copying and move representation, not generation."),
    ("ParseSAN/Initial", "Parse SAN · Nf3", "Legal-move caches are warm. Decode one simple SAN move without applying it."),
    ("LoadPGN/15Plies", "Load PGN · 15 plies", "Same unannotated main line and final FEN; corentings builds a richer game tree."),
]


def escape(value):
    return html.escape(str(value), quote=True)


def script_json(value):
    return json.dumps(value, ensure_ascii=False).replace("<", "\\u003c")


def time_label(ns):
    if ns >= 1_000_000:
        return f"{ns / 1_000_000:,.3f} ms"
    if ns >= 1_000:
        return f"{ns / 1_000:,.3f} µs"
    return f"{ns:,.2f} ns"


def benchmark_rows():
    samples = {}
    pattern = re.compile(r"^BenchmarkCompare/(\S+)\s+\d+\s+([\d.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op$")
    raw = (ROOT / "docs" / "benchmarks" / RAW).read_text()
    if "\nPASS\n" not in raw:
        raise ValueError("Benchmark run did not pass")
    for line in raw.splitlines():
        match = pattern.match(line)
        if match:
            key = re.sub(r"-\d+$", "", match[1])
            samples.setdefault(key, []).append([float(x) for x in match.groups()[1:]])
    rows = []
    for key, title, note in CASES:
        row = {"id": key, "title": title, "note": note}
        for library, name in (("chess-on-go", "ours"), ("corentings", "other")):
            values = samples[key + "/" + library]
            if len(values) != 8:
                raise ValueError(f"Expected 8 samples for {key}/{library}, got {len(values)}")
            ns, memory, allocs = [median(v[i] for v in values) for i in range(3)]
            row[name] = {"ns": ns, "bytes": memory, "allocs": allocs, "formatted": time_label(ns),
                         "min": min(v[0] for v in values), "max": max(v[0] for v in values)}
        ratio = row["other"]["ns"] / row["ours"]["ns"]
        row["ratio"] = ratio
        row["result"] = "Similar FEN export timings" if key.startswith("FENExport") else f"{ratio:.1f}× time ratio (corentings / chess-on-go)"
        rows.append(row)
    return rows


def benchmark_table(rows):
    parts = ['<div class="table-wrap"><table><caption class="sr-only">Median time, bytes, and allocations per operation; eight samples per library.</caption><thead><tr><th>Workload</th><th class="numeric">chess-on-go</th><th class="numeric">corentings</th><th class="numeric">Time ratio</th></tr></thead><tbody>']
    for row in rows:
        parts.append(f'<tr><td>{escape(row["title"])}<small>min–max shown below each median</small></td>')
        for name in ("ours", "other"):
            value = row[name]
            parts.append(f'<td class="numeric"><strong>{value["formatted"]}</strong><small>{time_label(value["min"])}–{time_label(value["max"])}</small><small>{value["bytes"]:,.0f} B · {value["allocs"]:g} allocs</small></td>')
        label = "≈ 1×" if row["id"].startswith("FENExport") else f'{row["ratio"]:.1f}×'
        parts.append(f'<td class="numeric">{label}</td></tr>')
    return "".join(parts) + "</tbody></table></div>"


def chart(rows, interactive):
    row = rows[0]
    max_ns = max(row["ours"]["ns"], row["other"]["ns"])
    parts = ['<div class="chart">']
    if interactive:
        options = "".join(f'<option value="{r["id"]}">{escape(r["title"])}</option>' for r in rows)
        parts.append(f'<div class="chart-control" hidden><label for="workload">Choose a workload</label><select id="workload">{options}</select></div>')
    parts.append(f'<p id="chart-title"><strong>{escape(row["title"])}</strong></p>')
    for name, label in (("ours", "chess-on-go"), ("other", "corentings/chess")):
        value = row[name]
        style = " other" if name == "other" else ""
        parts.append(f'<div class="bar-row"><div class="bar-label"><span>{label}</span><strong id="time-{name}">{value["formatted"]}</strong></div><div class="bar-track" aria-hidden="true"><div id="bar-{name}" class="bar{style}" style="width:{value["ns"]/max_ns*100:.3f}%"></div></div></div>')
    parts.append(f'<p class="chart-result" id="chart-result" aria-live="polite">{escape(row["result"])}</p><p class="chart-note" id="chart-note">{escape(row["note"])}</p>')
    parts.append('<p class="chart-note">Apple M5 Pro · Go 1.27.1 · eight samples · lower is better</p></div>')
    if interactive:
        parts.append(f'<script type="application/json" id="benchmark-data">{script_json(rows)}</script>')
    return "".join(parts)


GO_TOKENS = re.compile(r'(?P<comment>//[^\n]*|/\*[\s\S]*?\*/)|(?P<string>"(?:\\.|[^"\\])*"|`[\s\S]*?`)|(?P<keyword>\b(?:package|import|func|return|if|else|for|range|var|type|struct|defer|nil|true|false|string|int)\b)|(?P<number>\b\d+\b)')


def highlight_go(source):
    out, start = [], 0
    for token in GO_TOKENS.finditer(source):
        out.append(escape(source[start:token.start()]))
        out.append(f'<span class="syn-{token.lastgroup}">{escape(token[0])}</span>')
        start = token.end()
    return "".join(out) + escape(source[start:])


def example(name):
    if not re.fullmatch(r"[a-z]+", name):
        raise ValueError("Invalid example name")
    code = (ROOT / "examples" / name / "main.go").read_text().rstrip()
    return (f'<div class="codeblock"><div class="codeheading"><span>examples/{name}/main.go</span>'
            '<button type="button" data-copy hidden>Copy</button></div>'
            f'<pre><code>{highlight_go(code)}</code></pre></div>')


class Document(HTMLParser):
    def __init__(self):
        super().__init__(convert_charrefs=True)
        self.ids, self.links, self.words = set(), [], []
        self.title_count = 0
        self.h1_count = 0

    def handle_starttag(self, tag, attrs):
        attributes = dict(attrs)
        if tag == "title":
            self.title_count += 1
        if tag == "h1":
            self.h1_count += 1
        if "id" in attributes:
            value = attributes["id"]
            if value in self.ids:
                raise ValueError(f"Duplicate ID: {value}")
            self.ids.add(value)
        for key in ("href", "src"):
            if key in attributes:
                self.links.append(attributes[key])

    def handle_data(self, text):
        self.words.append(text)


def check_links():
    parsed = {}
    for file in OUTPUT.rglob("*.html"):
        parser = Document()
        parser.feed(file.read_text())
        if parser.title_count != 1 or parser.h1_count != 1:
            raise ValueError(f"{file.name}: expected one title and h1")
        parsed[file.resolve()] = parser
    count = 0
    for file, parser in parsed.items():
        for link in parser.links:
            url = urlsplit(link)
            if url.scheme or url.netloc:
                continue
            if url.path.startswith("/chess-on-go/"):
                target = (OUTPUT / unquote(url.path.removeprefix("/chess-on-go/"))).resolve()
            else:
                target = (file.parent / unquote(url.path)).resolve() if url.path else file
            if not target.is_relative_to(OUTPUT.resolve()):
                raise ValueError(f"Link escapes site: {file.name}: {link}")
            if target.is_dir():
                target /= "index.html"
            if not target.is_file():
                raise ValueError(f"Broken link: {file.name}: {link}")
            if url.fragment and target in parsed and unquote(url.fragment) not in parsed[target].ids:
                raise ValueError(f"Missing anchor: {file.name}: {link}")
            count += 1
    return len(parsed), count


def main():
    pages = json.loads((SOURCE / "pages.json").read_text())
    template = (SOURCE / "template.html").read_text()
    rows = benchmark_rows()
    contents, search_pages = {}, []
    for page in pages:
        content = (SOURCE / "pages" / page["source"]).read_text()
        content = re.sub(r"\{\{example:([a-z]+)\}\}", lambda m: example(m[1]), content)
        content = content.replace("{{benchmark-chart}}", chart(rows, True))
        content = content.replace("{{benchmark-table}}", benchmark_table(rows))
        content = content.replace("{{home-benchmark}}", chart(rows, False))
        contents[page["path"]] = content
        parser = Document()
        parser.feed(content)
        search_pages.append({"title": page["title"], "path": page["path"], "text": " ".join(parser.words)})
    OUTPUT.mkdir(exist_ok=True)
    shutil.copytree(SOURCE / "assets", OUTPUT / "assets", dirs_exist_ok=True)
    shutil.copytree(ROOT / "docs" / "benchmarks", OUTPUT / "benchmarks", dirs_exist_ok=True)
    for index, page in enumerate(pages):
        prefix = "../" * (len(Path(page["path"]).parts) - 1)
        topnav = []
        for label, path, section in (("Docs", "docs/index.html", "Documentation"), ("Examples", "examples/index.html", "Examples"), ("Performance", "performance/index.html", "Performance")):
            current = ' aria-current="page"' if page["section"] == section else ""
            topnav.append(f'<a href="{prefix}{path}"{current}>{label}</a>')
        content = contents[page["path"]]
        if page["path"] == "index.html":
            layout = content
        else:
            links = []
            for target in pages[1:]:
                current = ' aria-current="page"' if target == page else ""
                links.append(f'<a href="{prefix}{target["path"]}"{current}>{escape(target["title"])}</a>')
            search = [{"title": x["title"], "href": prefix + x["path"], "text": x["text"]} for x in search_pages]
            sidebar = ('<aside class="sidebar" aria-label="Documentation navigation"><div class="sidebar-inner">'
                       '<p class="side-label">Documentation</p><nav aria-label="Documentation">' + "".join(links) + '</nav>'
                       '<div class="search-box" hidden><label class="sr-only" for="docs-search">Search documentation</label>'
                       '<input type="search" id="docs-search" placeholder="Search the docs…" autocomplete="off" aria-controls="search-results">'
                       '<ul class="search-results" id="search-results" aria-live="polite"></ul></div>'
                       '<nav class="side-external" aria-label="External resources"><a href="https://pkg.go.dev/github.com/chesshere-com/chess-on-go">Full Go reference ↗</a>'
                       '<a href="https://www.chesshere.com">Play on ChessHere ↗</a></nav></div></aside>')
            adjacent = ""
            if index > 1:
                previous = pages[index-1]
                adjacent += f'<a href="{prefix}{previous["path"]}">← {escape(previous["title"])}</a>'
            if index < len(pages)-1:
                following = pages[index+1]
                adjacent += f'<a href="{prefix}{following["path"]}">{escape(following["title"])} →</a>'
            layout = (f'<div class="docs-layout">{sidebar}<main class="prose" id="main"><p class="eyebrow">{page["section"]}</p>'
                      f'<h1>{escape(page["title"])}</h1>{content}<nav class="prev-next" aria-label="Adjacent pages">{adjacent}</nav></main></div>'
                      f'<script type="application/json" id="search-data">{script_json(search)}</script>')
        output = template.replace("{{layout}}", layout).replace("{{topnav}}", "".join(topnav))
        for key, value in (("title", page["title"]), ("description", page["description"]), ("root", prefix), ("path", page["path"])):
            output = output.replace("{{" + key + "}}", escape(value))
        if re.search(r"\{\{[^}]+\}\}", output):
            raise ValueError(f"Unresolved template in {page['path']}")
        destination = OUTPUT / page["path"]
        destination.parent.mkdir(parents=True, exist_ok=True)
        destination.write_text(output)
    # A real 404 page, with absolute links so it works at any missing path.
    (OUTPUT / "404.html").write_text('<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Page not found · chess-on-go</title><link rel="stylesheet" href="/chess-on-go/assets/site.css"><main class="shell section"><p class="eyebrow">404</p><h1>That page is not on the board.</h1><p><a href="/chess-on-go/">Go to the documentation homepage →</a></p></main></html>')
    page_count, link_count = check_links()
    (OUTPUT / ".nojekyll").write_text("")
    (OUTPUT / "sitemap.xml").write_text('<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">' + "".join(f'<url><loc>https://chesshere-com.github.io/chess-on-go/{p["path"]}</loc></url>' for p in pages) + '</urlset>\n')
    (OUTPUT / "assets" / "benchmarks.json").write_text(json.dumps(rows, indent=2) + "\n")
    size = sum(p.stat().st_size for p in OUTPUT.rglob("*") if p.is_file())
    print(f"Built {page_count} pages; checked {link_count} local links. Total: {size:,} bytes.")


if __name__ == "__main__":
    main()
