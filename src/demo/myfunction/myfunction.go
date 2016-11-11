package myfunction

import (
	"fmt"
	"os"
)

func Testmyfunction() {
	testIf()
	testGoTo()
	testFor()
	testSwitch()
	testFunction()
	biancan(3, 4)
	testAdd1()
	testAdd2()
	testDefer()
	testZhiType()
}

// if
func testIf() {

	x := 15
	if x > 10 {
		fmt.Println("x is greater than 10")
	} else {
		fmt.Println("x is less than 10")
	}

	if x := computedValue(); x > 10 {
		fmt.Println("x is greater than 10")
	} else {
		fmt.Println("x is less than 10")
	}
}

func computedValue() (a int) {

	return 35
}

// goto 用goto跳转到必须在当前函数内定义的标签 标签名是大小写敏感的。
func testGoTo() {

	i := 0
Here: // 这行的第一个词，以冒号结束作为标签
	fmt.Println(i)
	i++
	if i < 55 {
		goto Here //跳转到Here去
	}
}

// for
func testFor() {

	sum := 0
	for index := 0; index < 10; index++ {
		sum += index
	}
	fmt.Println("sum is equal to ", sum)

	sum1 := 1
	for sum1 < 1000 {
		sum1 += sum1
	}
	fmt.Println("sum1 is equal to ", sum1)

	// break操作是跳出当前循环，continue是跳过本次循环。
	for index := 10; index > 0; index-- {
		if index == 5 {
			break // 或者continue
		}
		fmt.Println(index)
	}

	// for配合range可以用于读取slice和map的数据
	rating := map[string]float32{"C": 5, "Go": 4.5, "Python": 4.5, "C++": 2}
	for k, v := range rating {
		fmt.Println("map's key:", k)
		fmt.Println("map's val:", v)
	}

	// 对于“声明而未被调用”的变量, 编译器会报错, 在这种情况下, 可以使用_来丢弃不需要的返回值
	for _, v := range rating {
		fmt.Println("map's val:", v)
	}
}

// switch
func testSwitch() {

	i := 10
	switch i {
	case 1:
		fmt.Println("i is equal to 1")
	case 2, 3, 4:
		fmt.Println("i is equal to 2, 3 or 4")
	case 10:
		fmt.Println("i is equal to 10")
	default:
		fmt.Println("All I know is that i is an integer")
	}

	// Go里面switch默认相当于每个case最后带有break，匹配成功后不会自动向下执行其他case，而是跳出整个switch, 但是可以使用fallthrough强制执行后面的case代码。
	integer := 6
	switch integer {
	case 4:
		fmt.Println("The integer was <= 4")
		fallthrough
	case 5:
		fmt.Println("The integer was <= 5")
		fallthrough
	case 6:
		fmt.Println("The integer was <= 6")
		fallthrough
	case 7:
		fmt.Println("The integer was <= 7")
		fallthrough
	case 8:
		fmt.Println("The integer was <= 8")
		fallthrough
	default:
		fmt.Println("default case")
	}
}

// 函数
/*
关键字func用来声明一个函数funcName
函数可以有一个或者多个参数，每个参数后面带有类型，通过,分隔
函数可以返回多个值
上面返回值声明了两个变量output1和output2，如果你不想声明也可以，直接就两个类型
如果只有一个返回值且不声明返回值变量，那么你可以省略 包括返回值 的括号
如果没有返回值，那么就直接省略最后的返回信息
如果有返回值， 那么必须在函数的外层添加return语句
*/
func testFunction() {

	x := 3
	y := 4
	z := 5

	max_xy := max(x, y) //调用函数max(x, y)
	max_xz := max(x, z) //调用函数max(x, z)

	fmt.Printf("max(%d, %d) = %d\n", x, y, max_xy)
	fmt.Printf("max(%d, %d) = %d\n", x, z, max_xz)
	fmt.Printf("max(%d, %d) = %d\n", y, z, max(y, z)) // 也可在这直接调用它

	// Go语言比C更先进的特性，其中一点就是函数能够返回多个值。
	xx := 3
	yy := 4

	xPLUSy, xTIMESy := SumAndProduct(xx, yy)

	fmt.Printf("%d + %d = %d\n", xx, yy, xPLUSy)
	fmt.Printf("%d * %d = %d\n", xx, yy, xTIMESy)

}

// 返回a、b中最大值.
func max(a, b int) int {

	if a > b {
		return a
	}
	return b
}

//返回 A+B 和 A*B
func SumAndProduct(A, B int) (int, int) {

	return A + B, A * B
}

// 变参
func biancan(arg ...int) {

	// arg ...int告诉Go这个函数接受不定数量的参数。注意，这些参数的类型全部是int。在函数体中，变量arg是一个int的slice：
	for _, n := range arg {
		fmt.Printf("And the number is: %d\n", n)
	}
}

// 传值和传指针
/*
传指针使得多个函数能操作同一个对象。
传指针比较轻量级 (8bytes),只是传内存地址，我们可以用指针传递体积大的结构体。如果用参数值传递的话, 在每次copy上面就会花费相对较多的系统开销（内存和时间）。所以当你要传递大的结构体的时候，用指针是一个明智的选择。
Go语言中channel，slice，map这三种类型的实现机制类似指针，所以可以直接传递，而不用取地址后传递指针。（注：若函数需改变slice的长度，则仍需要取地址传递指针）
*/

