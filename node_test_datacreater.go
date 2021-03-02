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


const TEST_DATACREATER_RIVERNODE_NAME = "testdatacreater"
 

type TestDataCreaterConfig struct{
	UniqueId 		string	/*其所属上层Conn的唯一识别标识*/
	Signals 		chan Signal /*发送给主进程的信号队列，就像Qt的信号与槽*/
	//Errors 			chan error
		 
	StepSec			time.Duration

	News 			chan []byte
}

func (p *TestDataCreaterConfig)Name()string{
	return TEST_DATACREATER_RIVERNODE_NAME
}

type TestDataCreater struct{
	sourceTable			[][]byte     
	config 				*TestDataCreaterConfig
}

func (p *TestDataCreater)Name()string{
	return TEST_DATACREATER_RIVERNODE_NAME
}

func (p *TestDataCreater)Init(testDataCreaterConfigAbs Config) error{
	if testDataCreaterConfigAbs.Name() != TEST_DATACREATER_RIVERNODE_NAME {
		return errors.New("testdatacreator river-node init error, "+ 
		                  "config must TestDataCreaterConfig")
	}


	v := reflect.ValueOf(testDataCreaterConfigAbs)
	c := v.Interface().(*TestDataCreaterConfig)


	if c.UniqueId == ""{
		return errors.New("testdatacreater river-node init error uniqueId is nil")
	}

	if c.Signals == nil /*|| c.Errors == nil*/{
		return errors.New("testdatacreater river-node init error, "+
						  "Signals is nil")
	}

	if c.StepSec == (0 * time.Second) || c.News == nil{
		return errors.New("testdatacreater river-node init error, "+
						  "StepSec or News is nil")
	}
	
	
	p.config = c

	p.sourceTable = [][]byte{
		[]byte{0x01, 0x02, 0x03, 0x04},
		[]byte{0x05, 0x06, 0x07, 0x08},
		[]byte{0x01, 0x04, 0x09, 0x13},}
	
	return nil
}


var (
	test_datacreater_signal_run Signal
)

func (p *TestDataCreater)Run(){
	test_datacreater_signal_run = NewSignal(TESTDATACREATER_RUN,p.config.UniqueId,
		"开始借助package testdatacreater进行测试，"+
		"此包的作用是将一个临时的[]byte数据源river-node化")
	p.config.Signals <- test_datacreater_signal_run

	go func(){
		for i:=0;i<=len(p.sourceTable);i++{
			if i == len(p.sourceTable){i = 0}

			p.config.News <- p.sourceTable[i]

			time.Sleep(p.config.StepSec)
		}
	}()	
}



func NewTestDataCreater() NodeAbstract {
	return &TestDataCreater{}
}


func init() {
	Register(TEST_DATACREATER_RIVERNODE_NAME, NewTestDataCreater)
	logger.Info("预加载完成，testdatacreator适配器已预加载至package river_node.Nodes结构内")
}