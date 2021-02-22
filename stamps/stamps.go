package stamps

import (
	"river-node"
	//"river-node/define"	/*暂时没有需要返回给上层的signal内容*/
	"river-node/logger"
	
	"bytes"
	"reflect"
	"errors"
)

const RIVER_NODE_NAME = "stamps"

var signal_run river_node.Signal

type StampsConfig struct{
	UniqueId 		string	/*其所属上层Conn的唯一识别标识*/
	 Signals 		chan int /*发送给主进程的信号队列，就像Qt的信号与槽*/
	  Errors 		chan error
	/** 分为三种，HEAD、TAIL、HEADANDTAIL
	 * 当是HEADANDTAIL模式，切len(stamp)>1时
	 * 按照如下顺序拼接：
	 * stamp1+raw(首部);raw+stamp2(尾部);stamp3+raw(首部);raw+stamp4(尾部);stamp5+raw(首部);....
	 * 这样的首尾交替规律的拼接方式
	 */
		Mode 		int 
	Breaking 		[]byte /*戳与数据间的分隔符，可以为nil*/
	  Stamps 		[][]byte /*允许输入多个，会按顺序依次拼接*/
		Raws 		chan []byte /*从主线程发来的信号队列，就像Qt的信号与槽*/


	    News 		chan []byte /*校验通过切去掉校验码的新切片*/
}

func (p *StampsConfig)Name()string{
	return RIVER_NODE_NAME
}



type Stamps struct{
	bytesHandler		*bytes.Buffer
	 tailHandler		*bytes.Buffer
	      config 		*StampsConfig
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


	if c.Signals == nil || cs.Errors ==nil{
		return errors.New("stamps river-node init error, Signals or Errors is nil")
	}

	if c.Raws == nil || cs.News ==nil{
		return errors.New("stamps river-node init error, Raws or News is nil")
	}
	
	
	p.config = c

	p.bytesHandler = bytes.NewBuffer([]byte{})

	if(p.config.Mode == HEADANDTAIL) { p.tailHandler = bytes.NewBuffer([]byte{}) } 

	
	signal_run = river_node.NewSignal(river_node.STAMPS_RUN, p.UniqueId)

	return nil
}


/** Run()必须确保不返回error、不fmt错误
 * 这些问题都要在Init()函数内实现纠错与警告
 * Run()的一切问题都通过signal的方式传递至管道
 */
func (p *Stamps)Run(){
	p.Config.Signals <- signal_run

	switch p.config.Mode{
	case HEAD:
		go func(){
			for raw := range p.config.Raws{
				p.stampToHead(raw)
			}
		}()
	case TAIL:
		go func(){
			for raw := range p.config.RawinChan{
				p.stampToTail(raw)
			}
		}()
	case HEADANDTAIL:
		go func(){
			for raw := range p.config.RawinChan{
				p.stampToHeadAndTail(raw)
			}
		}()
	}
}



/** 由于通信管道的存在似乎就无法为其设计初始化默认值的操作了
 * 但这个函数也是必须的，因为上层一定会遍历各个map
 * 从而识别并确认都有哪些已经注册并在册的预编译适配器
 */

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
	for _, stamp :=range p.config.Stamps{
		p.bytesHandler.Write(stamp)
		p.bytesHandler.Write(p.config.Breaking)
	}

	p.bytesHandler.Write(raw)

	p.config.News <-p.bytesHandler.Bytes()
}

func (p *Stamps)stampToTail(raw []byte){
	p.bytesHandler.Reset()

	p.bytesHandler.Write(raw)

	for _, stamp :=range p.config.Stamps{
		p.bytesHandler.Write(p.config.Breaking)
		p.bytesHandler.Write(stamp)
	}

	p.config.News <-p.bytesHandler.Bytes()
}

func (p *Stamps)stampToHeadAndTail(raw []byte){

	p.bytesHandler.Reset();    ishead :=true;    p.tailHandler.Reset()

	for _, stamp :=range p.config.Stamps{
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

