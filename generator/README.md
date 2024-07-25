## Example

```go
package main

import (
	"fmt"
	"github.com/wsshow/op/generator"
)

type Result struct {
	Code int
	Desc string
	Data any
}

func main() {

	// 生成器函数
	genFunc := func(y generator.Yield[Result]) {
		for i := 0; i < 10; i++ {
			result := y.Yield(Result{Code: i, Desc: fmt.Sprintf("current-[%d]", i), Data: i * 2})
			fmt.Println("result:", result)
		}
	}

	// 创建生成器
	gen := generator.NewGenerator(genFunc)

	// 从生成器中获取值
	for {
		value, done := gen.Next()
		if done {
			break
		}
		// 处理Next传值
		if value.Code == 5 {
			fmt.Printf("Generated value: %+v\n", value)
			value, _ = gen.Next("gen next value after 5")
		}
		fmt.Printf("Generated value: %+v\n", value)
	}

	// 生成器结束后的状态
	fmt.Println(gen.Next())
	fmt.Println(gen.Next())
	fmt.Println(gen.Next())
	fmt.Println(gen.Next())
	fmt.Println(gen.Next())
	fmt.Println(gen.Next())
	fmt.Println(gen.Next())
	fmt.Println(gen.Next())

}
```

