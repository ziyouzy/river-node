/*p.config.eventChan与p.config.rawinChan均在上层make与close*/

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

const HB_NODE_NAME = "heartbeating心跳包适配器"

type HeartBeatingConfig struct{
	UniqueId 		string	/*其所属上层Conn的唯一识别标识*/
	Events 			chan EventAbs /*发送给主进程的信号队列，就像Qt的信号与槽*/
	Errors 			chan error

	Raws			chan []byte
	News			chan []byte

	TimeoutSec 		time.Duration
	Limit  			int
		 
	//News 此包虽然不会生成新的管道数据，但是必须作为river中最后一个管道
}

func (p *HeartBeatingConfig)Name()string{
	return HB_NODE_NAME
}


/** node节点自身所包含的字段要么是config，
 * 要么是一些golang如time.Timer的内置字段
 * 后期也可能会遇到装配一些第三方包内工具对象的时候
 */
type HeartBeating struct{
	timer 				*time.Timer
	config 				*HeartBeatingConfig

	countor 			int

	warpError_Panich 	error
}

func (p *HeartBeating)Name()string{
	return HB_NODE_NAME
}

func (p *HeartBeating)Construct(heartBeatingConfigAbs Config) error{
	if heartBeatingConfigAbs.Name() != HB_NODE_NAME {
		return errors.New(
			fmt.Sprintf("[river-node type:%s] init error, config must HeartBreatingConfig", 
			p.Name()))
	}


	v := reflect.ValueOf(heartBeatingConfigAbs)
	c := v.Interface().(*HeartBeatingConfig)

	if c.UniqueId == ""{
		return errors.New(
			fmt.Sprintf("[river-node type:%s] init error, uniqueId is nil", 
			p.Name()))
	}

	if c.TimeoutSec == (0 * time.Second) || c.Limit ==0{
		return errors.New(
			fmt.Sprintf("[uid:%s] init error, timeout or timeoutlimit is nil",
			c.UniqueId))
	}

	if c.Events == nil || c.Errors == nil || c.Raws == nil{
		return errors.New(
			fmt.Sprintf("[uid:%s] init error, Raws or Events or Errors Raws is nil",
			c.UniqueId))
	}
	
	p.config = c

	p.warpError_Panich = fmt.Errorf(
		"%w",NewEvent(HEARTBREATING_PANICH, c.UniqueId, "", nil, ""))

	p.config.News = make(chan []byte)
	
	return nil
}



func (p *HeartBeating)Run(){
	p.config.Events <-NewEvent(
		HEARTBREATING_RUN, p.config.UniqueId,"",nil,
		
		fmt.Sprintf("节点开始运行，最大超时秒数为%v, 最大超时次数为%d,"+
			"该适配器无诸如“Mode”相关的配置参数。",
			p.config.TimeoutSec, p.config.Limit))

	p.timer = time.NewTimer(p.config.TimeoutSec)

	go func(){
		defer p.reactiveDestruct()
		for {
			select {
			//心跳包超时
			case <-p.timer.C:
				/*必须先检查一下Raws内部是否还存在数据*/				
				if len(p.config.Raws)>0{
					_ = <-p.config.Raws
					p.config.Errors <-fmt.Errorf(
						"%v", NewEvent(HEARTBREATING_TIMERLIMITED, p.config.UniqueId, "", nil,
						"发生了“计时器超时下的数据临界事件“,Raws管道已正常排空")) 
				}

				if p.countor < p.config.Limit{
					p.timer.Reset(p.config.TimeoutSec)

					p.config.Errors <-fmt.Errorf(
						"%v", NewEvent(HEARTBREATING_TIMEOUT, p.config.UniqueId, "", nil,
						fmt.Sprintf("连续第%d次超时，当前系统设定的最大超时次数为%d",
						p.countor, p.config.Limit)))
					
					p.countor++
				}else{
					p.config.Errors <-fmt.Errorf(
						"%v", NewEvent(HEARTBREATING_TIMEOUT, p.config.UniqueId, "", nil,
						fmt.Sprintf("连续第%d次超时已超过系统设定的最大超时次数，"+
						"系统设定的最大超时次数为%d", p.countor, p.config.Limit)))
				
					p.config.Errors <-p.warpError_Panich
					p.countor =-1

				}

			//心跳包未超时
			case baits, ok :=<-p.config.Raws:
				if !ok { 
					return 
				}

				//被析构前夕的等待状态
				if p.countor <0{
					_ =baits
					continue
				}

				p.config.News<-baits

				/*Reset一个timer前必须先正确的关闭它*/
				if p.timer.Stop() == TIMER_STOPAFTEREXPIRE{ 
					_ = <-p.timer.C 
					p.config.Errors <-fmt.Errorf(
						"%v",NewEvent(HEARTBREATING_TIMERLIMITED, p.config.UniqueId, "", nil,
						"发生了“计时器未超时下的数据临界事件”,计时器自身的管道已正常排空")) 
				}
				
				if p.countor != 0{
					p.config.Events <-NewEvent(
						HEARTBREATING_RECOVERED,p.config.UniqueId, "", nil,
						fmt.Sprintf("已从第%d次超时中恢复，当前系统设定的最大超时次数为%d",
						p.countor, p.config.Limit))
						
					p.countor =0
				}

				p.timer.Reset(p.config.TimeoutSec)
			}
		}
	}()		
}


//被动 - reactive
//被动析构是检测到Raws被上层关闭后的响应式析构操作
func (p *HeartBeating)reactiveDestruct(){
	//析构操作在前，管道内就算有新事件也不需要了
	_ = p.timer.Stop()
	
	close(p.config.News)

	p.config.Events <-NewEvent(
		HEARTBREATING_REACTIVE_DESTRUCT,p.config.UniqueId,"",nil,
		fmt.Sprintf("触发了隐式析构方法"))
}

func NewHeartbBreating() NodeAbstract {
	return &HeartBeating{}
}


func init() {
	Register(HB_NODE_NAME, NewHeartbBreating)
	logger.Info(
		fmt.Sprintf("预加载完成，[river-node type:%s]已预加载至package river_node.Nodes结构内",
		HB_NODE_NAME))
}
	









