// Package log provides a way to log messages to the screen with the appropriate coloring and formatting for readability
package log

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
)

var (
	ReasonSkip            = *color.New(color.FgYellow)
	ReasonError           = *color.New(color.FgRed)
	ReasonRemoveTriggered = *color.New(color.FgGreen)
	ReasonWaitPending     = *color.New(color.FgBlue)
	ReasonWaitDependency  = *color.New(color.FgCyan)
	ReasonSuccess         = *color.New(color.FgGreen)
	ReasonHold            = *color.New(color.FgMagenta)
)

var (
	ColorRegion             = *color.New(color.Bold)
	ColorResourceType       = *color.New()
	ColorResourceID         = *color.New(color.Bold)
	ColorResourceProperties = *color.New(color.Italic)
)

// Sorted -- Format the resource properties in sorted order ready for printing.
// This ensures that multiple runs of aws-nuke produce stable output so
// that they can be compared with each other.
func Sorted(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		if strings.HasPrefix(k, "_") {
			continue
		}

		keys = append(keys, k)
	}
	sort.Strings(keys)
	sorted := make([]string, 0, len(m))
	for k := range keys {
		sorted = append(sorted, fmt.Sprintf("%s: %q", keys[k], m[keys[k]]))
	}
	return fmt.Sprintf("[%s]", strings.Join(sorted, ", "))
}
