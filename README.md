# op

## array
| 方法                                        | 描述                                                        |
| :------------------------------------------ | :---------------------------------------------------------- |
| NewArray() *Array                           | 创建一个新的Array对象                                       |
| Add(elems ...interface{})                   | 添加一个或多个元素到Array对象                               |
| Remove(e interface{})                       | 从Array对象中删除指定的元素                                 |
| RemoveAll(e interface{})                    | 从Array对象中删除所有指定的元素                             |
| Contain(e interface{}) bool                 | 检查Array对象是否包含指定的元素                             |
| Count() int                                 | 返回Array对象中元素的数量                                   |
| ForEach(f func(e interface{}))              | 对Array对象中的每个元素应用指定的函数                       |
| Clear()                                     | 清空Array对象中的所有元素                                   |
| Data()                                      | 返回Array对象中所有元素的切片                               |
| Sort(less func(i, j int) bool)              | 对Array对象中的元素进行排序                                 |
| Filter(f func(e interface{}) bool) *Array   | 返回一个新的Array对象，其中包含符合指定条件的元素           |
| Map(f func(interface{}) interface{}) *Array | 对Array对象中的每个元素应用指定的函数,并返回一个新Array对象 |

## string
| 方法                                | 描述                                       |
| :---------------------------------- | :----------------------------------------- |
| NewString(s string) *String         | 创建一个String对象并初始化字符串           |
| Contain(substr string) bool         | 判断字符串是否包含指定的子串               |
| Index(substr string) int            | 返回指定子串在字符串中第一次出现的位置     |
| LastIndex(substr string) int        | 返回指定子串在字符串中最后一次出现的位置   |
| Split(sep string) []string          | 返回字符串按照指定分隔符分割后的切片       |
| Length() int                        | 返回字符串的长度                           |
| ReplaceAll(old, new string) *String | 替换字符串中所有指定的子串为新的字符串     |
| String() string                     | 返回字符串对象的字符串表示                 |
| ToInt() (int, error)                | 将字符串转换为整型，如果无法转换则返回错误 |
| Concat(ss ...string) *String        | 拼接多个字符串到当前字符串对象中           |

## queue
| 方法                                        | 描述                                                        |
| :------------------------------------------ | :---------------------------------------------------------- |
| NewQueue() *Queue                           | 创建一个新的Queue对象                                       |
| Enqueue(items ...interface{})               | 将一个或多个项添加到队列末尾                                |
| Dequeue() interface{}                       | 从队列的开头移除并返回项                                    |
| Peek() interface{}                          | 返回队列开头的项，但不会移除它                              |
| Count() int                                 | 返回队列中项的数量                                          |
| Contains(item interface{}) bool             | 检查队列中是否包含指定的项                                  |
| ToSlice() []interface{}                     | 将队列转换为切片                                            |
| IsEmpty() bool                              | 检查队列是否为空                                            |
| Clear()                                     | 清空队列                                                    |
| ForEach(f func(interface{}))                | 对Queue对象中的每个元素应用指定的函数                       |
| Map(f func(interface{}) interface{}) *Queue | 对Queue对象中的每个元素应用指定的函数,并返回一个新Queue对象 |