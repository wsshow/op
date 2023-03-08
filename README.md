# op

## array

| 方法                               | 描述                                                |
| :--------------------------------- | :-------------------------------------------------- |
| NewArray()                         | 创建一个新的 Array 对象                             |
| Add(elems ...interface{})          | 添加一个或多个元素到 Array 对象                     |
| Remove(e interface{})              | 从 Array 对象中删除指定的元素                       |
| RemoveAll(e interface{})           | 从 Array 对象中删除所有指定的元素                   |
| Contain(e interface{})             | 检查 Array 对象是否包含指定的元素                   |
| Count()                            | 返回 Array 对象中元素的数量                         |
| ForEach(f func(e interface{}))     | 对 Array 对象中的每个元素应用指定的函数             |
| Clear()                            | 清空 Array 对象中的所有元素                         |
| Data()                             | 返回 Array 对象中所有元素的切片                     |
| Sort(less func(i, j int) bool)     | 对 Array 对象中的元素进行排序                       |
| Filter(f func(e interface{}) bool) | 返回一个新的 Array 对象，其中包含符合指定条件的元素 |

## string

| 方法       | 描述                                       |
| :--------- | :----------------------------------------- |
| NewString  | 创建一个 String 对象并初始化字符串         |
| Contain    | 判断字符串是否包含指定的子串               |
| Index      | 返回指定子串在字符串中第一次出现的位置     |
| LastIndex  | 返回指定子串在字符串中最后一次出现的位置   |
| Split      | 返回字符串按照指定分隔符分割后的切片       |
| Length     | 返回字符串的长度                           |
| ReplaceAll | 替换字符串中所有指定的子串为新的字符串     |
| String     | 返回字符串对象的字符串表示                 |
| ToInt      | 将字符串转换为整型，如果无法转换则返回错误 |
| Concat     | 拼接多个字符串到当前字符串对象中           |

## queue

| 方法     | 描述                             |
| :------- | :------------------------------- |
| Enqueue  | 将一个或多个项添加到队列末尾。   |
| Dequeue  | 从队列的开头移除并返回项。       |
| Peek     | 返回队列开头的项，但不会移除它。 |
| Count    | 返回队列中项的数量。             |
| Contains | 检查队列中是否包含指定的项。     |
| ToSlice  | 将队列转换为切片。               |
| IsEmpty  | 检查队列是否为空。               |
| Clear    | 清空队列。                       |
