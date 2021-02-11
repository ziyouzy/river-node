/** 此测试针对的是：
 * zadapter.go中的Adapters
 * 以及package zadapter的一些conf、log的相关组件
 */
package zadapter

import (
	"fmt"
	"time"
	"testing"	
)

func TestZadapter(t *testing.T) {
	testBytesSenderCH := make(chan []byte)
	NewLogger()
	Logger_Info("test Log Info")

	mainTestSignalCh :=make(chan int)


	mainTestHbRawCh :=make(chan struct{})
	//无NewCh需求

	/** 详细的演示会在zadapter/test包内进行
	 * mainTestStampsRawCh :=make(chan []byte)
	 * mainTestStampsNewChan :=make(chan []byte)

	 * mainTestCRCRawCh :=make(chan []byte)
	 * mainTestCRCNewChan :=make(chan []byte)
	 */
	

	/** 这一行的前提前提条件是Adapters这个map不为空
	 * 也就是说已通过如下方式实现了初始化：
	 * import _ "zadapter/heartbreating"或
	 * import _ "zadapter/crc"或
	 * import _ "zadapter/stamps"
	 */
	heartBeatingAbsf := Adapters[HEARTBEATING_ADAPTER_NAME]

	/** heartBeatingAbsf只是一个能返回接口实体的函数
	 * 在此适配器的逻辑，接口实体的形成必然滞后于一个实现了他的结构类
	 * 于是这里表面上实现的是个接口，其实实现的是个结构类
	 * 而这样的设计目的在于预加载，因为接口不会耗费资源
	 * 但是这个结构类很耗费
	 * 可以理解成执行heartBeatingAbsf这个函数才是创建了一个真正的对象
	 * 上一步只是从目录里挑出一个需要使用的对象的“标签”
	 * 逻辑上有一点像是个“路径”
	 */
	heartBeatingAbs :=heartBeatingAbsf()


	/** 拿到对象类后，内部字段很多都是空的，所以需要进行初始化
	 * 这里的实质也是对数据流动进行管道对接操作
	 * 归根揭底数据流动也只是一种设计模式，和这里所设计的适配器模式一样都是设计模式
	 * 只要是设计模式，那么首要任务都是为了方便日后的代码维护
	 * 同时，对于下面要进行的管道对接操作其实不该在这里实现的，在设计哲学上说不通
	 * 而应该是在“使用package zadapter”里进行相关的操作
	 * 下面只是简单的演示一下
	 */
	heartBeatingConfig := &HeartBeatingConfig{
		timeout : 8 * time.Second,
		uniqueId : "testCh",
		signalChan : mainTestSignalCh,
		rawChan : mainTestHbRawCh,
	}
	if err := heartBeatingAbs.Init(heartBeatingConfig); err == nil {
		Logger_Info("test adapter init success")
	}

	heartBeatingAbs.Run()




	//实现生成数据源的管道
	go func(){
		defer close(testBytesSenderCH)
		for {
			testBytesSenderCH <- []byte{0x01, 0x02, 0x03,}
			time.Sleep(time.Second)
		}
	}()



	/** 这里这个线程只实现针对心跳包适配器管道的数据注入操作操作
	 * 实际应用时这个线程内部会实现所有所需适配器管道的数据注入操作
	 *
	 * 注意，这里会进行的是“！所有！”数据源的接收操作
	 *
	 */
	go func(){
		defer close(mainTestHbRawCh)
		//defer close(mainTestStampsRawCh)
		//defer close(mainTestCRCRawCh)
		for bytes := range testBytesSenderCH{
			fmt.Println("bytes is", bytes)

			mainTestHbRawCh <- struct{}{}
			//mainTestStampsRawCh <- bytes	这里不过多进行演示，详细的演示会在zadapter/test包内进行	
			//mainTestCRCRawCh <- bytes		这里不过多进行演示，详细的演示会在zadapter/test包内进行
	 	}
	}()




	/**这里进行signal的统一回收工作
	 *
	 * 注意，这里会进行的是“！所有！”信号的回收
	 *
	 */

	go func(){
		defer close(mainTestSignalCh)
		for signal := range mainTestSignalCh{
			switch signal{
			case define.HEARTBREATING_NORMAL:
				fmt.Println("signal:", "HEARTBREATING_NORMAL")
			case define.HEARTBREATING_TIMEOUT:
				fmt.Println("signal:", "HEARTBREATING_TIMEOUT")
			case define.CRC_NORMAL:
				fmt.Println("这里不过多进行演示，详细的演示会在zadapter/test包内进行")
			case define.CRC_UPSIDEDOWN:
				fmt.Println("这里不过多进行演示，详细的演示会在zadapter/test包内进行")
			case define.CRC_NOTPASS:
				fmt.Println("这里不过多进行演示，详细的演示会在zadapter/test包内进行")
			case define.ANOTHEREXAMPLE_TEST1:
				fmt.Println("这里不过多进行演示，详细的演示会在zadapter/test包内进行")
			case define.ANOTHEREXAMPLE_TEST2:
				fmt.Println("这里不过多进行演示，详细的演示会在zadapter/test包内进行")
			case define.ANOTHEREXAMPLE_TEST3:
				fmt.Println("这里不过多进行演示，详细的演示会在zadapter/test包内进行")
			case define.ANOTHEREXAMPLE_ERR:
				fmt.Println("这里不过多进行演示，详细的演示会在zadapter/test包内进行")
			default:
				fmt.Println("未知的适配器返回了未知的信号类型这里不过多进行演示，"+
							"详细的演示会在zadapter/test包内进行")
			}			
		}
	}()

	/** 最后一个线程用来实现各个适配器所生成新数据的数据调度工作
	 * 全局只会有1个signal回收管道
	 * 每个适配器都会有独立的数据注入管道
	 * 之后，根据实际需求，有的适配器会拥有新数据管道，有的则没有
	 * 而所有“需要有新数据”的适配器管道都会在此线程实现汇总
	 */
	 //go func(){
		//defer close(mainTestStampsNewChan)
		//defer close(mainTestCRCNewChan)

		/** 如何才算是一次完整的数据流“分流”事件呢？
		 * 首先数据分流是数据处理的一种形式，或者说是分型
		 * 数据处理只有两种形式，分流与截流
		 * 从某个数据管道拿到事件后（for range拿到某一个切片）
		 * 可能会做的一些事，包括但不限于：
		 * 1.原始数据流截断
		    （直接让数据消失，等同于 _ =xxx的效果
			 他不属于数据分流，而是节流的一种
			 “数据分流”是“数据处理”的一种分型）
			 
		 * 2.原始数据流转化为一个新数据流(新流只有一个也算分流)
		 * 3.原始数据流转化为多个新数据流
			（转化的方式可以基于元数据流数据
			 也可以是直接引入新的流）

		 * 4.其他的数据截流处理方式：
		   如fmt.Println()输出
		   或最后一次不需创建新管道的数据转换+录入数据库
		   等

		 * 5.数据合流(聚合，无论是借助缓存实现的还是不借助缓存实现的)
		 
		 * 而在这里进行根本就不属于上面任何一种，因为上一次数据处理（分流）已经完成了
		 * 这里的for range其实是已经开始了新一轮的数据处理（分流）
		 
		 * 其实可以概括成，每个for range{}都是一个独立的“数据处理节点”
		 */

		//for newrcrcbytes := range mainTestCRCNewChan{

		//}

		//for newStampsbytes := range mainTestStampsNewChan{

		//}
	//}()

	//这里的测试并没有实现CRC适配器与Stamps包的任何对象，所以测试到此未知

	//复杂些的，会同时包含数据处理与传递的测试实例请看package zadapter/test所包含的内容
	
	select{}
}