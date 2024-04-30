package log

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	resourceType, ok := entry.Data["type"].(string)
	if !ok {
		return nil, nil
	}

	if _, ok := entry.Data["owner"]; !ok {
		return nil, nil
	}
	if _, ok := entry.Data["resource"]; !ok {
		return nil, nil
	}
	if _, ok := entry.Data["state"]; !ok {
		return nil, nil
	}

	owner := entry.Data["owner"].(string)
	resource := entry.Data["resource"].(string)
	state := entry.Data["state"].(int)

	var sortedFields = make([]string, 0)
	for k, v := range entry.Data {
		if strings.HasPrefix(k, "prop:") {
			sortedFields = append(sortedFields, fmt.Sprintf("%s: %q", k[5:], v))
		}
	}

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
		ColorRegion.Sprintf(owner),
		ColorResourceType.Sprintf(resourceType),
		ColorResourceID.Sprintf(resource),
		ColorResourceProperties.Sprintf("[%s]", strings.Join(sortedFields, ", ")),
		msgColor.Sprintf(entry.Message))

	return []byte(msg), nil
}
