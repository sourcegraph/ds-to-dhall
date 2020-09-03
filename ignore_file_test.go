package main

import "testing"

func TestMatchIgnore(t *testing.T) {
	fixtures := []struct {
		pattern  string
		path     string
		expected bool
	}{
		{
			pattern:  "*foo.yaml",
			path:     "/Users/uwe/work/tarfoo.yaml",
			expected: true,
		},
		{
			pattern:  "foo.yaml",
			path:     "/Users/uwe/work/tarfoo.yaml",
			expected: false,
		},
		{
			pattern:  "foo.yaml",
			path:     "/Users/uwe/work/foo.yaml",
			expected: true,
		},
		{
			pattern:  "*foo.yaml",
			path:     "/Users/uwe/work/tarfoo.yaml",
			expected: true,
		},
		{
			pattern:  "tar*.yaml",
			path:     "/Users/uwe/work/tarfoo.yaml",
			expected: true,
		},
		{
			pattern:  "tar?.yaml",
			path:     "/Users/uwe/work/tarfoo.yaml",
			expected: false,
		},
		{
			pattern:  "wo*/*foo.yaml",
			path:     "/Users/uwe/work/tarfoo.yaml",
			expected: true,
		},
		{
			pattern:  "uwe/work/*foo.yaml",
			path:     "/Users/uwe/work/tarfoo.yaml",
			expected: true,
		},
		{
			pattern:  "gus/*foo.yaml",
			path:     "/Users/uwe/work/tarfoo.yaml",
			expected: false,
		},
		{
			pattern:  "*foo.*",
			path:     "/Users/uwe/work/tarfoo.yaml",
			expected: true,
		},
		{
			pattern:  "*foo.yaml",
			path:     "uwe/work/tarfoo.yaml",
			expected: true,
		},
		{
			pattern:  "foo.yaml",
			path:     "work/tarfoo.yaml",
			expected: false,
		},
		{
			pattern:  "foo.yaml",
			path:     "uwe/work/foo.yaml",
			expected: true,
		},
		{
			pattern:  "*foo.yaml",
			path:     "work/tarfoo.yaml",
			expected: true,
		},
		{
			pattern:  "tar*.yaml",
			path:     "Users/uwe/work/tarfoo.yaml",
			expected: true,
		},
		{
			pattern:  "tar?.yaml",
			path:     "Users/uwe/work/tarfoo.yaml",
			expected: false,
		},
		{
			pattern:  "wo*/*foo.yaml",
			path:     "tarfoo.yaml",
			expected: false,
		},
		{
			pattern:  "uwe/work/*foo.yaml",
			path:     "work/tarfoo.yaml",
			expected: false,
		},
		{
			pattern:  "gus/*foo.yaml",
			path:     "work/tarfoo.yaml",
			expected: false,
		},
		{
			pattern:  "*foo.*",
			path:     "work/tarfoo.yaml",
			expected: true,
		},
		{
			pattern:  "uwe/work",
			path:     "/Users/uwe/work",
			expected: true,
		},
	}

	for _, fx := range fixtures {
		ign, err := matchIgnore(fx.pattern, fx.path)
		if err != nil {
			t.Errorf("error matching %s against pattern %s: %v", fx.path, fx.pattern, err)
		}
		if ign != fx.expected {
			t.Errorf("extected %t, got %t matching %s against pattern %s", fx.expected, ign, fx.path, fx.pattern)
		}
	}
}
