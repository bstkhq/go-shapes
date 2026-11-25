package pixcheck

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// DumpDirEnv is the environment variable that makes [Report.Check] save the
// worst failure as PNGs, along with its [Metric.DiffImage]:
//
//	PIXCHECK_DUMP_DIR=/tmp/pixcheck go test -run ^TestName$ .
//
// Files are named after the test, and numbered when more than one report fails
// in the same test. Relative paths start from the directory of the package
// being tested.
const DumpDirEnv = "PIXCHECK_DUMP_DIR"

// dumps counts the failures saved by each test, to number the files.
var dumps struct {
	sync.Mutex
	count map[string]int
}

// Report is a helper for tests that collects and checks comparison results:
//
//	report := pixcheck.Report{Rule: pixcheck.Rule{MaxDelta: 1}}
//	for ... {
//		// draw got and want
//		report.Compare(label, got, want)
//	}
//	report.Check(t)
//
// The zero value is ready to use and checks for identical images. Comparisons
// that need different rules must use separate reports.
type Report struct {
	// Rule decides which comparisons pass. Set it only once on creation.
	Rule Rule

	// FormatPixel converts pixels to text when showing the worst failure, as
	// in [Format]. If unset, pixels are shown by their alpha values.
	FormatPixel func(color.RGBA) string

	results []Result
	worst   struct { // the worst failure, with its images
		Result
		got, want *image.RGBA
	}
}

// Result is the outcome of one comparison in a [Report].
type Result struct {
	Label string
	Diff  Diff
}

// Compare reads got and want, measures their difference based on Report.Rule,
// and adds the result to the report. It must be called inside [Run].
func (r *Report) Compare(label string, got, want *ebiten.Image) Diff {
	return r.CompareRGBA(label, Read(got), Read(want))
}

// CompareRGBA is like [Report.Compare] for images already in memory, e.g. to
// compare with expected images written with [Parse]. got and want must not be
// modified after being passed to this function.
func (r *Report) CompareRGBA(label string, got, want *image.RGBA) Diff {
	result := Result{label, r.Rule.Metric.Compare(got, want)}
	r.results = append(r.results, result)
	if !r.Rule.Accepts(result.Diff) && (r.worst.Result == Result{} || r.Rule.Worse(result.Diff, r.worst.Diff)) {
		r.worst.Result, r.worst.got, r.worst.want = result, got, want
	}
	return result.Diff
}

// Check fails t if any comparison doesn't pass the report's [Rule], or if there
// were no comparisons at all. Failures are listed in order, and the worst one
// is shown as text. It can also be saved as PNGs (see [DumpDirEnv]).
//
// If all comparisons pass, Check logs the largest delta found.
func (r *Report) Check(t testing.TB) {
	const MaxListed = 10
	const WindowSize = 15 // odd, so the worst pixel can be at the center

	t.Helper()
	if len(r.results) == 0 {
		t.Error("no comparisons were made")
		return
	}

	failed := r.failed()
	if len(failed) == 0 {
		t.Logf("all %d comparisons pass (%s)", len(r.results), r.largestPassing())
		return
	}

	var msg strings.Builder
	fmt.Fprintf(&msg, "%d of %d comparisons fail (rule: %s):", len(failed), len(r.results), r.Rule)
	for _, result := range failed[:min(len(failed), MaxListed)] {
		fmt.Fprintf(&msg, "\n  %s: %s", result.Label, result.Diff)
		if r.Rule.MaxDelta > 0 {
			fmt.Fprintf(&msg, ", %d over max delta", result.Diff.Over(r.Rule.MaxDelta))
		}
	}
	if len(failed) > MaxListed {
		fmt.Fprintf(&msg, "\n  ... and %d more", len(failed)-MaxListed)
	}
	formatPixel := r.FormatPixel
	if formatPixel == nil {
		formatPixel = formatAlphaPixel
	}
	worst := r.worst
	win := window(worst.got.Rect.Size(), worst.Diff.MaxAt, WindowSize)
	fmt.Fprintf(&msg, "\nworst failure (%s), pixels (%d, %d) to (%d, %d), max delta %d at (%d, %d):",
		worst.Label, win.Min.X, win.Min.Y, win.Max.X-1, win.Max.Y-1,
		worst.Diff.MaxDelta(), worst.Diff.MaxAt.X, worst.Diff.MaxAt.Y)
	gotWin := worst.got.SubImage(win.Add(worst.got.Rect.Min)).(*image.RGBA)
	wantWin := worst.want.SubImage(win.Add(worst.want.Rect.Min)).(*image.RGBA)
	indent := func(text string) string { // one level below "got:" and "want:"
		return "\n    " + strings.ReplaceAll(text, "\n", "\n    ")
	}
	fmt.Fprintf(&msg, "\n  got:%s", indent(Format(gotWin, formatPixel)))
	fmt.Fprintf(&msg, "\n  want:%s", indent(Format(wantWin, formatPixel)))
	if dir := os.Getenv(DumpDirEnv); dir == "" {
		fmt.Fprintf(&msg, "\nset %s to save the worst failure as PNGs", DumpDirEnv)
	} else if prefix, err := r.dumpWorst(dir, t.Name()); err != nil {
		fmt.Fprintf(&msg, "\nfailed to save the worst failure: %v", err)
	} else {
		fmt.Fprintf(&msg, "\nworst failure (%s) saved as %s_{got,want,diff}.png", worst.Label, prefix)
	}
	t.Error(msg.String())
}

// window returns an area of up to maxSize, centered as well as possible at
// the given centerAt for an area of the given size.
func window(size, centerAt image.Point, maxSize int) image.Rectangle {
	axis := func(length, centerAt int) (start, end int) {
		winLength := min(length, maxSize)
		start = max(0, min(centerAt-winLength/2, length-winLength))
		return start, start + winLength
	}
	x0, x1 := axis(size.X, centerAt.X)
	y0, y1 := axis(size.Y, centerAt.Y)
	return image.Rect(x0, y0, x1, y1)
}

// failed returns select(r.results, !r.Rule.Accepts)
func (r *Report) failed() []Result {
	var failed []Result
	for _, result := range r.results {
		if !r.Rule.Accepts(result.Diff) {
			failed = append(failed, result)
		}
	}
	return failed
}

// largestPassing describes the passing result with the largest delta.
func (r *Report) largestPassing() string {
	largest := r.results[0]
	for _, result := range r.results[1:] {
		if result.Diff.MaxDelta() > largest.Diff.MaxDelta() {
			largest = result
		}
	}
	if largest.Diff.MaxDelta() == 0 {
		return "all pixels match"
	}
	return fmt.Sprintf("largest delta in %s: %s", largest.Label, largest.Diff)
}

// dumpWorst saves the images of the worst failure and their deltas, and
// returns the common path prefix of the files.
func (r *Report) dumpWorst(dir, name string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	dumps.Lock()
	if dumps.count == nil {
		dumps.count = make(map[string]int)
	}
	dumps.count[name] += 1
	number := dumps.count[name]
	dumps.Unlock()

	prefix := filepath.Join(dir, strings.NewReplacer("/", "_", " ", "_").Replace(name))
	if number > 1 {
		prefix += fmt.Sprintf("_%d", number)
	}
	images := map[string]*image.RGBA{
		"got":  r.worst.got,
		"want": r.worst.want,
		"diff": r.Rule.Metric.DiffImage(r.worst.got, r.worst.want),
	}
	for suffix, img := range images {
		if err := SavePNG(prefix+"_"+suffix+".png", img); err != nil {
			return "", err
		}
	}
	return prefix, nil
}
