/*p.config.signalChan与p.config.rawinChan均在上层make与close*/

/** HeartBeating不仅仅可以用作tcp的心跳包

 * 其他的链接类型，或者是某个管道，也无论是长或短连接需求均适用
 */
package river_node

import (
	"github.com/ziyouzy/logger"

	"fmt"
	"time"
	"reflect"
	"errors"
)

const HB_RIVERNODE_NAME = "heartbeating"

type HeartBeatingConfig struct{
	UniqueId 		string	/*其所属上层Conn的唯一识别标识*/
	Signals 		chan Signal /*发送给主进程的信号队列，就像Qt的信号与槽*/
	Errors 			chan error

	/** 虽然是面向[]byte的适配器，但是并不需要[]byte做任何操作
	 * 所以在这里遵循golang的设计哲学
	 * 使用struct{}作为事件的传递介质
	 */

	TimeoutSec 		time.Duration
	TimeoutLimit    int
	Raws 			chan struct{} /*从主线程发来的信号队列，就像Qt的信号与槽*/
		 
	//News 此包不会生成新的管道数据
}

func (p *HeartBeatingConfig)Name()string{
	return HB_RIVERNODE_NAME
}


/** node节点自身所包含的字段要么是config，
 * 要么是一些golang如time.Timer的内置字段
 * 后期也可能会遇到装配一些第三方包内工具对象的时候
 */
type HeartBeating struct{
	timer 				*time.Timer
	config 				*HeartBeatingConfig

	timeout_countor 	int

	countor 			int
	signal_run 			Signal
	signal_fused 		Signal

	stop 				chan struct{}
}

func (p *HeartBeating)Name()string{
	return HB_RIVERNODE_NAME
}

func (p *HeartBeating)Construct(heartBeatingConfigAbs Config) error{
	if heartBeatingConfigAbs.Name() != HB_RIVERNODE_NAME {
		return errors.New("heartbeating river-node init error, "+ 
		             "config must HeartBreatingConfig")
	}


	v := reflect.ValueOf(heartBeatingConfigAbs)
	c := v.Interface().(*HeartBeatingConfig)


	if c.TimeoutSec == (0 * time.Second) || c.UniqueId == "" || c.TimeoutLimit ==0{
		return errors.New("heartbeating river-node init error, "+
		             "timeout or uniqueId or timeoutlimit is nil")
	}

	if c.Signals == nil || c.Errors == nil || c.Raws == nil{
		return errors.New("heartbeating river-node init error, "+
					 "Raws or Signals or Errors Raws is nil")
	}

	
	
	p.config = c

	p.signal_run = NewSignal(HEARTBREATING_RUN,p.config.UniqueId,
	 fmt.Sprintf("heartbeating适配器开始运行，其UniqueId为%s, 最大超时秒数为%d, "+
		"最大超时次数为%v, 该适配器无诸如“Mode”相关的配置参数。",
		p.config.UniqueId, p.config.TimeoutSec, p.config.TimeoutLimit))
	p.signal_fused = NewSignal(HEARTBREATING_FUSED,c.UniqueId, "")

	p.stop =make(chan struct{})
	
	return nil
}



/** Run()必须确保不返回error、不fmt错误

 * 这些问题都要在Init()函数内实现纠错与警告
 * 如发生了错误(非异常)则需要本包自行解决，而不能通过singal借助上层完成
 */


