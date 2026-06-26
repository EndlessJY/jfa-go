package linecache

import (
	"fmt"
	"testing"
)

func TestLineCacheKeepsLastNLines(t *testing.T) {
	wr := NewLineCache(3)
	for _, line := range []string{"one", "two", "three", "four"} {
		fmt.Fprintln(wr, line)
	}

	if got, want := wr.String(), "two\nthree\nfour\n"; got != want {
		t.Fatalf("unexpected cache contents:\nwant %q\ngot  %q", want, got)
	}
}

func TestLineCacheAcceptsMultipleLinesPerWrite(t *testing.T) {
	wr := NewLineCache(4)
	n, err := wr.Write([]byte("one\ntwo\nthree\n"))
	if err != nil {
		t.Fatal(err)
	}
	if n != len("one\ntwo\nthree\n") {
		t.Fatalf("unexpected write count: %d", n)
	}

	if got, want := wr.String(), "one\ntwo\nthree\n"; got != want {
		t.Fatalf("unexpected cache contents:\nwant %q\ngot  %q", want, got)
	}
}
