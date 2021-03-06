package helpers

import (
	"fmt"
	"testing"
)

func TestCreateNGrams(t *testing.T) {
	result := CreateNgrams("こんにちは", 2)
	expected := []string{"こん", "んに", "にち", "ちは"}
	fmt.Printf("%v\n", result)
	for i, r := range expected {
		if r != result[i] {
			t.Fail()
		}
	}
}

func TestNoDup(t *testing.T) {
	result := CreateNgrams("むむ", 1)
	if len(result) != 1 {
		t.Fail()
	}
}
