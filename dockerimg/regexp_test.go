package dockerimg

import (
	"strings"
	"testing"
)

func TestReferenceRegexp(t *testing.T) {
	sr := strings.NewReader("image: index.docker.io/sourcegraph/gitserver:insiders@sha256:a8bbb0e7ba41b812166d5df154d270801716a309fc2ff08132dcfc1c6e61d4c0")

	var imgRefs []*ImageReference

	err := processReader(sr, &imgRefs, make(map[string]struct{}))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(imgRefs)
}
