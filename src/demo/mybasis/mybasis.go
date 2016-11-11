package mybasis

import (
	"errors"
	"fmt"
)

func TestVaribale() {
	myVariableName()
	myConstName()
	nzjclx()
	errorType()
	testIota()
	testArray()
	testSlice()
	testMap()
}

// 变量的定义(Go对于已声明但未使用的变量会在编译阶段报错)
func myVariableName() {

	var variableName string
	fmt.Println(variableName)

	var vname1, vname2, vname3 string
	fmt.Printf("vname1 = %v\nvname2 = %v\nvname3 = %v\n\n", vname1, vname2, vname3)

	var vname4, vname5, vname6 string = "4", "5", "6"
	fmt.Printf("vname4 = %v\nvname5 = %v\nvname6 = %v\n\n", vname4, vname5, vname6)

	var vname11, vname22, vname33 = "11", "22", "33"
	fmt.Printf("vname11 = %v\nvname22 = %v\nvname33 = %v\n\n", vname11, vname22, vname33)

	vname01, vname02, vname03 := "01", "02", "03"
	fmt.Printf("vname01 = %v\nvname02 = %v\nvname03 = %v\n\n", vname01, vname02, vname03)

	// _（下划线）是个特殊的变量名，任何赋予它的值都会被丢弃。在这个例子中，我们将值35赋予b，并同时丢弃34：
	_, b := 34, 35
	fmt.Println(b)
}

// 常量
func myConstName() {

	const constname = "123"
	const pi float32 = 3.1415926
}

// 内置基础类型
func nzjclx() {

	// Boolean

	// 在Go中，布尔值的类型为bool，值是true或false，默认为false。
	var available bool
	available = true
	valid := false
	fmt.Printf("available = %v\nvalid = %v\n\n", available, valid)

	// 数值类型

	/* 整数类型有无符号和带符号两种。
	Go同时支持int和uint，这两种类型的长度相同，但具体长度取决于不同编译器的实现。
	Go里面也有直接定义好位数的类型：rune, int8, int16, int32, int64和byte, uint8, uint16, uint32, uint64。
	其中rune是int32的别称，byte是uint8的别称。这些类型的变量之间不允许互相赋值或操作，不然会在编译时引起编译器报错。*/

	// 浮点数的类型有float32和float64两种（没有float类型），默认是float64。

	/* Go还支持复数。它的默认类型是complex128（64位实数+64位虚数）。如果需要小一些的，也有complex64(32位实数+32位虚数)。
	复数的形式为RE + IMi，其中RE是实数部分，IM是虚数部分，而最后的i是虚数单位。*/
	var c complex64 = 5 + 5i
	fmt.Println(c)

	// 字符串(在Go中字符串是不可变的)

	var frenchHello string
	frenchHello = "test"
	fmt.Println(frenchHello)

	var emptyString string = ""
	fmt.Println(emptyString)

	no, yes, maybe := "no", "yes", "maybe"
	fmt.Printf("no = %v\nyes = %v\nmaybe = %v\n\n", no, yes, maybe)

	// 修改字符串1
	s := "hello"
	cc := []byte(s) // 将字符串 s 转换为 []byte 类型
	cc[0] = 'c'
	s2 := string(cc) // 再转换回 string 类型
	fmt.Println(s2)

	// 连接2个字符串
	a := "hello"
	b := " world"
	s3 := a + b
	fmt.Println(s3)

	// 修改字符串2
	as := "hello"
	as = "c" + as[1:] // 字符串虽不能更改，但可进行切片操作
	fmt.Println(as)

	// 声明多行字符串 `
	mm := `hello
		world`
	fmt.Println(mm)
}

// 错误类型
func errorType() {
	err := errors.New("emit macho dwarf: elf header corrupted\n\n")
	if err != nil {
		fmt.Print(err)
	}
}

