/** 此测试针对的是：
 * river_node.go中的Nodes
 * 以及package river-node的一些conf、log的相关组件
 */
package river_node

import (
	"river_node/heartbeating"
	"river_node/define"
	"river_node/logger"

	"fmt"
	"time"
	"testing"	
)

func TestNode(t *testing.T) {
	logger.NewLogger()
	defer logger.Flush()

	logger.Info("test Log Info")


	/*此管道的作用是测试信号的生成*/
	testBytesSenderCH := make(chan []byte)
	/*主线程中signal的统一管理管道*/
	mainTestSignalCh := make(chan int)
	/*心跳包的事件注入管道*/
	mainTestHbRawCh :=make(chan struct{})





	/** 暂不设计NewChl类管道的测试需求需求
	 *  详细的演示会在river-node/test包内进行
	    mainTestStampsRawCh :=make(chan []byte)
	    mainTestStampsNewChan :=make(chan []byte)

	    mainTestCRCRawCh :=make(chan []byte)
	    mainTestCRCNewChan :=make(chan []byte)
	 */
	

	/** 这一行的前提前提条件是Nodes这个map不为空

	 *  也就是说已通过如下方式实现了初始化：
	    import "river-node/heartbreating"或
	    import "river-node/crc"或
	    import "river-node/stamps"

	 *  golang的机制决定了即使在遵循单向调用链模式的前提现某个包被多个包同时引用
	    这个被引用包无论是整体还是内部字段都只会存在唯一的一份副本，即使是package fmt也是如此

	 *  于是引入顺序偶尔也变得比较重要，比如当前的情况：
	    "river-node/heartbeating"
	    "river-node/define"
	    "river-node/logger"
	    由于logger是被所有功能包都会调用的工具包，所以必须确保其先完成初始化
	    不会这里也不是必须在最下方，因为最下方的只是最先被塞进内存，或者说完成预加载
	    真正的初始化则是当前测试函数的第一句
	    只要确保logger的“初始化”在所有的包真正去使用他之前完成即可
	 */
	heartBeatingAbsf := Nodes[heartbeating.NODE_NAME]


	/** heartBeatingAbsf只是一个能返回接口实体的函数

	 *  在此适配器的逻辑，接口实体的形成必然滞后于一个实现了他的结构类
	    于是这里表面上实现的是个接口，其实实现的是个结构类
	    而这样的设计目的在于预加载，因为接口不会耗费资源
	    但是这个结构类很耗费

	 *  可以理解成执行heartBeatingAbsf这个函数才是创建了一个真正的对象
	    上一步只是从目录里挑出一个需要使用的对象的“标签”
	    逻辑上有一点像是个“路径”
	 */
	fmt.Println("heartBeatingAbsf：",heartBeatingAbsf)
	heartBeatingAbs :=heartBeatingAbsf()


	/** 拿到对象类后，内部字段很多都是空的，所以需要进行初始化

	 *  这里的实质也是对数据流动进行管道对接操作
	    归根揭底数据流动也只是一种设计模式，和这里所设计的适配器模式一样都是设计模式
	    只要是设计模式，那么首要任务都是为了方便日后的代码维护
	    同时，对于下面要进行的管道对接操作其实不该在这里实现的，在设计哲学上说不通
		而应该是在“使用package river-node”里进行相关的操作
		
	 *  下面只是简单的演示一下

	 */

	heartBeatingConfig := &heartbeating.HeartBeatingConfig{
		Timeout : 8 * time.Second,
		UniqueId : "testCh",
		SignalChan : mainTestSignalCh,
		RawChan : mainTestHbRawCh,
	}

	if err := heartBeatingAbs.Init(heartBeatingConfig); err == nil {
		Logger_Info("test river-node init success")
	}

	heartBeatingAbs.Run()


	/** 业务流程的第一步，先进行signal的统一回收工作 
	    注意，这里会进行的是“！所有！”信号的回收
		此线程的作用是信号的处理
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
				fmt.Println("这里不过多进行演示，详细的演示会在river-node/test包内进行")
			case define.CRC_UPSIDEDOWN:
				fmt.Println("这里不过多进行演示，详细的演示会在river-node/test包内进行")
			case define.CRC_NOTPASS:
				fmt.Println("这里不过多进行演示，详细的演示会在river-node/test包内进行")
			case define.ANOTHEREXAMPLE_TEST1:
				fmt.Println("这里不过多进行演示，详细的演示会在river-node/test包内进行")
			case define.ANOTHEREXAMPLE_TEST2:
				fmt.Println("这里不过多进行演示，详细的演示会在river-node/test包内进行")
			case define.ANOTHEREXAMPLE_TEST3:
				fmt.Println("这里不过多进行演示，详细的演示会在river-node/test包内进行")
			case define.ANOTHEREXAMPLE_ERR:
				fmt.Println("这里不过多进行演示，详细的演示会在river-node/test包内进行")
			default:
				fmt.Println("未知的适配器返回了未知的信号类型这里不过多进行演示，"+
							"详细的演示会在river-node/test包内进行")
			}			
		}
	}()


	/*为完成测试，如下携程实现的是向最初管道注入数据源*/

	/** 这个管道不属于整体系统的一部分，只是为了产生数据
	 * 真正系统的规则如下：
	 * 所有属于系统的管道会在主函数的开端统一进行make操作
	   在任何一个数据处理节点代码块中(go func(){for range{}}结构)
	   defer后跟随的是被处理的数据管道
	   range后跟随的是处理后诞生的新管道
	   所联合使用的适配器包，如heartbeating内是不会存在数据处理节点的
	   即使存在go func(){for range{}}结构，他也不属于整体数据流动设计模式的一部分
	   而仅仅是为适配器包自身实现某种功能所写出的代码逻辑
	 */
	go func(){
		defer close(testBytesSenderCH)
		for {
			testBytesSenderCH <- []byte{0x01, 0x02, 0x03,}
			time.Sleep(time.Second)
		}
	}()


	/** 之后，如下携程在当前测试中只会实现针对心跳包适配器管道的数据注入操作操作
	  
	 * 注意，实际应用时这个线程内部不会实现整个程序所有所需适配器管道的数据注入操作
	   而是只实现for range所指向目标管道将会分流出管道的数据注入操作
	 
	 * 注意，这里是一个“数据处理节点”，后面会有更详细的说明
	 
	 * 此线程的作用是信号的处理
	 */
	go func(){
		defer close(mainTestHbRawCh)
		//defer close(mainTestStampsRawCh)
		//defer close(mainTestCRCRawCh)
		for bytes := range testBytesSenderCH{
			fmt.Println("bytes is", bytes)

			mainTestHbRawCh <- struct{}{}
			//mainTestStampsRawCh <- bytes	这里不过多进行演示，详细的演示会在river-node/test包内进行	
			//mainTestCRCRawCh <- bytes		这里不过多进行演示，详细的演示会在river-node/test包内进行
	 	}
	}()


	/** 如下线程用来实现各个适配器所生成新数据的数据调度工作
	 * 要明确全局只会有1个signal回收管道，这个回收管道所在携程就是主携程，设计时要时刻考虑、遵循单向调用链模式
	 
	 * 之后的思路，每个适配器都会有独立的数据注入管道
	 * 都是通过各自的config对象类实现“管道的对接”与“新事件的生成与注入”

	 * 不过根据实际需求，有的适配器会拥有新数据管道，有的则没有
	  
	 * 而所有“新的数据管道”，根据业务需求，都会通过和上方携程同样的for range{}
	 * 实现新的“数据处理节点”

	 * 下方携程就是个大致的例子
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

	//复杂些的，会同时包含数据处理与传递的测试实例请看package river-node/test所包含的内容
	
	select{}
}