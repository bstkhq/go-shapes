package shapes

import (
	"fmt"
	"iter"
	"math/bits"
	"os"
	"time"
)

// Warning represents an operation warning. Warnings can be checked through
// Renderer.Warnings, and warning handlers can be configured through
// [Warnings.SetHandler]().
type Warning uint32

const (
	// operation failures
	WarnMissingSourceOpSkipped Warning = 1 << iota

	// spatial / geometric limits
	WarnRadiusClamped     // blur, glow, shadows, ...
	WarnNumSamplesClamped // blur, glow, ...
	WarnNotEnoughSamplesOpSkipped
	WarnThicknessClamped // strokes, morphological ops, ...
	WarnDistanceClamped  // jfa, ...

	// value integrity
	WarnNegativeValueZeroed
	WarnNegativeValueOpSkipped // vague but uncommon in practice
	WarnInvalidAlphaClamped    // out of [0, 1] range
	WarnInvalidTintClamped     // out of [0, 1] range
	WarnInvalidRateClamped     // out of [0, 1] range
	WarnInvalidBiasClamped     // out of [-1, 1] range
	WarnInvalidTextAlign
	WarnInvalidFlag
	WarnLowToleranceRaised // below 0.1
	WarnRepeatedFlag
	WarnInconsistentRangeOpSkipped
	WarnInconsistentRangeInvalidated

	// implementation-specific
	WarnInvalidNoiseSeedClamped
	WarnTooManyVertexAttribs
	WarnTooManyColorsClamped
	WarnInvalidApertureClamped

	warnSentinel
)

const unknownWarningPrefix = "unknown warning"

func (w Warning) Message() string {
	switch w {
	case WarnMissingSourceOpSkipped:
		return "missing source/mask image, operation skipped"
	case WarnRadiusClamped:
		return "radius exceeds max value, clamped"
	case WarnNumSamplesClamped:
		return "numSamples exceeds max value, clamped"
	case WarnNotEnoughSamplesOpSkipped:
		return "numSamples must be strictly positive, operation skipped"
	case WarnThicknessClamped:
		return "thickness exceeds max value, clamped"
	case WarnDistanceClamped:
		return "distance exceeds max value, clamped"
	case WarnNegativeValueZeroed:
		return "negative value not valid in context, clamped to zero"
	case WarnNegativeValueOpSkipped:
		return "negative value not valid in context, operation skipped"
	case WarnInvalidAlphaClamped:
		return "alpha value out of range, clamped"
	case WarnInvalidRateClamped:
		return "rate value out of range, clamped"
	case WarnInvalidBiasClamped:
		return "bias value out of [-1, 1] range, clamped"
	case WarnInvalidTintClamped:
		return "tint value out of [0, 1] range, clamped"
	case WarnInvalidTextAlign:
		return "invalid text align"
	case WarnInvalidFlag:
		return "invalid Flag in context"
	case WarnLowToleranceRaised:
		return "CircleShaderOptions.Tolerance != 0 but below minimum of 0.1, raised to 0.1"
	case WarnRepeatedFlag:
		return "redundant repeated Flag"
	case WarnInconsistentRangeOpSkipped:
		return "inconsistent range values (e.g. min/max, start/end, in/to/out), operation skipped"
	case WarnInconsistentRangeInvalidated:
		return "inconsistent range values (e.g. min/max, start/end, in/to/out), operation result invalidated"
	case WarnInvalidNoiseSeedClamped:
		return "noise seed out of [0, 1] range, clamped"
	case WarnTooManyVertexAttribs:
		return "too many vertex attributes, ignored excess ones"
	case WarnTooManyColorsClamped:
		return "too many color values, clamped to max"
	case WarnInvalidApertureClamped:
		return "aperture out of [0, 2*Pi], clamped"
	default:
		return fmt.Sprintf("%s 0x%08x", unknownWarningPrefix, w)
	}
}

func (r *Renderer) warnClampNonNegArgF32(value, maxValue float32, clampWarning Warning) float32 {
	if value > maxValue {
		r.Warnings.report(clampWarning, value)
		return maxValue
	} else if value < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, value)
		return 0
	}
	return value
}

// Warnings is a register of problems detected by the renderer during operations.
//
// Most users will only care about [Warnings.SetHandler]() to make warnings
// fatal during debug or log them in production/release builds.
type Warnings struct {
	reports Warning
	handler func(Warning, any, bool)
}

// NewWarningPanicHandler returns a handler for [Warnings.SetHandler]()
// that panics right away. Useful during debug to discover where the
// warnings are coming from.
func NewWarningPanicHandler() func(Warning, any, bool) {
	return func(warning Warning, value any, _ bool) {
		panic(fmt.Sprintf("%s (value=%v)", warning.Message(), value))
	}
}

// NewWarningLogOnceHandler returns a handler for [Warnings.SetHandler]()
// that logs warnings only the first time they are seen. Useful as a
// default handler on production/release builds.
func NewWarningLogOnceHandler() func(Warning, any, bool) {
	return func(warning Warning, value any, alreadySeen bool) {
		if !alreadySeen {
			ts := time.Now().Format("Mon Jan 2 15:04:05")
			fmt.Fprintf(os.Stderr, "[%s] WARNING: %s value=%v valtype=%T ctx=shapes.Renderer", warning.Message(), value, value, ts)
		}
	}
}

// SetHandler allows setting a custom handler for warnings.
// 'value' is the exact value that triggered the warning. Most often an int,
// float32 or float64 (but might also be nil).
//
// See [NewWarningPanicHandler]() and [NewWarningLogOnceHandler]() for reference
// implementations.
func (w *Warnings) SetHandler(handler func(warning Warning, value any, alreadySeen bool)) {
	w.handler = handler
}

func (w *Warnings) report(warning Warning, value any) {
	if w.handler != nil {
		w.handler(warning, value, w.Has(warning))
	}
	w.reports |= warning
}

func (w *Warnings) HasAny() bool {
	return w.reports != 0
}

func (w *Warnings) Has(warning Warning) bool {
	return w.reports&warning != 0
}

func (w *Warnings) Reset() {
	w.reports = 0
}

func (w *Warnings) All() iter.Seq[Warning] {
	return allWarnings(w.reports)
}

func noWarningIter(func(Warning) bool) {}

func allWarnings(warnings Warning) iter.Seq[Warning] {
	if warnings == 0 {
		return noWarningIter
	}

	return func(yield func(Warning) bool) {
		var w Warning = 1
		for warnings != 0 {
			zeros := bits.TrailingZeros32(uint32(warnings))
			w <<= zeros
			warnings >>= zeros

			if !yield(w) {
				return
			}
			w <<= 1
			warnings >>= 1
		}
	}
}
