package config

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	if flag.Lookup("test.v") != nil {
		logrus.SetOutput(io.Discard)
	}
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetReportCaller(true)
}

func TestNew(t *testing.T) {
	opts := Options{
		Path: "testdata/example.yaml",
	}
	c, err := New(opts)
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewWithLogger(t *testing.T) {
	opts := Options{
		Path: "testdata/example.yaml",
		Log:  logrus.WithField("component", "test"),
	}
	c, err := New(opts)
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewNonExistentConfig(t *testing.T) {
	opts := Options{
		Path: "testdata/nonexistent.yaml",
	}
	_, err := New(opts)
	assert.Error(t, err)
}

func TestBlocklistDeprecations(t *testing.T) {
	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			if strings.HasSuffix(e.Caller.File, "pkg/config/config.go") {
				return
			}

			switch e.Caller.Line {
			case 119:
				assert.Equal(t, "deprecated configuration key 'account-blacklist' - please use 'blocklist' instead", e.Message)
			case 125:
				assert.Equal(t, "deprecated configuration key 'account-blocklist' - please use 'blocklist' instead", e.Message)
			}
		},
	})
	defer logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))

	opts := Options{
		Path: "testdata/deprecated.yaml",
	}

	c, err := New(opts)

	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, 2, len(c.Blocklist))
}

func TestHasBlocklist(t *testing.T) {
	opts := Options{
		Path: "testdata/example.yaml",
	}
	c, err := New(opts)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.True(t, c.HasBlocklist())
}

func TestInBlocklist(t *testing.T) {
	opts := Options{
		Path: "testdata/example.yaml",
	}
	c, err := New(opts)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.True(t, c.InBlocklist("1234567890"))
	assert.False(t, c.InBlocklist("1234567890123"))
}

func TestValidateAccount(t *testing.T) {
	opts := Options{
		Path: "testdata/example.yaml",
	}
	c, err := New(opts)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.Error(t, c.ValidateAccount("12345678901234"))
	assert.Error(t, c.ValidateAccount("1234567890"))
	assert.NoError(t, c.ValidateAccount("555133742"))
}

func TestFilters(t *testing.T) {
	opts := Options{
		Path: "testdata/example.yaml",
	}
	c, err := New(opts)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	// Test an account that is configured
	filters, err := c.Filters("555133742")
	assert.NoError(t, err)
	assert.NotNil(t, filters)
	assert.Equal(t, 3, len(filters))

	// Test an account that is not configured
	filters, err = c.Filters("1234567890")
	assert.Error(t, err)
	assert.Nil(t, filters)
}

func TestResourceTypeDeprecations(t *testing.T) {
	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			fmt.Println("here", e.Caller.File, e.Caller.Line)
			if strings.HasSuffix(e.Caller.File, "pkg/config/config.go") {
				return
			}

			if e.Caller.Line == 212 {
				assert.Equal(t, "deprecated resource type 'IamRole' - converting to 'IAMRole'", e.Message)
			}
		},
	})
	defer logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))

	opts := Options{
		Path:                  "testdata/deprecated-resources.yaml",
		Deprecations:          map[string]string{"IamRole": "IAMRole"},
		NoResolveDeprecations: true,
	}
	c, err := New(opts)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	err = c.ResolveDeprecations()
	assert.NoError(t, err)
}

func TestResourceTypeDeprecationsNoFilters(t *testing.T) {
	opts := Options{
		Path:                  "testdata/deprecated-resources-no-filters.yaml",
		Deprecations:          map[string]string{"IamRole": "IAMRole"},
		NoResolveDeprecations: true,
	}
	c, err := New(opts)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	err = c.ResolveDeprecations()
	assert.NoError(t, err)
}

func TestResourceTypeDeprecationsError(t *testing.T) {
	logrus.AddHook(&TestGlobalHook{
		t: t,
		tf: func(t *testing.T, e *logrus.Entry) {
			fmt.Println("here", e.Caller.File, e.Caller.Line)
			if strings.HasSuffix(e.Caller.File, "pkg/config/config.go") {
				return
			}

			if e.Caller.Line == 212 {
				assert.Equal(t, "deprecated resource type 'IamRole' - converting to 'IAMRole'", e.Message)
			}
		},
	})
	defer logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))

	opts := Options{
		Path:                  "testdata/deprecated-resources-error.yaml",
		Deprecations:          map[string]string{"IamRole": "IAMRole"},
		NoResolveDeprecations: true,
	}
	c, err := New(opts)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	err = c.ResolveDeprecations()
	assert.Error(t, err)
}

func TestInvalid(t *testing.T) {
	opts := Options{
		Path: "testdata/invalid.yaml",
	}
	_, err := New(opts)
	assert.Error(t, err)
}

func TestInvalidPreset(t *testing.T) {
	opts := Options{
		Path: "testdata/invalid-preset.yaml",
	}
	c, err := New(opts)
	assert.NoError(t, err)

	_, err = c.Filters("555133742")
	assert.Error(t, err)
}

func TestNoBlocklist(t *testing.T) {
	opts := Options{
		Path: "testdata/no-blocklist.yaml",
	}
	c, err := New(opts)
	assert.NoError(t, err)
	assert.False(t, c.HasBlocklist())

	err = c.ValidateAccount("555133742")
	assert.Error(t, err)
}
