package river_node

import (
	//"river-node/define"	/*暂时没有需要返回给上层的event内容*/
	"github.com/ziyouzy/logger"
	
	"fmt"
	"bytes"
	"reflect"
	"errors"
)

const STAMPS_RIVERNODE_NAME = "stamps适配器"

type StampsConfig struct{
	UniqueId 	    			string	/*其所属上层Conn的唯一识别标识*/
	Events 	    				chan Event /*发送给主进程的信号队列，就像Qt的信号与槽*/
	Errors 		    			chan error
	/** 分为三种，HEADS、TAILS、HEADSANDTAILS
	 * 当是HEADANDTAILS模式，切len(stamp)>1时
	 * 按照如下顺序拼接：
	 * stamp1+raw(首部);raw+stamp2(尾部);stamp3+raw(首部);raw+stamp4(尾部);stamp5+raw(首部);....
	 * 这样的首尾交替规律的拼接方式
	 */
	Mode 		    			int 
	AutoTimeStamp				bool
	Breaking 	    			[]byte /*戳与数据间的分隔符，可以为nil*/
	Stamps 		    			[][]byte /*允许输入多个，会按顺序依次拼接*/
	Raws 		    			chan []byte /*从主线程发来的信号队列，就像Qt的信号与槽*/

	News_Heads 		    		chan []byte
	News_Tails 		    		chan []byte
	News_HeadsAndTails 		    chan []byte
}

func (p *StampsConfig)Name()string{
	return STAMPS_RIVERNODE_NAME
}



type Stamps struct{
	bytesHandler	*bytes.Buffer
	tailHandler		*bytes.Buffer
	config 			*StampsConfig

	event_run     	Event

	stop			chan struct{}
}

func (p *Stamps)Name()string{
	return STAMPS_RIVERNODE_NAME
}

func (p *Stamps)Construct(stampsConfigAbs Config) error{
	if stampsConfigAbs.Name() != STAMPS_RIVERNODE_NAME {
		return errors.New(fmt.Sprintf("[%s] init error, config must StampsConfig",p.Name()))
	}


	v := reflect.ValueOf(stampsConfigAbs)
	c := v.Interface().(*StampsConfig)


	if c.Breaking == nil || c.Stamps == nil {
		return errors.New(fmt.Sprintf("[%s] init error, breaking or stamps is nil",p.Name()))
	}

	if c.Events == nil || c.Errors ==nil{
		return errors.New(fmt.Sprintf("[%s] init error, Events or Errors is nil",p.Name()))
	}

	if c.Raws == nil {
		return errors.New(fmt.Sprintf("[%s] init error, Raws is nil",p.Name()))
	}

	if c.News_Heads !=nil||c.News_Tails !=nil||c.News_HeadsAndTails !=nil{
		return errors.New(fmt.Sprintf("[%s] init error, News_Heads or News_Tails or News_HeadsAndTails ",p.Name()))
	}

	if len(c.Breaking)<3{
		return errors.New(fmt.Sprintf("[%s] init error, Breaking is to short, please len(Breaking) >= 3",p.Name()))
	}

	if c.Mode != HEADSANDTAILS && c.Mode != HEADS && c.Mode != TAILS{
		return errors.New(fmt.Sprintf("[%s] init error, unknown mode",p.Name()))
	}
	
	if c.Mode == HEADSANDTAILS && len(c.Stamps)<2{
		return errors.New(fmt.Sprintf("[%s] init error, mode is headandtail but only one stamp",p.Name()))
	}

	p.config = c

	p.bytesHandler = bytes.NewBuffer([]byte{})

	if(p.config.Mode == HEADSANDTAILS) { p.tailHandler = bytes.NewBuffer([]byte{}) } 


	timeStampStr :=""
	modeStr 	 :=""

	if p.config.AutoTimeStamp{
		timeStampStr = "不会首先自动添加时间戳"
	}else{
		timeStampStr = "会首先自动添加时间戳"
	}

	if p.config.Mode == HEADS{
		modeStr ="将某个或某些印章戳添加于数据头部"
		p.event_run = NewEvent(STAMPS_RUN, p.config.UniqueId, "",fmt.Sprintf("[heads %s]开始运行，"+
						  "其UniqueId为%s,Mode为%s并且%s。",p.Name(), p.config.UniqueId, modeStr,timeStampStr))
		p.config.News_Heads				=make(chan []byte)
	}else if p.config.Mode == TAILS{
		modeStr ="将某个或某些印章戳添加于数据尾部"
		p.event_run = NewEvent(STAMPS_RUN, p.config.UniqueId, "",fmt.Sprintf("[tails %s]开始运行，"+
						  "其UniqueId为%s,Mode为%s并且%s。",p.Name(), p.config.UniqueId, modeStr,timeStampStr))
		p.config.News_Tails				=make(chan []byte)
	}else if p.config.Mode == HEADSANDTAILS{
		modeStr ="将某些印章戳按照奇偶顺序依次添加于数据头部与尾部"
		p.event_run = NewEvent(STAMPS_RUN, p.config.UniqueId, "",fmt.Sprintf("[headandtails %s]开始运行，"+
						  "其UniqueId为%s,Mode为%s并且%s。",p.Name(), p.config.UniqueId, modeStr,timeStampStr))
		p.config.News_HeadsAndTails		=make(chan []byte)	
	}
	
	p.stop				=make(chan struct{})

	return nil
}


