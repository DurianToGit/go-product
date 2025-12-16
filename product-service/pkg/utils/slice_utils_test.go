package utils

import (
	"fmt"
	"testing"
)

func TestSliceContains(t *testing.T) {
	var slice = []string{"a", "b", "c"}
	if SliceContains(slice, "a") != true {
		t.Error("SliceContains failed")
	}
}

func TestSliceRemove(t *testing.T) {
	var slice = []string{"a", "b", "c"}
	slice = SliceRemove(slice, "a")
	if SliceContains(slice, "a") == true {
		t.Error("SliceRemove failed")
	}
}

func TestSliceChunk(t *testing.T) {
	var slice = []string{"a", "b", "c", "d", "e", "f"}
	chunks := SliceChunk(slice, 2)
	fmt.Println(chunks)
	if len(chunks) != 3 {
		t.Error("SliceChunk failed")
	}
}
