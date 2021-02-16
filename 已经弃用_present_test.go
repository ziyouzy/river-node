package river_node

import (
	"fmt"
	"time"
	"testing"	
)


type struct testConfig{
	b [4096]byte
}

type struct test{
	timer *time.Timer
	name string
	conf *testConfig
}

func (p *test)Init(c *testConfig){
	p.name ="测试结构类"
	p.conf =c
}

func (p *test)Run(){
	p.timer =NewTimer(2 * time.Second)
	fmt.Println("p.conf.b:",p.conf.b)
}

func (p *test)Destory(){
	p.name =""
	p.timer = nil
	p.conf = nil
}

func TestNode(t *testing.T) {
	tsConf :=&testConfig{
			b:{0x01,0x02,0x03,}
		}

	ts :=&test{}

	go func(){
		/** ts的内部字段都不会被销毁，但是各个字段的值会被重复赋值
		 * 或者说，只有某个字段或变量，对象会被销毁，如 i int =3的i
		
		 * 然而 p := &testConf{} 的&testConf{}并不会存在销毁一说，
		   如果他在某一时刻不再指向p了

		 * 这就是下面for循环会讨论的情况
		 */
		for{
			ts.Init(tsConf)//每次Init都会创建一个新内存，并将内存指向就有的字段
			/** 得出结论，字段与内存间的指向关系

			 * 可以理解成两种不同形式的状态保持：
			   1.如下面的select保持住了ts(或者说ts =&test{xxx}这个“整体”)不被销毁
			   2.而ts =&test{xxx}的后者通过=对前者的绑定确保了&test{xxx}不被销毁
			
			 * 之前所说的
			   “当某个变量不会再有内存指向他了，他所在的作用域结束后他就会被一起销毁”
			   这句话只是在介绍上面1的情况，这也是c++，java语言对于变量整体的销毁机制
			   无论是int i =12 还是 int* p =&i;的i或p都是这一规则
			   这操作并不属于gc功能的一部分，而仅仅是属于函数功能的一部分

			 * 而第二种情况才会用到gc的回收机制

			 * gc的回收机制和函数的变量销毁机制是对立统一的

			 * 同样对立统一的还有java与golang
			   java的函数会同时负责变量（无论这个变量的值是值类型还是引用类型）的销毁
			   以及“值”类型值所占用内存的销毁，java的GC则只负责指针类型“值”所占用内存的销毁
			   golang函数会同时负责变量（无论这个变量的值是值类型还是引用类型）的销毁
			   golang的GC则同时负责值类型“值”所占用内存的销毁以及指针类型“值”所占用内存的销毁

			 * 可以更加深化的再次理解：
			   当一个代码块生命周期结束后，代码块内所有创建的“变量名”都会被立刻销毁
			   但是变量名基本上都是被某一块内存引用所指向的目标
			   假设某个引用所指向目标的总数为5，这个代码快结束后则5-1=4
			   当1-1=0最终发生时则就等同于某个变量被update了新数据了，如：
			   ts :=&test{name:"123"}；ts =&test{name: "wtf"}
			   第一个&test{name:"123"}会立刻进入gc机制流程，而不是等代码块生命周期结束才会进入

			 * 再回到这句“当某个变量不会再有内存指向他了，他所在的作用域结束后他就会被一起销毁”
			 * 这里的变量纯技术层上指的是“变量名+变量值”但是在语义上指的其实是“变量值”
			 * 只要“变量名”所在的作用域结束了变量名就会立刻销毁
			   这似乎并不属于gc的一部分，而是函数特性的一部分
			   而与此同时如果“变量值”没有指向其他存活状态的作用域，那么“变量值”就会进入gc流程
			 */

			ts.Run()
			ts.Destory()

			time.Sleep(3 * time.Second)
		}
	}()

	select{}//在这里阻塞会让tsConf和ts都不会被销毁
}