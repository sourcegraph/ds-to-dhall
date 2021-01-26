package ds2dhall

import "testing"

func executeTestCommonRoot(pa, pb, expected string, t *testing.T) {
	c, err := commonPrefix([]string{pa, pb})
	if err != nil {
		t.Errorf("error while computing commong prefix for a = %s, b = %s: %v", pa, pb, err)
		return
	}

	if c != expected {
		t.Errorf("expected = %s, got = %s;     a = %s, b = %s", expected, c, pa, pb)
	}
}

func TestCommonPrefix(t *testing.T) {
	executeTestCommonRoot("/a/b/c/d/e/f", "/a/b/v", "/a/b", t)
	executeTestCommonRoot("/a/b/v", "/a/b/c/d/e/f", "/a/b", t)
	executeTestCommonRoot("/a", "/b", "/", t)
	executeTestCommonRoot("/a/b", "/a/b", "/a/b", t)
	executeTestCommonRoot("/a/b/c", "/a/b/d", "/a/b", t)
	executeTestCommonRoot("/a", "/a", "/a", t)
	executeTestCommonRoot("/a/c", "/a/b", "/a", t)
	executeTestCommonRoot("/a/b/c", "/a/b", "/a/b", t)
	executeTestCommonRoot("/a/b", "/a/b/c", "/a/b", t)
	executeTestCommonRoot("/a/b/v/", "/a/b/v", "/a/b/v", t)
	executeTestCommonRoot("/a", "/", "/", t)
}
