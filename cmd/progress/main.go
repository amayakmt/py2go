// Command progress draws the status of the roadmap: a map with one box per
// phase, and two badges. It reads docs/roadmap.md for the phases,
// docs/decisions.md for which are finished, and testdata/ for the test counts,
// so the pictures cannot say more than the repository does.
//
//	go run ./cmd/progress            writes docs/progress/*.svg
package main

import (
	"fmt"
	"html"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type phase struct {
	n      int
	title  string
	status string // done, next, planned
	tests  int
	steps  []string // the numbered steps of the phase in the roadmap
	cases  []string // the differential tests of the phase
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "progress:", err)
		os.Exit(1)
	}
}

func run() error {
	roadmap, err := os.ReadFile("docs/roadmap.md")
	if err != nil {
		return err
	}
	decisions, err := os.ReadFile("docs/decisions.md")
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`(?m)^### Phase (\d+): (.+)$`)
	sections := re.Split(string(roadmap), -1)
	var phases []phase
	for i, m := range re.FindAllStringSubmatch(string(roadmap), -1) {
		var n int
		fmt.Sscan(m[1], &n)
		p := phase{n: n, title: m[2], status: "planned"}
		p.steps = roadmapSteps(sections[i+1])
		if strings.Contains(string(decisions), fmt.Sprintf("Phase %d exit criterion met", n)) {
			p.status = "done"
		}
		cases, _ := filepath.Glob(fmt.Sprintf("testdata/phase%d/*/input.py", n))
		p.tests = len(cases)
		for _, c := range cases {
			p.cases = append(p.cases, filepath.Base(filepath.Dir(c)))
		}
		phases = append(phases, p)
	}
	for i := range phases {
		if phases[i].status == "planned" {
			phases[i].status = "next"
			break
		}
	}
	unsupported, _ := filepath.Glob("testdata/unsupported/*/input.py")
	total, done := 0, 0
	for _, p := range phases {
		total += p.tests
		if p.status == "done" {
			done++
		}
	}
	files := map[string]string{
		"docs/progress/progress.svg":       treemap(phases, rejectionNames(unsupported)),
		"docs/progress/badge-phases.svg":   badge("phases", fmt.Sprintf("%d / %d", done, len(phases)), "#3c9d4e"),
		"docs/progress/badge-tests.svg":    badge("differential tests", fmt.Sprint(total), "#3c7dd9"),
		"docs/progress/badge-rejected.svg": badge("rejection tests", fmt.Sprint(len(unsupported)), "#9a6fd1"),
	}
	for name, svg := range files {
		if err := os.WriteFile(name, []byte(svg), 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("%d of %d phases done, %d differential tests, %d rejection tests\n", done, len(phases), total, len(unsupported))
	return nil
}

// badge is a flat two-part label like the ones shields.io draws.
func badge(label, value, colour string) string {
	lw, vw := 8+len(label)*6, 8+len(value)*7
	w := lw + vw
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s: %s" font-family="Verdana,DejaVu Sans,sans-serif" font-size="11">
<rect width="%d" height="20" rx="3" fill="#555"/><rect x="%d" width="%d" height="20" rx="3" fill="%s"/><rect x="%d" width="4" height="20" fill="%s"/>
<text x="%d" y="14" fill="#fff" text-anchor="middle">%s</text><text x="%d" y="14" fill="#fff" text-anchor="middle">%s</text>
</svg>
`, w, html.EscapeString(label), html.EscapeString(value), w, lw, vw, colour, lw, colour, lw/2, html.EscapeString(label), lw+vw/2, html.EscapeString(value))
}

// roadmapSteps lists the numbered steps of a phase section, each shortened to
// its first sentence.
func roadmapSteps(section string) []string {
	var steps []string
	re := regexp.MustCompile(`^\d+\. (.+)$`)
	for _, line := range strings.Split(section, "\n") {
		if m := re.FindStringSubmatch(line); m != nil {
			s := m[1]
			if i := strings.Index(s, ". "); i > 0 {
				s = s[:i]
			}
			if len(s) > 110 {
				s = s[:107] + "..."
			}
			steps = append(steps, s)
		}
	}
	return steps
}

// rejectionNames are the names of the cases that pin down what is rejected.
func rejectionNames(inputs []string) []string {
	var names []string
	for _, in := range inputs {
		names = append(names, filepath.Base(filepath.Dir(in)))
	}
	sort.Strings(names)
	return names
}

// ---- the treemap ----

type rect struct{ x, y, w, h float64 }

type item struct {
	label string
	tip   string
	value float64
	fill  string
	cells []cell
}

type cell struct {
	tip  string
	fill string
}

const (
	green = "#2ea043"
	amber = "#d29922"
	grey  = "#6e7681"
	dark  = "#0d1117"
)

// squarify lays items out in a rectangle so that their areas follow their
// values and the boxes stay close to squares (Bruls, Huizing and van Wijk).
func squarify(items []item, r rect) []rect {
	total := 0.0
	for _, it := range items {
		total += it.value
	}
	out := make([]rect, len(items))
	idx := make([]int, len(items))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool { return items[idx[a]].value > items[idx[b]].value })
	scale := r.w * r.h / total
	var row []int
	worst := func(row []int, side float64) float64 {
		sum, mn, mx := 0.0, math.MaxFloat64, 0.0
		for _, i := range row {
			a := items[i].value * scale
			sum += a
			mn, mx = math.Min(mn, a), math.Max(mx, a)
		}
		return math.Max(side*side*mx/(sum*sum), sum*sum/(side*side*mn))
	}
	place := func(row []int) {
		sum := 0.0
		for _, i := range row {
			sum += items[i].value * scale
		}
		if r.w >= r.h { // a column on the left
			cw := sum / r.h
			y := r.y
			for _, i := range row {
				h := items[i].value * scale / cw
				out[i] = rect{r.x, y, cw, h}
				y += h
			}
			r = rect{r.x + cw, r.y, r.w - cw, r.h}
		} else { // a row on top
			rh := sum / r.w
			x := r.x
			for _, i := range row {
				w := items[i].value * scale / rh
				out[i] = rect{x, r.y, w, rh}
				x += w
			}
			r = rect{r.x, r.y + rh, r.w, r.h - rh}
		}
	}
	for _, i := range idx {
		side := math.Min(r.w, r.h)
		if len(row) == 0 || worst(append(append([]int{}, row...), i), side) <= worst(row, side) {
			row = append(row, i)
			continue
		}
		place(row)
		row = []int{i}
	}
	if len(row) > 0 {
		place(row)
	}
	return out
}

// panel draws one treemap: a title bar, then a box per group with a small
// square per cell.
func panel(b *strings.Builder, title string, items []item, r rect) {
	const bar = 26
	fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%d" fill="%s"/>`+"\n", r.x, r.y, r.w, bar, dark)
	fmt.Fprintf(b, `<text x="%.1f" y="%.1f" font-size="13" fill="#e6edf3">%s</text>`+"\n", r.x+6, r.y+17, html.EscapeString(title))
	body := rect{r.x, r.y + bar, r.w, r.h - bar}
	fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`+"\n", body.x, body.y, body.w, body.h, dark)
	for i, box := range squarify(items, body) {
		drawGroup(b, items[i], box)
	}
}

func drawGroup(b *strings.Builder, it item, r rect) {
	const inset, head = 1.5, 14
	fmt.Fprintf(b, `<g><title>%s</title>`, html.EscapeString(it.tip))
	fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#161b22" stroke="%s" stroke-width="1.5"/>`, r.x+inset/2, r.y+inset/2, r.w-inset, r.h-inset, dark)
	inner := rect{r.x + inset + 1, r.y + inset + 1, r.w - 2*inset - 2, r.h - 2*inset - 2}
	showLabel := r.w > 44 && r.h > 30
	if showLabel {
		inner.y += head
		inner.h -= head
	}
	n := len(it.cells)
	if n > 0 && inner.w > 4 && inner.h > 4 {
		cols := int(math.Ceil(math.Sqrt(float64(n) * inner.w / inner.h)))
		cols = max(cols, 1)
		rows := (n + cols - 1) / cols
		cw, ch := inner.w/float64(cols), inner.h/float64(rows)
		for k, c := range it.cells {
			x, y := inner.x+float64(k%cols)*cw, inner.y+float64(k/cols)*ch
			fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="%s" stroke-width="0.8"><title>%s</title></rect>`,
				x, y, cw, ch, c.fill, dark, html.EscapeString(c.tip))
		}
	}
	if showLabel {
		fmt.Fprintf(b, `<text x="%.1f" y="%.1f" font-size="10.5" fill="#e6edf3" font-weight="bold">%s</text>`, r.x+inset+4, r.y+inset+11, html.EscapeString(truncate(it.label, int(r.w/6.2))))
	}
	b.WriteString("</g>\n")
}

func truncate(s string, n int) string {
	if n < 3 || len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

// treemap draws two panels. The left has a square for every step of the
// roadmap, grouped by phase and green when the phase is done. The right has a
// square for every test: green for each differential test that passes, grey
// for each construct that is rejected on purpose.
func treemap(phases []phase, rejected []string) string {
	const width, height, gap = 1100.0, 400.0, 8.0
	var left, right []item
	steps, done := 0, 0
	for _, p := range phases {
		fill := map[string]string{"done": green, "next": amber, "planned": grey}[p.status]
		it := item{label: fmt.Sprintf("Phase %d", p.n), tip: fmt.Sprintf("Phase %d: %s (%s)", p.n, p.title, p.status), value: float64(len(p.steps))}
		for _, s := range p.steps {
			it.cells = append(it.cells, cell{tip: fmt.Sprintf("Phase %d: %s", p.n, s), fill: fill})
		}
		steps += len(p.steps)
		if p.status == "done" {
			done += len(p.steps)
		}
		left = append(left, it)

		if len(p.cases) > 0 {
			ti := item{label: fmt.Sprintf("Phase %d", p.n), tip: fmt.Sprintf("Phase %d: %d differential tests", p.n, len(p.cases)), value: float64(len(p.cases))}
			for _, c := range p.cases {
				ti.cells = append(ti.cells, cell{tip: fmt.Sprintf("phase %d: %s", p.n, c), fill: green})
			}
			right = append(right, ti)
		}
	}
	passing := 0
	for _, it := range right {
		passing += len(it.cells)
	}
	// The rejected constructs are grouped by the first word of their name.
	groups := map[string]*item{}
	var order []string
	for _, name := range rejected {
		key, _, _ := strings.Cut(name, "_")
		g := groups[key]
		if g == nil {
			g = &item{label: key, tip: "rejected: " + key + "…"}
			groups[key] = g
			order = append(order, key)
		}
		g.cells = append(g.cells, cell{tip: "rejected on purpose: " + name, fill: grey})
		g.value++
	}
	sort.Strings(order)
	for _, k := range order {
		right = append(right, *groups[k])
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" font-family="Verdana,DejaVu Sans,sans-serif">`+"\n", width, height, width, height)
	fmt.Fprintf(&b, `<rect width="%.0f" height="%.0f" fill="%s"/>`+"\n", width, height, dark)
	lw := (width - gap) * 0.52
	pct := 100 * float64(done) / float64(max(steps, 1))
	panel(&b, fmt.Sprintf("Roadmap steps: %.1f%% (%d/%d)", pct, done, steps), left, rect{0, 0, lw, height})
	panel(&b, fmt.Sprintf("Tests: %d differential, %d rejections", passing, len(rejected)), right, rect{lw + gap, 0, width - lw - gap, height})
	b.WriteString("</svg>\n")
	return b.String()
}
