//在此实现一个用来产生临时数据的river-node
//也可以借此测试下自定义river-node的实战可行性与创建&使用方式
package river_node

import (
	"github.com/ziyouzy/logger"
	
	//"fmt"
	"time"
	"reflect"
	"errors"
)


const RAWSIMULATOR_RIVERNODE_NAME = "testdatacreater"
 

type RawSimulatorConfig struct{
	UniqueId 				string	/*其所属上层Conn的唯一识别标识*/
	Events 					chan Event /*发送给主进程的信号队列，就像Qt的信号与槽*/
	//Errors 				chan error
		 
	StepSec					time.Duration

	News_ByteSlice 			chan []byte
}

func (p *RawSimulatorConfig)Name()string{
	return RAWSIMULATOR_RIVERNODE_NAME
}

type RawSimulator struct{
	sourceTable					[][]byte     
	config 						*RawSimulatorConfig

	rawsimulator_event_run 	Event
	stop                		chan struct{} 
}

func (p *RawSimulator)Name()string{
	return RAWSIMULATOR_RIVERNODE_NAME
}

func (p *RawSimulator)Construct(rawSimulatorConfigAbs Config) error{
	if rawSimulatorConfigAbs.Name() != RAWSIMULATOR_RIVERNODE_NAME {
		return errors.New("rawsimulator river-node init error, "+ 
		             "config must RawSimulatorConfig")
	}


	v := reflect.ValueOf(rawSimulatorConfigAbs)
	c := v.Interface().(*RawSimulatorConfig)


	if c.UniqueId == ""{
		return errors.New("rawsimulator river-node init error uniqueId is nil")
	}

	if c.Events == nil /*|| c.Errors == nil*/{
		return errors.New("rawsimulator river-node init error, Events is nil")
	}

	if c.StepSec == (0 * time.Second) || c.News_ByteSlice == nil{
		return errors.New("rawsimulator river-node init error, StepSec or News is nil")
	}
	
	
	p.config = c

	p.sourceTable = [][]byte{
		[]byte{0x01, 0x02, 0x03, 0x04},
		[]byte{0x05, 0x06, 0x07, 0x08},
		[]byte{0x01, 0x04, 0x09, 0x13},}

	p.rawsimulator_event_run = NewEvent(RAWSIMULATOR_RUN,p.config.UniqueId,
	 "开始借助package testdatacreater进行测试，此包的作用是将一个临时的[]byte数据源river-node化")

	p.stop =make(chan struct{})
	
	return nil
}


func (p *RawSimulator)Run(){
	p.config.Events <- p.rawsimulator_event_run

	go func(){
		defer p.reactiveDestruct()
		for i:=0;i<=len(p.sourceTable);i++{
			select {

			case <-p.stop:
				return

			default:
				if i == len(p.sourceTable){/*time.Sleep(10000*time.Second);*/i = 0}

				p.config.News_ByteSlice <- p.sourceTable[i]
	
				time.Sleep(p.config.StepSec)
			}
		}
	}()	
}

//作为数据源，这个方法到时有很多合理的使用场景
func (p *RawSimulator)ProactiveDestruct(){
	p.config.Events <-NewEvent(RAWSIMULATOR_PROACTIVEDESTRUCT,p.config.UniqueId,
	 "注意，由于某些原因数据源模拟器主动调用了显式析构方法")
	p.stop <-struct{}{}
}

func (p *RawSimulator)reactiveDestruct(){
	p.config.Events <-NewEvent(RAWSIMULATOR_REACTIVEDESTRUCT,p.config.UniqueId,
	 "数据源模拟器触发了隐式析构方法")

	//隐式析构的必要步骤，可触发下游river-node的析构责任链
	close(p.config.News_ByteSlice)
	//不关闭也行,如过整体对象被销毁了无论管道里有无数据也都会被销毁，但是close也算是万无一失
	close(p.stop) 
}

func NewRawSimulator() NodeAbstract {
	return &RawSimulator{}
}


func init() {
	Register(RAWSIMULATOR_RIVERNODE_NAME, NewRawSimulator)
	logger.Info("预加载完成，rawsimulator适配器已预加载至package river_node.Nodes结构内")
}