func (p *HeartBeating)Run(){
	p.config.Signals <- p.signal_run

	//并没有这个必要，心跳包只会发生销毁事件和再创建，不会发生重置事件
	// if p.timer ==nil{		  
	// 	p.timer = time.NewTimer(p.config.TimeoutSec)
	// }else{
	// 	p.timer.Reset(p.config.TimeoutSec)
	// 	p.config.Signals <- hb_signal_rebuild
	// }
	p.timer = time.NewTimer(p.config.TimeoutSec)

	go func(){
		defer p.reactiveDestruct()
		/** 
		 * 析构的时候此循环不会因p.Raws的close而退出(注意close并不代表p.Raws==nil)
		 * 因此需要设计独立的stop管道分支，分支的唯一作用是执行reactiveDestruct()跳出循环
		 */
		for {
			select {
			//心跳包超时
			case <-p.timer.C:
				/*必须先检查一下Raws内部是否还存在数据*/				
				if len(p.config.Raws)>0{
					_ = <-p.config.Raws
					p.config.Errors <-NewError(HEARTBREATING_TIMEOUT,p.config.UniqueId,
							fmt.Sprintf("heartbeating适配器发生了“计时器超时下的数据临界事件“,"+
							   "Raws管道已正常排空，uid为%s",p.config.UniqueId)) 
				}

				if p.countor < p.config.TimeoutLimit{
					p.timer.Reset(p.config.TimeoutSec)
					p.countor++
					p.config.Errors <- NewError(HEARTBREATING_TIMEOUT,p.config.UniqueId,
						    fmt.Sprintf("连续第%d次超时，当前系统设定的最大超时次数为%d",
						       p.countor,p.config.TimeoutLimit))
				}else{
					p.config.Errors <-NewError(HEARTBREATING_FUSED,p.config.UniqueId,
							fmt.Sprintf("连续第%d次超时已超过系统设定的最大超时次数，系统设定的最大超时"+
							   "次数为%d",p.countor,p.config.TimeoutLimit))
					p.countor =0
					p.config.Signals <- p.signal_fused
					return
				}

			//心跳包未超时
			case _, ok :=<-p.config.Raws:
				if !ok { fmt.Println("hb close"); return }
				/*Reset一个timer前必须先正确的关闭它*/
				if p.timer.Stop() == STOPAFTEREXPIRE{ 
					_ = <-p.timer.C 
					p.config.Errors <-NewError(HEARTBREATING_TIMERLIMITED,p.config.UniqueId,
							fmt.Sprintf("heartbeating适配器发生了“计时器未超时下的数据临界事件”,"+
							   "计时器自身的管道已正常排空，uid为%s",p.config.UniqueId)) 
				}
				
				if p.countor != 0{
					p.countor =0
					p.config.Signals <-NewSignal(HEARTBREATING_RECOVERED,p.config.UniqueId,
						    fmt.Sprintf("已从第%d次超时中恢复，当前系统设定的最大超时次数为%d",
							   p.countor,p.config.TimeoutLimit))
				}

				p.timer.Reset(p.config.TimeoutSec)

			case <-p.stop:
				return
			}
		}
	}()		
}


/** 
 * 对于上层逻辑来说，进行p.Raws的关闭操作后此包就会检测到关闭的操作
 * 从而隐式进行reactiveDestruct()

 * 真的主动执行他的话，p.Raws就会没法被消费了
 * 上层逻辑是会被卡主的,一定要慎用
 * 而且似乎现在这个项目也根本就用不到他，不过还是先设计出来吧
 */
func (p *HeartBeating)ProactiveDestruct(){
	/**
	 * 析构数据源,析构后Run()内部会自动触发下面的reactiveDestruct()方法
	 */
	//不应为其设计主动析构逻辑，因为Raws的关闭一定需要在外层进行
	//close(p.config.Raws)
	p.config.Signals <-NewSignal(HEARTBREATING_PROACTIVEDESTRUCT,p.config.UniqueId,
		    "注意，由于某些原因心跳包主动调用了显式析构方法")

	p.stop<-struct{}{}	
}

//被动 - reactive
//被动析构是检测到Raws被上层关闭后的响应式析构操作
func (p *HeartBeating)reactiveDestruct(){
	//必须执行Stop(),他和关闭一个io一样是必须的，否则timer内可能会存在携程泄露

	// if p.timer.Stop() == STOPAFTEREXPIRE{ 
	// 	_ = <-p.timer.C 
	// 	p.config.Errors <-NewError(HEARTBREATING_TIMERLIMITED,p.config.UniqueId,
	// 			fmt.Sprintf("heartbeating适配器发生了“计时器未超时下的数据临界事件”,"+
	// 			   "计时器自身的管道已正常排空，uid为%s",p.config.UniqueId)) 
	// }

	//析构操作在前，管道内就算有新事件也不需要了
	p.config.Signals <-NewSignal(HEARTBREATING_REACTIVEDESTRUCT,p.config.UniqueId,
		  	"心跳包触发了隐式析构方法")
	_ = p.timer.Stop()
	close(p.stop)
	
	//close(p.News)此适配器业务逻辑无需News管道

}

func NewHeartbBreating() NodeAbstract {
	return &HeartBeating{}
}


func init() {
	Register(HB_RIVERNODE_NAME, NewHeartbBreating)
	logger.Info("预加载完成，心跳包适配器已预加载至package river_node.Nodes结构内")
}
	









