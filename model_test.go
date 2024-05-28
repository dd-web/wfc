package wfc

import "testing"

func TestLoadImage(t *testing.T) {
	// _ = GenModelFromInputImage("input/SmallLines.jpg")
	err := NewSampleImage("coolimg", 32)
	if err != nil {
		panic(err)
	}
}
