package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/types"
)

func TestTypes_ResourceTypes(t *testing.T) {
	cases := []struct {
		name             string
		includes         []string
		targets          []string
		excludes         []string
		alternatives     []string
		cloudcontrol     []string
		wantIncludes     types.Collection
		wantAlternatives types.Collection
	}{
		{
			name:             "overlapping includes/targets",
			includes:         []string{"test"},
			targets:          []string{"test", "test2"},
			wantIncludes:     []string{"test", "test2"},
			wantAlternatives: nil,
		},
		{
			name:             "overlapping alternatives/cloudcontrol",
			alternatives:     []string{"test"},
			cloudcontrol:     []string{"test", "test2"},
			wantIncludes:     nil,
			wantAlternatives: []string{"test", "test2"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt := ResourceTypes{
				Includes:     tc.includes,
				Targets:      tc.targets,
				Excludes:     tc.excludes,
				Alternatives: tc.alternatives,
				CloudControl: tc.cloudcontrol,
			}

			gotIncludes := rt.GetIncludes()
			gotAlternatives := rt.GetAlternatives()

			assert.Equal(t, tc.wantIncludes, gotIncludes)
			assert.Equal(t, tc.wantAlternatives, gotAlternatives)
		})
	}
}
