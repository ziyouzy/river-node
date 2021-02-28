package stamps

import (
	"river-node"
	//"river-node/define"	/*暂时没有需要返回给上层的signal内容*/
	"river-node/logger"
	
	"fmt"
	"bytes"
	"reflect"
	"errors"
)

const RIVER_NODE_NAME = "stamps"

type StampsConfig struct{
	UniqueId 	    string	/*其所属上层Conn的唯一识别标识*/
	Signals 	    chan river_node.Signal /*发送给主进程的信号队列，就像Qt的信号与槽*/
	Errors 		    chan error
	/** 分为三种，HEAD、TAIL、HEADANDTAIL
	 * 当是HEADANDTAIL模式，切len(stamp)>1时
	 * 按照如下顺序拼接：
	 * stamp1+raw(首部);raw+stamp2(尾部);stamp3+raw(首部);raw+stamp4(尾部);stamp5+raw(首部);....
	 * 这样的首尾交替规律的拼接方式
	 */
	Mode 		    int 
	AutoTimeStamp	bool
	Breaking 	    []byte /*戳与数据间的分隔符，可以为nil*/
	Stamps 		    [][]byte /*允许输入多个，会按顺序依次拼接*/
	Raws 		    chan []byte /*从主线程发来的信号队列，就像Qt的信号与槽*/

	News 		    chan []byte /*校验通过切去掉校验码的新切片*/
}

func (p *StampsConfig)Name()string{
	return RIVER_NODE_NAME
}



type Stamps struct{
	bytesHandler	*bytes.Buffer
	tailHandler		*bytes.Buffer
	config 			*StampsConfig
}

func (p *Stamps)Name()string{
	return RIVER_NODE_NAME
}

func (p *Stamps)Init(stampsConfigAbs river_node.Config) error{
	if stampsConfigAbs.Name() != RIVER_NODE_NAME {
		return errors.New("stamps river-node init error, config must StampsConfig")
	}


	v := reflect.ValueOf(stampsConfigAbs)
	c := v.Interface().(*StampsConfig)


	if c.Breaking == nil || c.Stamps == nil {
		return errors.New("stamps river-node init error, breaking or stamps is nil")
	}

	if c.Signals == nil || c.Errors ==nil{
		return errors.New("stamps river-node init error, Signals or Errors is nil")
	}

	if c.Raws == nil || c.News ==nil{
		return errors.New("stamps river-node init error, Raws or News is nil")
	}

	if c.Mode != HEADANDTAIL && c.Mode != HEAD && c.Mode != TAIL{
		return errors.New("stamps river-node init error, unknown mode")
	}
	
	if c.Mode == HEADANDTAIL && len(c.Stamps)<2{
		return errors.New("stamps river-node init error, mode is headandtail but only one stamp")
	}

	p.config = c

	p.bytesHandler = bytes.NewBuffer([]byte{})

	if(p.config.Mode == HEADANDTAIL) { p.tailHandler = bytes.NewBuffer([]byte{}) } 

	return nil
}


/** Run()必须确保不返回error、不fmt错误
 * 这些问题都要在Init()函数内实现纠错与警告
 * Run()的一切问题都通过signal的方式传递至管道
 */

var (
	signal_run river_node.Signal
	modeStr string
	timeStampStr string
)

func (p *Stamps)Run(){
	if p.config.AutoTimeStamp{
		timeStampStr = "不会首先自动添加时间戳"
	}else{
		timeStampStr = "会首先自动添加时间戳"
	}

	if p.config.Mode == HEAD{
		modeStr ="将某个或某些印章戳添加于数据头部"
	}else if p.config.Mode == TAIL{
		modeStr ="将某个或某些印章戳添加于数据尾部"
	}else if p.config.Mode == HEADANDTAIL{
		modeStr ="将某些印章戳按照奇偶顺序依次添加于数据头部与尾部"
	}
	
	signal_run = river_node.NewSignal(river_node.STAMPS_RUN, p.config.UniqueId, 
				 fmt.Sprintf("stamps适配器开始运行，其UniqueId为%s,Mode为%s并且%s。",
							 p.config.UniqueId, modeStr,timeStampStr))

	p.config.Signals <- signal_run

	switch p.config.Mode{
	case HEAD:
		go func(){
			for raw := range p.config.Raws{
				p.stampToHead(raw)
			}
		}()
	case TAIL:
		go func(){
			for raw := range p.config.Raws{
				p.stampToTail(raw)
			}
		}()
	case HEADANDTAIL:
		go func(){
			for raw := range p.config.Raws{
				p.stampToHeadAndTail(raw)
			}
		}()
	}
}




func NewStamps() river_node.NodeAbstract {
	return &Stamps{}
}


func init() {
	river_node.Register(RIVER_NODE_NAME, NewStamps)
	logger.Info("预加载完成，印章适配器已预加载至package river_node.Nodes结构内")
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

	p.config.News <-p.bytesHandler.Bytes()
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

	p.config.News <-p.bytesHandler.Bytes()
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

	p.config.News <-p.bytesHandler.Bytes()
}