// iota枚举
func testIota() {
	const (
		x = iota // x == 0
		y = iota // y == 1
		z = iota // z == 2
		w        // 常量声明省略值时，默认和之前一个值的字面相同。这里隐式地说w = iota，因此w == 3。其实上面y和z可同样不用"= iota"
	)

	const v = iota // 每遇到一个const关键字，iota就会重置，此时v == 0

	const (
		e, f, g = iota, iota, iota //e=0,f=0,g=0 iota在同一行值相同
	)

	const (
		a  = iota // a = 0
		b  = "B"
		c  = iota // c = 2
		gg        // gg = 3
	)
}

// 数组
func testArray() {

	var arr [5]int
	arr[0] = 11
	arr[2] = 22
	fmt.Println(arr)

	a := [3]int{1, 2, 3}
	b := [10]int{1, 2, 3}
	c := [...]int{4, 5, 6}
	fmt.Printf("a = %v\nb = %v\nc = %v\n\n", a, b, c)

	doubleArray := [2][4]int{[4]int{1, 2, 3, 4}, [4]int{5, 6, 7, 8}}
	easyArray := [2][4]int{{1, 2, 3, 4}, {5, 6, 7, 8}}
	fmt.Printf("doubleArray = %v\neasyArray = %v\n\n", doubleArray, easyArray)
}

// slice
// slice并不是真正意义上的动态数组，而是一个引用类型。slice总是指向一个底层array，slice的声明也可以像array一样，只是不需要长度。
func testSlice() {
	var fslice []int // 和声明array一样，只是少了长度
	slice := []byte{'a', 'b', 'c', 'd'}
	var ar = [10]byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'} // 数组
	var a, b []byte
	a = ar[2:5]
	b = ar[3:5]
	fmt.Printf("fslice = %v\nslice = %v\nar = %v\na = %v\nb = %v\n\n", fslice, slice, ar, a, b)

	/*
		slice是引用类型，所以当引用改变其中元素的值时，其它的所有引用都会改变该值
		len 获取slice的长度
		cap 获取slice的最大容量
		append 向slice里面追加一个或者多个元素，然后返回一个和slice一样类型的slice
		copy 函数copy从源slice的src中复制元素到目标dst，并且返回复制的元素的个数
		slice的index只能是int类型
	*/
}

// map 也就是Python中字典的概念，它的格式为map[keyType]valueType
func testMap() {

	// 声明一个key是字符串，值为int的字典,这种方式的声明需要在使用之前使用make初始化
	var numbers map[string]int
	fmt.Printf("numbers = %v\n", numbers)

	num := make(map[string]int)
	num["one"] = 1
	fmt.Printf("num = %v\n", num["one"])

	/*
	   map是无序的，每次打印出来的map都会不一样，它不能通过index获取，而必须通过key获取
	   map的长度是不固定的，也就是和slice一样，也是一种引用类型
	   内置的len函数同样适用于map，返回map拥有的key的数量
	   map的值可以很方便的修改，通过num["one"]=11可以很容易的把key为one的字典值改为11
	   map和其他基本型别不同，它不是thread-safe，在多个go-routine存取时，必须使用mutex lock机制
	*/

	// 删除
	rating := map[string]float32{"C": 5, "Go": 4.5, "Python": 4.5, "C++": 2}
	csharpRating, ok := rating["C#"]
	if ok {
		fmt.Println("C# is in the map and its rating is ", csharpRating)
	} else {
		fmt.Println("We have no rating associated with C# in the map")
	}

	delete(rating, "C")
	fmt.Println(rating)
}

// make new 操作
// make用于内建类型（map、slice 和channel）的内存分配。new用于各种类型的内存分配。
// new返回指针。
// make返回初始化后的（非零）值。

/* 关于“零值”，所指并非是空值，而是一种“变量未填充前”的默认值，通常为0。 此处罗列 部分类型 的 “零值”
int     0
int8    0
int32   0
int64   0
uint    0x0
rune    0 //rune的实际类型是 int32
byte    0x0 // byte的实际类型是 uint8
float32 0 //长度为 4 byte
float64 0 //长度为 8 byte
bool    false
string  ""
*/
