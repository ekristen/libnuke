package log

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

type CustomFormatter struct {
	FallbackFormatter logrus.Formatter
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) { //nolint:gocyclo
	if f.FallbackFormatter == nil {
		f.FallbackFormatter = &logrus.TextFormatter{}
	}

	if entry == nil {
		return nil, nil
	}

	resourceType, ok := entry.Data["type"].(string)
	if !ok {
		return f.FallbackFormatter.Format(entry)
	}

	if _, ok := entry.Data["owner"]; !ok {
		return f.FallbackFormatter.Format(entry)
	}
	if _, ok := entry.Data["state"]; !ok {
		return f.FallbackFormatter.Format(entry)
	}
	if _, ok := entry.Data["name"]; !ok {
		return f.FallbackFormatter.Format(entry)
	}

	owner := entry.Data["owner"].(string)
	resourceName := entry.Data["name"].(string)
	state := entry.Data["state"].(int)

	var sortedFields = make([]string, 0)
	for k, v := range entry.Data {
		if strings.HasPrefix(k, "prop:") {
			if strings.HasPrefix(k, "prop:_") {
				continue
			}

			sortedFields = append(sortedFields, fmt.Sprintf("%s: %q", k[5:], v))
		}
	}

	sort.Strings(sortedFields)

	msgColor := ReasonSuccess
	switch state {
	case 0, 1, 8:
		msgColor = ReasonSuccess
	case 2:
		msgColor = ReasonHold
	case 3:
		msgColor = ReasonRemoveTriggered
	case 4:
		msgColor = ReasonWaitDependency
	case 5:
		msgColor = ReasonWaitPending
	case 6:
		msgColor = ReasonError
	case 7:
		msgColor = ReasonSkip
	}

	msg := fmt.Sprintf("%s - %s - %s - %s - %s\n",
		ColorRegion.Sprint(owner),
		ColorResourceType.Sprint(resourceType),
		ColorResourceID.Sprint(resourceName),
		ColorResourceProperties.Sprintf("[%s]", strings.Join(sortedFields, ", ")),
		msgColor.Sprint(entry.Message))

	return []byte(msg), nil
}
