/*p.config.signalChan与p.config.slotChan均在上层make与close*/


/** HeartBeating不仅仅可以用作tcp的心跳包，其他的链接类型，如果有长连接需求也适用
 * 具体的使用方式是，当外层完成通过net.Conn封装自定义Conn时后，将自定义Conn作为参数传入Handler方法
 * 自定义Conn（如下面的ZYUnifiedConn）是个接口，实现了Attach方法
 * (如：
	    ZYHB :=(HeartBeatHandler)HeartBeating(zqy_go_logger)
		ZYUnifiedConn ：=NewZYUnifiedConn(xxx,xxx,xxx)
		ZYUnifiedConn.Attach("heartbeating", ZYHB)
 * )
 * 函数数据类型是引用类型,最后的函数类型参数是比较重要的核心
 * 他被设计成自定义Conn的一个内部字段
 * 这样就可以把心跳包逻辑从整体套接字通信逻辑中抽离出来
 * 同时，ZYHB是个单例，他的Handler可以分别应用于多个自定义conn，自定义conn的内部是tcp，udp，snmp也都是可以的
 */


package heartbeating

import (

	//logger "github.com/phachon/go-logger"
	"zadapter"
	"zadapter/define"

	"fmt"
	"time"
	"reflect"
	"errors"
)

const ADAPTER_NAME = "heartbeating"



type HeartBeatingConfig struct{

	timeout time.Duration

	uniqueId string	/*其所属上层Conn的唯一识别标识*/
	
	signalChan chan int /*发送给主进程的信号队列，就像Qt的信号与槽*/


	/** 虽然是面向[]byte的适配器，但是并不需要[]byte做任何操作
	 * 所以在这里遵循golang的设计哲学
	 * 使用struct{}作为事件的传递介质
	 */

	rawChan chan struct{} /*从主线程发来的信号队列，就像Qt的信号与槽*/

	//l *logger.Logger
}

func (p *HeartBeatingConfig)Name()string{
	return ADAPTER_NAME
}



type HeartBeating struct{
	//timeout time.Second
	timer *time.Timer

	//sl []byte

	config *HeartBeatingConfig
}

func (p *HeartBeating)Name()string{
	return ADAPTER_NAME
}

func (p *HeartBeating)Init(heartBeatingConfigAbs zadapter.Config) error{
	if heartBeatingConfigAbs.Name() != ADAPTER_NAME {
		return errors.New("heartbeating adapter init error, config must HeartBreatingConfig")
	}


	vhb := reflect.ValueOf(heartBeatingConfigAbs)
	chb := vhb.Interface().(*HeartBeatingConfig)


	if chb.timeout == (0 * time.Second) || chb.uniqueId == "" {
		return errors.New("heartbeating adapter init error, timeout or uniqueId is nil")
	}

	if chb.signalChan == nil || chb.slotChan == nil{
		return errors.New("heartbeating adapter init error, slotChan or signalChan is nil")
	}
	
	
	p.config = chb


	return nil
}


/** Run()必须确保不返回error、不fmt错误
 * 这些问题都要在Init()函数内实现纠错与警告
 * Run()的一切问题都通过signal的方式传递至管道
 */
func (p *HeartBeating)Run(){

	if p.timer ==nil{

		/** 这里针对的是第一个从slot传来的事件
		 * 只有第一此会进行NewTimer() 
		 * timer被下方携程引用，生命周期100%由下方携程匿名函数的生命周期决定 
		 */
		  
		p.timer = time.NewTimer(p.config.timeout)
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
			case <-p.timer.C:
				
				p.config.signalChan<-define.HEARTBREATING_TIMEOUT

				// p.config.logger.Warning(fmt.Sprintln("UniqueId为 %s 的链接超时，"+
				// 									 "心跳包模块会自动做好析构工作"+
				// 									 "但不会干预上层套接字Conn的业务逻辑与销毁", 
				// 									 p.config.UniqueId))
				return

			case <-p.config.slotChan:

				/*Reset一个timer前必须先正确的关闭它*/
				if p.timer.Stop() == STOPAFTEREXPIRE{ 
				// 	p.config.logger.Warning("当心跳包进行timer的Reset操作时与timer自身的到期事件"+
				// 							"发生了race condition（竞争条件之下心跳事件发生在前")
			    	_ = <-p.timer.C 
				}
				
				/*正式Reset*/
				p.timer.Reset(p.config.timeout)

				p.config.signalChan<-define.HEARTBREATING_NORMAL
			}
		}
	}()		
}



/** 由于通信管道的存在似乎就无法为其设计初始化默认值的操作了
 * 但这个函数也是必须的，因为上层一定会遍历各个map
 * 从而识别并确认都有哪些已经注册并在册的预编译适配器
 */

func NewHeartbBreating() zadapter.AdapterAbstract {
	return &HeartBeating{}
}


func init() {
	zadapter.Register(ADAPTER_NAME, NewHeartbBreating)
}
	









