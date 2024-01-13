package featureflag

// Original Source https://github.com/kubernetes/kops/v1.28.2/pkg/featureflag/featureflag.go

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// FeatureFlag defines a feature flag
type FeatureFlag struct {
	Key          string
	enabled      *bool
	defaultValue *bool
}

// Enabled checks if the flag is enabled
func (f *FeatureFlag) Enabled() bool {
	if f.enabled != nil {
		return *f.enabled
	}
	if f.defaultValue != nil {
		return *f.defaultValue
	}
	return false
}

// FeatureFlags defines a list of feature flags
type FeatureFlags struct {
	flags      map[string]*FeatureFlag
	flagsMutex sync.Mutex
}

// New creates a new feature flag
func (ffc *FeatureFlags) New(key string, defaultValue, value *bool) *FeatureFlag {
	ffc.flagsMutex.Lock()
	defer ffc.flagsMutex.Unlock()

	if ffc.flags == nil {
		ffc.flags = make(map[string]*FeatureFlag)
	}

	fl := ffc.flags[key]
	if fl == nil {
		fl = &FeatureFlag{
			Key: key,
		}
		ffc.flags[key] = fl
	}

	if fl.defaultValue == nil {
		fl.defaultValue = defaultValue
	}

	if value == Bool(true) {
		fl.enabled = value
	}

	return fl
}

// ParseFlags responsible for parse out the feature flag usage
func (ffc *FeatureFlags) ParseFlags(f string) {
	ffc.flagsMutex.Lock()
	defer ffc.flagsMutex.Unlock()

	f = strings.TrimSpace(f)
	for _, s := range strings.Split(f, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		enabled := true
		var ff *FeatureFlag
		if s[0] == '+' || s[0] == '-' {
			ff = ffc.flags[s[1:]]
			if s[0] == '-' {
				enabled = false
			}
		} else {
			ff = ffc.flags[s]
		}
		if ff != nil {
			logrus.Debugf("FeatureFlag %q=%v", ff.Key, enabled)
			ff.enabled = &enabled
		} else {
			logrus.Debugf("Unknown FeatureFlag %q", s)
		}
	}
}

// Get returns given FeatureFlag.
func (ffc *FeatureFlags) Get(flagName string) (*FeatureFlag, error) {
	ffc.flagsMutex.Lock()
	defer ffc.flagsMutex.Unlock()

	flag, found := ffc.flags[flagName]
	if !found {
		return nil, fmt.Errorf("flag %s not found", flagName)
	}
	return flag, nil
}

// Bool returns a pointer to the boolean value
func Bool(b bool) *bool {
	return &b
}