func testAdd1() {

	x := 3
	fmt.Println("x = ", x)    // 应该输出 "x = 3"
	x1 := add1(x)             //调用add1(x)
	fmt.Println("x+1 = ", x1) // 应该输出"x+1 = 4"
	fmt.Println("x = ", x)    // 应该输出"x = 3"
}

func add1(a int) int {
	a = a + 1 // 我们改变了a的值
	return a  //返回一个新值
}

func testAdd2() {

	x := 3
	fmt.Println("x = ", x)    // 应该输出 "x = 3"
	x1 := add2(&x)            //调用add1(x)
	fmt.Println("x+1 = ", x1) // 应该输出"x+1 = 4"
	fmt.Println("x = ", x)    // 应该输出"x = 3"
}

func add2(a *int) int {
	*a = *a + 1 // 我们改变了a的值
	return *a   //返回一个新值
}

// defer
/*
Go语言中有种不错的设计，即延迟（defer）语句，你可以在函数中添加多个defer语句。
当函数执行到最后时，这些defer语句会按照逆序执行，最后该函数返回。
特别是当你在进行一些打开资源的操作时，遇到错误需要提前返回，在返回前你需要关闭相应的资源，不然很容易造成资源泄露等问题。
*/
func testDefer() {

	for i := 0; i < 5; i++ {
		defer fmt.Printf("%d ", i)
	}
}

// 函数作为值、类型
// 在Go中函数也是一种变量，我们可以通过type来定义它，它的类型就是所有拥有相同的参数，相同的返回值的一种类型
func testZhiType() {

	slice := []int{1, 2, 3, 4, 5, 7}
	fmt.Println("slice = ", slice)
	odd := filter(slice, isOdd) // 函数当做值来传递了
	fmt.Println("Odd elements of slice are: ", odd)
	even := filter(slice, isEven) // 函数当做值来传递了
	fmt.Println("Even elements of slice are: ", even)
}

type testInt func(int) bool // 声明了一个函数类型

func isOdd(integer int) bool {
	if integer%2 == 0 {
		return false
	}
	return true
}

func isEven(integer int) bool {
	if integer%2 == 0 {
		return true
	}
	return false
}

// 声明的函数类型在这个地方当做了一个参数
func filter(slice []int, f testInt) []int {
	var result []int
	for _, value := range slice {
		if f(value) {
			result = append(result, value)
		}
	}
	return result
}

// Panic和Recover
/*
Go没有像Java那样的异常机制，它不能抛出异常，而是使用了panic和recover机制。
一定要记住，你应当把它作为最后的手段来使用，也就是说，你的代码中应当没有，或者很少有panic的东西。

Panic
是一个内建函数，可以中断原有的控制流程，进入一个令人恐慌的流程中。
当函数F调用panic，函数F的执行被中断，但是F中的延迟函数会正常执行，然后F返回到调用它的地方。
在调用的地方，F的行为就像调用了panic。
这一过程继续向上，直到发生panic的goroutine中所有调用的函数返回，此时程序退出。
恐慌可以直接调用panic产生。也可以由运行时错误产生，例如访问越界的数组。

Recover
是一个内建的函数，可以让进入令人恐慌的流程中的goroutine恢复过来。recover仅在延迟函数中有效。
在正常的执行过程中，调用recover会返回nil，并且没有其它任何效果。
如果当前的goroutine陷入恐慌，调用recover可以捕获到panic的输入值，并且恢复正常的执行。
*/

var user = os.Getenv("USER")

func init() {
	if user == "" {
		panic("no value for $USER")
	}
}

func throwsPanic(f func()) (b bool) {
	defer func() {
		if x := recover(); x != nil {
			b = true
		}
	}()
	f() //执行函数f，如果f中出现了panic，那么就可以恢复回来
	return
}

// main函数和init函数
/*
Go里面有两个保留的函数：init函数（能够应用于所有的package）和main函数（只能应用于package main）。
这两个函数在定义时不能有任何的参数和返回值。
虽然一个package里面可以写任意多个init函数，但这无论是对于可读性还是以后的可维护性来说，我们都强烈建议用户在一个package中每个文件只写一个init函数。
Go程序会自动调用init()和main()，所以你不需要在任何地方调用这两个函数。每个package中的init函数都是可选的，但package main就必须包含一个main函数。
*/

// import
/*
1 点操作

我们有时候会看到如下的方式导入包

import(
    . "fmt"
)
这个点操作的含义就是这个包导入之后在你调用这个包的函数时，你可以省略前缀的包名，也就是前面你调用的fmt.Println("hello world")可以省略的写成Println("hello world")

2 别名操作

别名操作顾名思义我们可以把包命名成另一个我们用起来容易记忆的名字

import(
    f "fmt"
)
别名操作的话调用包函数时前缀变成了我们的前缀，即f.Println("hello world")

3 _操作

这个操作经常是让很多人费解的一个操作符，请看下面这个import

import (
    "database/sql"
    _ "github.com/ziutek/mymysql/godrv"
)
_操作其实是引入该包，而不直接使用包里面的函数，而是调用了该包里面的init函数。
*/
