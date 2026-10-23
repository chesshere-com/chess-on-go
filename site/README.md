# Documentation website

The site uses static HTML, CSS, a small progressive-enhancement script, and a
Python standard-library builder. No npm install or JavaScript framework is needed.
It follows the restrained charcoal/green style of the ChessHere chessboard site.

From the repository root:

```sh
python3 scripts/build_site.py
python3 -m http.server 8000 --directory _site
```

Open http://localhost:8000. The generated output is ignored by Git.

- Edit page content in `site/pages/`; page titles/navigation live in `pages.json`.
- Complete Go snippets are included from `examples/*/main.go`. Change the actual
  program rather than maintaining a duplicate snippet in HTML.
- Benchmark charts/tables are built from the committed raw Go benchmark output.
  The measured engine revision, release, machine, and timing boundaries are
  recorded on the performance page and in the raw sample header.
- `scripts/build_site.py` checks all local links, fragments, duplicate IDs, and
  page title / primary-heading counts. Relative asset URLs work both locally and
  under GitHub Pages' `/chess-on-go/` project prefix.
- Navigation, documentation, code samples, benchmark bars, and the full results
  table work without JavaScript. JavaScript adds copy buttons, search, a color
  theme switch, and the benchmark workload selector.

## GitHub Pages

The `Documentation site` workflow builds and verifies the examples and benchmark
equivalence on pull requests. Pushes to `main` publish through the official
GitHub Pages artifact/deployment actions. The initial
`codex/documentation-site` branch can also publish, allowing the first website to
go live before the source PR is merged; delete that branch after merging.

Enable **Settings → Pages → Build and deployment → GitHub Actions**.
The `github-pages` environment must allow the publishing branch. This site uses
the standard project URL:

https://chesshere-com.github.io/chess-on-go/

No custom domain is required. The site links to https://www.chesshere.com;
it does not change that site's domain or content.

## Review

```sh
go test ./examples/...
(cd benchmarks/corentings && go test ./...)
python3 scripts/build_site.py
```

Also check desktop and narrow layouts, keyboard navigation, search, copy,
theme switching, and the performance selector in a browser. Benchmark numbers
are historical measurements, not live browser timings.
