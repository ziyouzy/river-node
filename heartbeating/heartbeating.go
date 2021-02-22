/*p.config.signalChan与p.config.rawinChan均在上层make与close*/

/** HeartBeating不仅仅可以用作tcp的心跳包

 * 其他的链接类型，或者是某个管道，也无论是长或短连接需求均适用
 */
package heartbeating

import (
	"river-node"
	"river-node/define"
	"river-node/logger"

	"time"
	"reflect"
	"errors"
)

const RIVER_NODE_NAME = "heartbeating"

var signal_run,signal_rebuild,signal_normal,error_timeout river_node.Signal 

type HeartBeatingConfig struct{
	UniqueId 		string	/*其所属上层Conn的唯一识别标识*/
	 Signals 		chan river_node.Signal /*发送给主进程的信号队列，就像Qt的信号与槽*/
	  Errors 		chan error

	/** 虽然是面向[]byte的适配器，但是并不需要[]byte做任何操作
	 * 所以在这里遵循golang的设计哲学
	 * 使用struct{}作为事件的传递介质
	 */

	 Timeout 		time.Duration
		Raws 		chan struct{} /*从主线程发来的信号队列，就像Qt的信号与槽*/
		 


		//News 此包不会生成新的管道数据

}

func (p *HeartBeatingConfig)Name()string{
	return RIVER_NODE_NAME
}


/** rnode节点自身所包含的字段要么是config，
 * 要么是一些golang如time.Timer的内置字段
 * 后期也可能会遇到装配一些第三方包内工具对象的时候
 */
type HeartBeating struct{
	timer 			*time.Timer
	config 			*HeartBeatingConfig
}

func (p *HeartBeating)Name()string{
	return RIVER_NODE_NAME
}

func (p *HeartBeating)Init(heartBeatingConfigAbs river_node.Config) error{
	if heartBeatingConfigAbs.Name() != RIVER_NODE_NAME {
		return errors.New("heartbeating river-node init error, config must HeartBreatingConfig")
	}


	v := reflect.ValueOf(heartBeatingConfigAbs)
	c := v.Interface().(*HeartBeatingConfig)


	if c.Timeout == (0 * time.Second) || c.UniqueId == "" {
		return errors.New("heartbeating river-node init error, timeout or uniqueId is nil")
	}

	if c.Signals == nil || c.Raws == nil || c.Errors == nil{
		return errors.New("heartbeating river-node init error, Raws or Signals "+ 
		                  "or Errors is nil")
	}
	
	
	p.config = c

	signal_run 		= river_node.NewSignal(river_node.HEARTBREATING_RUN,p.UniqueId)
	signal_rebuild 	= river_node.NewSignal(river_node.HEARTBREATING_REBUILD,p.UniqueId)
	signal_normal 	= river_node.NewSignal(river_node.HEARTBREATING_NORMAL,p.UniqueId)
	error_timeout 	= river_node.NewError(river_node.HEARTBREATING_TIMEOUT,p.UniqueId)

	return nil
}



/** Run()必须确保不返回error、不fmt错误

 * 这些问题都要在Init()函数内实现纠错与警告
 * Run()的一切问题都通过signal的方式传递至管道
 */
func (p *HeartBeating)Run(){
	p.config.Signals <- signal_run
	if p.timer ==nil{

		/** 这里针对的是第一个从slot传来的事件
		 * 只有第一此会进行NewTimer() 
		 * timer被下方携程引用，生命周期100%由下方携程匿名函数的生命周期决定 
		 */
		  
		p.timer = time.NewTimer(p.config.Timeout)
	}else{

		p.timer.Reset(p.config.Timeout)
		p.config.Signals <- signal_rebuild
		
	}


	go func(){
			
		/** 一般情况下“for循环退出”这一事件会直接析构掉p.timer
		 *(将实现p.timer==nil，但是心跳包自身是否销毁由上层逻辑决定)

		 * 如果上层决定心跳包超时触发析构整个心跳包所在Conn连接
		 * 那么各个适配器数组以及数组内所有适配器也会被一并清除
		 * 具体要看上层逻辑

		 * 因为p.timer的生命周期是此包生命周期的核心
		 * 当for循环退出后，此匿名函数也就不会在继续维系p.timer的引用
		 * 匿名函数后方再没有其他的逻辑会设计、实现“作用域状态保持”的逻辑代码 

		 * p.config.signalChan可能会出现拥堵造成进程泄露
		 * 但是该管道其实并不属于此包，而是程序枢纽层的消息机制管道
		 * 此位置是否会阻塞直接取决于上层的设计逻辑
		 * 上层一定会确保不会出现逻辑疏漏所造成的管道阻塞
		 *
		 * p.config.slotChan也可能会出现拥堵造成进程泄露
		 * 此管道内事件的消费属于此心跳包的职责，需要认真设计确保不出错
		 * 
		 */

		for {
			select {
			case <-p.timer.C://心跳包超时				
				p.config.Errors <- error_timeout
				if len(p.config.Raws)>0{ _ = <-p.config.Raws }
				return

			case <-p.config.Raws:

				/*Reset一个timer前必须先正确的关闭它*/
				if p.timer.Stop() == STOPAFTEREXPIRE{ _ = <-p.timer.C }
				/*正式Reset*/
				p.timer.Reset(p.config.Timeout)
				p.config.Signals <- signal_normal

			}
		}
	}()		
}



/** 下面是对package river-node中的map进行初始化

 * 真正使用他的上层一定会遍历各个map
 * 从而识别并确认都有哪些已经注册并在册的预编译适配器
 */

func NewHeartbBreating() river_node.NodeAbstract {
	return &HeartBeating{}
}


func init() {
	river_node.Register(RIVER_NODE_NAME, NewHeartbBreating)
	logger.Info("预加载完成，心跳包适配器已预加载至package river_node.Nodes结构内")
}
	









