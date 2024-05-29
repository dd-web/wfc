package wfc

import (
	"fmt"
	"testing"
)

func TestLoadImage(t *testing.T) {

	// err := NewSampleImage("coolimg", 512)
	// if err != nil {
	// 	panic(err)
	// }

	model, err := NewWFModelSet("input/BlackAndWhiteZigZags.png", 15, 15)
	if err != nil {
		panic(err)
	}

	fmt.Printf("model created.\n")

	if err = model.Save(); err != nil {
		panic(err)
	}
}
