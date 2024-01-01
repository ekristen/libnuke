package resource

import (
	"github.com/ekristen/cloud-nuke-sdk/pkg/featureflag"
	"github.com/ekristen/cloud-nuke-sdk/pkg/types"
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
