package test

import (
	"github.com/camsiabor/qcom/util"
	"testing"
)

func TestTry(t *testing.T) {

}

func BenchmarkTry(b *testing.B) {
	var i interface{}
	i = 100.5

	for n := 1; n <= 1000; n++ {

		var p = util.AsInt(i, 0)

		p = p + 1

	}
}
