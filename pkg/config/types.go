package config

import (
	"github.com/ekristen/libnuke/pkg/filter"
	"github.com/ekristen/libnuke/pkg/types"
	"github.com/sirupsen/logrus"
)

// Account is a collection of filters and resource types that are to be included or excluded from the nuke process.
// While the word Account is used, it is not limited to AWS accounts. It can be used for any type of grouping of
// resources. For example, you could have an account for your AWS account, another for your GCP project, and another
// for your Azure tenant. It's tool implementation dependent.
type Account struct {
	// Filters is a collection of filters that are to be included during the nuke process for the specific account.
	Filters filter.Filters `yaml:"filters"`

	// ResourceTypes is a collection of resource types that are to be included or excluded from the nuke process for
	// the specific account.
	ResourceTypes ResourceTypes `yaml:"resource-types"`

	// Presets is a list of presets that are to be used for the specific account configuration. The presets are
	// defined in the top level Presets field.
	Presets []string `yaml:"presets"`
}

// Preset is a collection of filters that are to be included during the nuke process.
type Preset struct {
	Filters filter.Filters `yaml:"filters"`
}

// ResourceTypes is a collection of resource types that are to be included or excluded from the nuke process. The
// Includes and Excludes fields are mutually exclusive. If a resource type is listed in both the Includes and Excludes
// fields then the Excludes field will take precedence. Additionally, the Alternatives field is a list of resource types
// that are to be used instead of the default resource. The primary use case for this is AWS Cloud Control API resources.
type ResourceTypes struct {
	// Includes is a list of resource types that are to be included during the nuke process. If a resource type is
	// listed in both the Includes and Excludes fields then the Excludes field will take precedence.
	Includes types.Collection `yaml:"includes"`

	// Excludes is a list of resource types that are to be excluded during the nuke process. If a resource type is
	// listed in both the Includes and Excludes fields then the Excludes field will take precedence.
	Excludes types.Collection `yaml:"excludes"`

	// Alternatives is a list of resource types that are to be used instead of the default resource. The primary use
	// case for this is AWS Cloud Control API resources. If a resource has been registered with the Cloud Control API
	// then we want to use that resource instead of the default resource. This is a Resource level alternative, not
	// a resource instance (i.e. all resources of this type will use the alternative resource, not just the resources
	// that are associated with the alternative resource).
	Alternatives types.Collection `yaml:"alternatives"`

	// Targets is a list of resource types that are to be included during the nuke process. If a resource type is
	// listed in both the Targets and Excludes fields then the Excludes field will take precedence.
	// Deprecated: Use Includes instead.
	Targets types.Collection `yaml:"targets"`

	// CloudControl is a list of resource types that are to be used with the Cloud Control API. This is a Resource
	// level alternative. This was left in place to make the transition to libnuke and ekristen/aws-nuke@v3 easier
	// for existing users.
	// Deprecated: Use Alternatives instead.
	CloudControl types.Collection `yaml:"cloud-control"`
}

// GetIncludes returns the combined list of includes and targets. This is left over from the AWS Nuke
// tool and is deprecated. It was left to make the transition to the library and ekristen/aws-nuke@v3 easier for
// existing users. This will be removed in 4.x of ekristen/aws-nuke.
func (r *ResourceTypes) GetIncludes() types.Collection {
	var combined types.Collection

	if r.Targets != nil {
		logrus.Warn("'targets' is deprecated. Please use 'includes' instead.")
		combined = combined.Union(r.Targets)
	}

	combined = combined.Union(r.Includes)
	return combined
}

// GetAlternatives returns the combined list of cloud control and alternatives. This is left over from the AWS Nuke
// tool and is deprecated. It was left to make the transition to the library and ekristen/aws-nuke@v3 easier for
// existing users. This will be removed in 4.x of ekristen/aws-nuke.
func (r *ResourceTypes) GetAlternatives() types.Collection {
	var combined types.Collection

	if r.CloudControl != nil {
		logrus.Warn("'cloud-control' is deprecated. Please use 'alternatives' instead.")
		combined = combined.Union(r.CloudControl)
	}

	combined = combined.Union(r.Alternatives)
	return combined
}
