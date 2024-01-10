package resource

import (
	"github.com/ekristen/libnuke/pkg/featureflag"
	"github.com/ekristen/libnuke/pkg/types"
)

type Resource interface {
	Remove() error
}

type Filter interface {
	Resource
	Filter() error
}

type LegacyStringer interface {
	Resource
	String() string
}

type PropertyGetter interface {
	Resource
	Properties() types.Properties
}

type FeatureFlagGetter interface {
	Resource
	FeatureFlags(*featureflag.FeatureFlags)
}