func (p *Stamps)Run(){
	p.config.Events <-p.event_run

	switch p.config.Mode{
	case HEADS:
		go func(){
			defer p.reactiveDestruct()
			for{
				select{
				case raw, ok := <-p.config.Raws:
					if !ok{ return }
					p.stampToHead(raw)
				case <-p.stop:
					return
				}
			}
		}()
	case TAILS:
		go func(){
			defer p.reactiveDestruct()
			for{
				select{
				case raw, ok := <-p.config.Raws:
					if !ok{ return }
					p.stampToTail(raw)
				case <-p.stop:
					return
				}
			}
		}()
	case HEADSANDTAILS:
		go func(){
			defer p.reactiveDestruct()
			for{
				select{
				case raw, ok := <-p.config.Raws:
					if !ok{ return }
					p.stampToHeadAndTail(raw)
				case <-p.stop:
					return
				}
			}
		}()
	}
}

func (p *Stamps)ProactiveDestruct(){
	p.config.Events <-NewEvent(STAMPS_PROACTIVE_DESTRUCT,p.config.UniqueId,"",
		    fmt.Sprintf("注意，由于某些原因[%s]主动调用了显式析构方法", p.Name()))

	p.stop<-struct{}{}	
}


func (p *Stamps)reactiveDestruct(){
	p.config.Events <-NewEvent(STAMPS_REACTIVE_DESTRUCT,p.config.UniqueId,"",
		fmt.Sprintf("[%s]触发了隐式析构方法",p.Name()))	

	switch p.config.Mode{
	case HEADS:
		close(p.config.News_Heads)
	case TAILS:
		close(p.config.News_Tails)
	case HEADSANDTAILS:
		close(p.config.News_HeadsAndTails)
	}

	close(p.stop)
}


func NewStamps() NodeAbstract {
	return &Stamps{}
}


func init() {
	Register(STAMPS_RIVERNODE_NAME, NewStamps)
	logger.Info(fmt.Sprintf("预加载完成，[%s]已预加载至package river_node.Nodes结构内",STAMPS_RIVERNODE_NAME))
}



/*------------以下是所需的功能方法-------------*/



func (p *Stamps)stampToHead(raw []byte){
	p.bytesHandler.Reset()
	for index, stamp :=range p.config.Stamps{
		if index ==0&&p.config.AutoTimeStamp{
			p.bytesHandler.Write(NanoTimeStamp())
			p.bytesHandler.Write(p.config.Breaking)
		}
		p.bytesHandler.Write(stamp)
		p.bytesHandler.Write(p.config.Breaking)
	}

	p.bytesHandler.Write(raw)

	p.config.News_Heads <-append([]byte{},p.bytesHandler.Bytes()...)
}

func (p *Stamps)stampToTail(raw []byte){
	p.bytesHandler.Reset()

	p.bytesHandler.Write(raw)

	for index, stamp :=range p.config.Stamps{
		if index ==0&&p.config.AutoTimeStamp{
			p.bytesHandler.Write(NanoTimeStamp())
			p.bytesHandler.Write(p.config.Breaking)
		}
		p.bytesHandler.Write(p.config.Breaking)
		p.bytesHandler.Write(stamp)
	}

	p.config.News_Tails <-append([]byte{},p.bytesHandler.Bytes()...)
}

func (p *Stamps)stampToHeadAndTail(raw []byte){

	p.bytesHandler.Reset();    ishead :=true;    p.tailHandler.Reset()

	for index, stamp :=range p.config.Stamps{
		if index ==0&&p.config.AutoTimeStamp{
			p.bytesHandler.Write(NanoTimeStamp())
			p.bytesHandler.Write(p.config.Breaking)
		}

		if ishead{
			p.bytesHandler.Write(stamp)
			p.bytesHandler.Write(p.config.Breaking)				
		}else{
			p.tailHandler.Write(p.config.Breaking)
			p.tailHandler.Write(stamp)
		}

		ishead =!ishead	
	}

	p.bytesHandler.Write(raw);    p.bytesHandler.Write(p.tailHandler.Bytes())

	p.config.News_HeadsAndTails <-append([]byte{}, p.bytesHandler.Bytes()...)
}



