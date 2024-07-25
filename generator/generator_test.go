package generator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Result struct {
	Code int
	Desc string
	Data any
}

func TestNextResult(t *testing.T) {
	genFunc := func(y Yield[Result]) {
		for i := 0; i < 10; i++ {
			result := y.Yield(Result{Code: i, Desc: fmt.Sprintf("current-[%d]", i), Data: i * 2})
			if i == 5 {
				assert.Equal(t, "test", result)
			}
		}
	}

	// 创建生成器
	gen := NewGenerator(genFunc)

	for {
		value, done := gen.Next()
		if done {
			break
		}
		if value.Code == 5 {
			value, _ = gen.Next("test")
		}
	}
}
