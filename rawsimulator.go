//在此实现一个用来产生临时数据的river-node
//也可以借此测试下自定义river-node的实战可行性与创建&使用方式
package river_node

import (
	"github.com/ziyouzy/logger"
	
	"fmt"
	"time"
	"reflect"
	"errors"
)


const RAWSIMULATOR_RIVERNODE_NAME = "testdatacreater数据源模拟器"
 

type RawSimulatorConfig struct{
	UniqueId 				string	/*其所属上层Conn的唯一识别标识*/
	Events 					chan Event /*发送给主进程的信号队列，就像Qt的信号与槽*/
		 
	StepSec					time.Duration

	News		 			chan []byte
}

func (p *RawSimulatorConfig)Name()string{
	return RAWSIMULATOR_RIVERNODE_NAME
}

type RawSimulator struct{
	sourceTable					[][]byte     
	config 						*RawSimulatorConfig

	rawsimulator_event_run 		Event
	stop                		chan struct{} 
}

func (p *RawSimulator)Name()string{
	return RAWSIMULATOR_RIVERNODE_NAME
}

func (p *RawSimulator)Construct(rawSimulatorConfigAbs Config) error{
	if rawSimulatorConfigAbs.Name() != RAWSIMULATOR_RIVERNODE_NAME {
		return errors.New(fmt.Sprintf("[%s] init error, config must RawSimulatorConfig",p.Name()))
	}


	v := reflect.ValueOf(rawSimulatorConfigAbs)
	c := v.Interface().(*RawSimulatorConfig)


	if c.UniqueId == ""{
		return errors.New(fmt.Sprintf("[%s] init error uniqueId is nil",p.Name()))
	}

	if c.Events == nil{
		return errors.New(fmt.Sprintf("[%s] init error, Events is nil",p.Name()))
	}

	if c.StepSec == (0 * time.Second) || c.News != nil{
		return errors.New(fmt.Sprintf("[%s] init error, StepSec or News is not nil",p.Name()))
	}
	
	
	p.config = c

	// p.sourceTable = [][]byte{
	// 	[]byte{0x01, 0x02, 0x03, 0x04,},
	// 	[]byte{0x05, 0x06, 0x07, 0x08,},
	// 	[]byte{0x01, 0x04, 0x09, 0x13,},
	// 	[]byte{0xF1,0x05,0x00,0x00,0xFF,0x00,},
	// 	[]byte{0xF1,0x05,0x00,0x00,0x00,0x00,},}

	p.sourceTable =	[][]byte{{0xF1,0x01,0x00,0x00,0x00,0x08, 0x29,0x3C},
		{0xF1,0x02,0x00,0x20,0x00,0x08,0x6C,0xF6},}

	p.rawsimulator_event_run = NewEvent(RAWSIMULATOR_RUN,p.config.UniqueId,"",
							   fmt.Sprintf("开始借助[%s]进行测试，此包的作用是将一个临时的[]byte数据源river-node化",p.Name()))

	p.config.News				= make(chan []byte)
	p.stop						= make(chan struct{})
	
	return nil
}


func (p *RawSimulator)Run(){
	p.config.Events <-p.rawsimulator_event_run

	go func(){
		defer p.reactiveDestruct()
		for i:=0;i<=len(p.sourceTable);i++{
			select {

			case <-p.stop:
				return

			default:
				if i == len(p.sourceTable){/*time.Sleep(10000*time.Second);*/i = 0}

				p.config.News <- p.sourceTable[i]
	
				time.Sleep(p.config.StepSec)
			}
		}
	}()	
}

//作为数据源，这个方法到时有很多合理的使用场景
func (p *RawSimulator)ProactiveDestruct(){
	p.config.Events <-NewEvent(RAWSIMULATOR_PROACTIVE_DESTRUCT,p.config.UniqueId,"",
	 				  fmt.Sprintf("注意，由于某些原因[%s]主动调用了显式析构方法",p.Name()))
	p.stop <-struct{}{}
}

func (p *RawSimulator)reactiveDestruct(){
	p.config.Events <-NewEvent(RAWSIMULATOR_REACTIVE_DESTRUCT,p.config.UniqueId,"",
	 				  fmt.Sprintf("[%s]触发了隐式析构方法",p.Name()))

	close(p.config.News)
	close(p.stop) 
}

func NewRawSimulator() NodeAbstract {
	return &RawSimulator{}
}


func init() {
	Register(RAWSIMULATOR_RIVERNODE_NAME, NewRawSimulator)
	logger.Info(fmt.Sprintf("预加载完成，[%s]已预加载至package river_node.Nodes结构内",p.Name())
}