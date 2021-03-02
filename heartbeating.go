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

var hb_signal_rebuild,hb_signal_normal,hb_signal_panic Signal 

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
	timer 			*time.Timer
	config 			*HeartBeatingConfig
}

func (p *HeartBeating)Name()string{
	return HB_RIVERNODE_NAME
}

func (p *HeartBeating)Init(heartBeatingConfigAbs Config) error{
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

	if c.Signals == nil || c.Errors == nil{
		return errors.New("heartbeating river-node init error, "+
					 "Raws or Signals or Errors is nil")
	}

	//可以不传入外层Raws的指针，因为心跳包目前看来只用来检测“频率”
	if c.Raws == nil{
		c.Raws = make(chan struct{})
	}
	
	
	p.config = c

	hb_signal_rebuild = NewSignal(HEARTBREATING_REBUILD,c.UniqueId, "")
	hb_signal_normal  = NewSignal(HEARTBREATING_NORMAL,c.UniqueId, "")
	hb_signal_panic = NewSignal(HEARTBREATING_PANIC,c.UniqueId, "")
	
	return nil
}



/** Run()必须确保不返回error、不fmt错误

 * 这些问题都要在Init()函数内实现纠错与警告
 * 如发生了错误(非异常)则需要本包自行解决，而不能通过singal借助上层完成
 */

var (
	hb_count int
	hb_signal_run Signal
)

func (p *HeartBeating)Run(){
	hb_signal_run = NewSignal(HEARTBREATING_RUN,p.config.UniqueId,
				 fmt.Sprintf("heartbeating适配器开始运行，其UniqueId为%s, 最大超时秒数为%d, "+
				    "最大超时次数为%d, 该适配器无诸如“Mode”相关的配置参数。",
					p.config.UniqueId, p.config.TimeoutSec, p.config.TimeoutLimit))

	p.config.Signals <- hb_signal_run
	if p.timer ==nil{		  
		p.timer = time.NewTimer(p.config.TimeoutSec)
	}else{
		p.timer.Reset(p.config.TimeoutSec)
		p.config.Signals <- hb_signal_rebuild
	}


	go func(){
		for {
			select {
			//心跳包超时
			case <-p.timer.C:
				/*必须先检查一下Raws内部是否还存在数据*/				
				if len(p.config.Raws)>0{
					_ = <-p.config.Raws
					p.config.Errors <-NewError(HEARTBREATING_TIMERLIMITED,p.config.UniqueId,
						fmt.Sprintf("heartbeating适配器发生了“计时器超时下的数据临界事件“,Raws管道已"+
						   "正常排空，uid为%s",p.config.UniqueId)) 
				}

				if hb_count < p.config.TimeoutLimit{
					p.timer.Reset(p.config.TimeoutSec)
					hb_count++
					p.config.Errors <- NewError(HEARTBREATING_TIMEOUT,
												p.config.UniqueId,fmt.Sprintf("连续第%d次超时，"+
												"当前系统设定的最大超时次数为%d",
												hb_count,p.config.TimeoutLimit))
				}else{
					p.config.Errors <-NewError(HEARTBREATING_PANIC,
												p.config.UniqueId, fmt.Sprintf("连续第%d次超时"+
												"已超过系统设定的最大超时次数，系统设定的最大超时次数为%d",
												hb_count,p.config.TimeoutLimit))
					hb_count =0
					p.config.Signals <- hb_signal_panic
					//进行析构，但是暂时先不进行
					return
				}

			//心跳包未超时
			case <-p.config.Raws:
				/*Reset一个timer前必须先正确的关闭它*/
				if p.timer.Stop() == STOPAFTEREXPIRE{ 
					_ = <-p.timer.C 
					p.config.Errors <-NewError(HEARTBREATING_TIMERLIMITED,
												p.config.UniqueId,fmt.Sprintf("heartbeating适配器"+
												"发生了“计时器未超时下的数据临界事件”,计时器自身的管道"+
												"已正常排空，uid为%s",p.config.UniqueId)) 
				}
				
				if hb_count != 0{
					hb_count =0
					p.config.Signals <-NewSignal(HEARTBREATING_RECOVERED,
												 p.config.UniqueId, fmt.Sprintf("已从第%d次超时"+
												 "中恢复，当前系统设定的最大超时次数为%d",
												 hb_count,p.config.TimeoutLimit))
				}

				p.timer.Reset(p.config.TimeoutSec)
				p.config.Signals <- hb_signal_normal
			}
		}
	}()		
}



func NewHeartbBreating() NodeAbstract {
	return &HeartBeating{}
}


func init() {
	Register(HB_RIVERNODE_NAME, NewHeartbBreating)
	logger.Info("预加载完成，心跳包适配器已预加载至package river_node.Nodes结构内")
}
	









