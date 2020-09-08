package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestParserAndTransformer(t *testing.T) {
	f, err := os.Open("test_data/record_type.dhall")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	bf := bufio.NewReader(f)

	rt, err := parseRecordType(bf)
	if err != nil {
		t.Fatal(err)
	}

	transformRecordType(rt)

	var sb strings.Builder

	rt.ToDhall(&sb, 1)

	fmt.Println(sb.String())
}
