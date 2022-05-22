package main

import (
	"math"
	"testing"
)

func TestLatLonToDecimal(t *testing.T) {
	lat := latToDecimal("5052.8766")
	lon := lonToDecimal("00703.9644")

	if math.Abs(lat-50.881276667) > 0.000001 {
		t.Fatalf("bad conversion, want: %v, got: %v", 50.881276667, lat)
	}

	if math.Abs(lon-7.066073333) > 0.000001 {
		t.Fatalf("bad conversion, want: %v, got: %v", 7.066073333, lon)
	}
}
