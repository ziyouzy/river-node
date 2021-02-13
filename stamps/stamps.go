package stamps

import (

	"zadapter"
	
	/*暂时没有需要返回给上层的signal内容*/
	//"zadapter/define"

	//"fmt"
	"bytes"
	"reflect"
	"errors"
)

const ADAPTER_NAME = "stamps"



type StampsConfig struct{

	UniqueId string	/*其所属上层Conn的唯一识别标识*/
	
	/** 分为三种，HEAD、TAIL、HEADANDTAIL
	 * 当是HEADANDTAIL模式，切len(stamp)>1时
	 * 按照如下顺序拼接：
	 * stamp1+raw(首部);raw+stamp2(尾部);stamp3+raw(首部);raw+stamp4(尾部);stamp5+raw(首部);....
	 * 这样的首尾交替规律的拼接方式
	 */
	Mode int 

	Breaking []byte /*戳与数据间的分隔符，可以为nil*/

	Stamps [][]byte /*允许输入多个，会按顺序依次拼接*/

	SignalChan chan int /*发送给主进程的信号队列，就像Qt的信号与槽*/

	RawinChan chan []byte /*从主线程发来的信号队列，就像Qt的信号与槽*/

	NewoutChan chan []byte /*校验通过切去掉校验码的新切片*/
}

func (p *StampsConfig)Name()string{
	return ADAPTER_NAME
}



type Stamps struct{
	bytesHandler *bytes.Buffer
	tailHandler *bytes.Buffer
	config *StampsConfig
}

func (p *Stamps)Name()string{
	return ADAPTER_NAME
}

func (p *Stamps)Init(stampsConfigAbs zadapter.Config) error{
	if stampsConfigAbs.Name() != ADAPTER_NAME {
		return errors.New("stamps adapter init error, config must StampsConfig")
	}


	vs := reflect.ValueOf(stampsConfigAbs)
	cs := vs.Interface().(*StampsConfig)


	if cs.breaking == nil || cs.stamps == nil {
		return errors.New("stamps adapter init error, breaking or stamps is nil")
	}

	/*此适配器的signalChan可以为空*/
	if cs.rawChan == nil || cs.newChan ==nil{
		return errors.New("stamps adapter init error, slotChan or signalChan "+
		                  "or newChan is nil")
	}
	
	
	p.config = cs

	p.bytesHandler = bytes.NewBuffer([]byte{})

	if(p.config.mode == HEADANDTAIL) { p.tailHandler = bytes.NewBuffer([]byte{}) } 

	return nil
}


/** Run()必须确保不返回error、不fmt错误
 * 这些问题都要在Init()函数内实现纠错与警告
 * Run()的一切问题都通过signal的方式传递至管道
 */
func (p *Stamps)Run(){
	switch p.config.mode{
	case HEAD:
		go func(){
			for raw := range p.config.rawChan{
				p.stampToHead(raw)
			}
		}()
	case TAIL:
		go func(){
			for raw := range p.config.rawChan{
				p.stampToTail(raw)
			}
		}()
	case HEADANDTAIL:
		go func(){
			for raw := range p.config.rawChan{
				p.stampToHeadAndTail(raw)
			}
		}()
	}
}



/** 由于通信管道的存在似乎就无法为其设计初始化默认值的操作了
 * 但这个函数也是必须的，因为上层一定会遍历各个map
 * 从而识别并确认都有哪些已经注册并在册的预编译适配器
 */

func NewStamps() zadapter.AdapterAbstract {
	return &Stamps{}
}


func init() {
	zadapter.Register(ADAPTER_NAME, NewStamps)
}




/*------------以下是所需的功能方法-------------*/



func (p *Stamps)stampToHead(raw []byte){
	p.bytesHandler.Reset()
	for _, stamp :=range p.config.stamps{
		p.bytesHandler.Write(stamp)
		p.bytesHandler.Write(p.config.breaking)
	}

	p.bytesHandler.Write(raw)

	p.config.newChan <-p.bytesHandler.Bytes()
}

func (p *Stamps)stampToTail(raw []byte){
	p.bytesHandler.Reset()

	p.bytesHandler.Write(raw)

	for _, stamp :=range p.config.stamps{
		p.bytesHandler.Write(p.config.breaking)
		p.bytesHandler.Write(stamp)
	}

	p.config.newChan <-p.bytesHandler.Bytes()
}

func (p *Stamps)stampToHeadAndTail(raw []byte){

	p.bytesHandler.Reset();    ishead :=true;    p.tailHandler.Reset()

	for _, stamp :=range p.config.stamps{
		if ishead{
			p.bytesHandler.Write(stamp)
			p.bytesHandler.Write(p.config.breaking)				
		}else{
			p.tailHandler.Write(p.config.breaking)
			p.tailHandler.Write(stamp)
		}

		ishead =!ishead	
	}

	p.bytesHandler.Write(raw);    p.bytesHandler.Write(p.tailHandler.Bytes())

	p.config.newChan <-p.bytesHandler.Bytes()
}

