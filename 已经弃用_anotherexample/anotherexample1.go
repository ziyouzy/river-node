/** 本包的目的在于演示如何设计一个非字节切片数据类型的适配器
 * 此包作为范例，其目标数据类型是一个自定义结构体
 * 而对于实际场景中实现功能时，各种类型都可以用类似的方式去设计
 * 最终完成最适合的某种适配器
 */

package anotherexample

import (
	/** 引入zadaptr包与another包都遵循了单向调用链原则
	 * 虽然文件夹部署上river-node文件夹在最外层
	 * anotherexample文件夹在中间层
	 * another文件夹在最内层
	 * 但是从包的相互引用逻辑上最末是anotherexample包
	 */

	"river-node/define"
	"river-node"
	"river-node/anotherexample/another"

	"math/rand"
	"time"
	"reflect"
	"errors"
	"fmt"
)


const RIVER_NODE_NAME1 = "anotherexample"


/*范例Config*/
type AnotherExmaple1Config struct{

	UniqueId string	/*其所属上层数据通道(如Conn)的唯一识别标识*/

	Mode int

	SignalChan chan int /*发送给主进程的信号队列，就像Qt的信号与槽*/

	RawinChan chan []byte

	NewoutChan chan *another.AnotherStruct
}



func (p *AnotherExmaple1Config)Name()string{
	return RIVER_NODE_NAME1
}



type AnotherExmaple1 struct{
	rand int
	config *AnotherExmaple1Config
}


func (p *AnotherExmaple1)Name()string{
	return RIVER_NODE_NAME1
}


func (p *AnotherExmaple1)Init(anotherExmaple1ConfigAbs river_node.Config) error{
	if anotherExmaple1ConfigAbs.Name() != RIVER_NODE_NAME1 {
		return errors.New("anotherexmaple1 river-node init error, config must AnotherExmaple1Config")
	}


	value := reflect.ValueOf(anotherExmaple1ConfigAbs)
	config := value.Interface().(*AnotherExmaple1Config)

	if config.signalChan == nil||config.rawChan == nil||config.newChan == nil{
		return errors.New("anotherexmaple1 river-node init error, slotChan or signalChan "+
		                  "or newChan is nil")
	}

	p.config = config


	rand.Seed(time.Now().UnixNano())
	p.rand = rand.Int()

	switch p.config.mode{
	case TEST1:
		fmt.Println("type is anotherexample1, mode is TEST1")
	case TEST2:
		fmt.Println("type is anotherexample1, mode is TEST2")
	case TEST3:
		fmt.Println("type is anotherexample1, mode is TEST3")
	default:
		return errors.New("anotherexmaple1 river-node init error, unknown mode")
	}
	
	return nil
}



func (p *AnotherExmaple1)Run(){
	switch p.config.mode{
	case TEST1:
		go func(){
			for sl := range p.config.rawChan{

				/*仅仅作为示范：*/
				p.config.signalChan<-define.ANOTHEREXAMPLE_TEST1
				p.config.newChan<-&another.AnotherStruct{
						TimeStamp : time.Now().UnixNano(),
						Tip : "仅仅作为示范(AnotherExmaple1)",
						TestSl: []float64{1.2,1.33,1,444},
						//rawSl: sl,//本来每次循环结束sl就会销毁，但是这样写sl就会持久化了
						RawSl: append([]byte{},sl...),//这样是进行深拷贝，sl不会在有外部引用他了，sl每次循环结束都被销毁
					}

			}
		}()
	default:
		/*byte为int8，rune为int32*/
		p.config.signalChan<-define.ANOTHEREXAMPLE_ERR
	}
}



func NewAnotherExmaple1() river_node.NodeAbstract {
	return &AnotherExmaple1{}
}


func init() {
	river_node.Register(RIVER_NODE_NAME1, NewAnotherExmaple1)
}

