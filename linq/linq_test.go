package linq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Student struct {
	Name string
	Age  int
	Sex  int
}

var students = []Student{
	{"张三", 28, 0},
	{"李四", 29, 1},
	{"王五", 20, 1},
	{"赵六", 17, 1},
	{"孙七", 22, 0},
	{"周八", 23, 0},
	{"吴九", 24, 1},
	{"郑十", 25, 1},
	{"王十一", 26, 0},
}

func TestAll(t *testing.T) {
	arr := From(students).Where(func(i Student) bool {
		return i.Age > 25
	}).Where(func(i Student) bool {
		return i.Sex == 1
	}).Sort(func(a, b Student) bool {
		return a.Age < b.Age
	}).Results()
	assert.Equal(t, []Student{{"李四", 29, 1}}, arr)
}
