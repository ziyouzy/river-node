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
	"river-node"
	"river-node/define"
	"river-node/anotherexample/another"

	"fmt"
	"math/rand"
	"time"
	"reflect"
	"errors"
)


const RIVER_NODE_NAME2 = "anotherexample"


/*范例Config*/
type AnotherExmaple2Config struct{

	UniqueId string	/*其所属上层数据通道(如Conn)的唯一识别标识*/

	Mode int

	EventChan chan int /*发送给主进程的信号队列，就像Qt的信号与槽*/

	RawinChan chan another.AnotherFunc

	NewoutChan chan []byte
}

func (p *AnotherExmaple2Config)Name()string{
	return RIVER_NODE_NAME2
}

type AnotherExmaple2 struct{
	rand int
	config *AnotherExmaple2Config
}

func (p *AnotherExmaple2)Name()string{
	return RIVER_NODE_NAME2
}

func (p *AnotherExmaple2)Init(anotherExmaple2ConfigAbs river_node.Config) error{
	if anotherExmaple2ConfigAbs.Name() != RIVER_NODE_NAME2 {
		return errors.New("anotherexmaple2 river-node init error, config must AnotherExmaple2Config")
	}


	value := reflect.ValueOf(anotherExmaple2ConfigAbs)
	config := value.Interface().(*AnotherExmaple2Config)


	if config.eventChan == nil||config.rawChan == nil||config.newChan == nil{
		return errors.New("anotherexmaple2 river-node init error, slotChan or eventChan "+
		                  "or newChan is nil")
	}

	p.config = config

	rand.Seed(time.Now().UnixNano())
	p.rand = rand.Int()


	switch p.config.mode{
	case TEST1:
		fmt.Println("type is anotherexample2, mode is TEST1")
	case TEST2:
		fmt.Println("type is anotherexample2, mode is TEST2")
	case TEST3:
		fmt.Println("type is anotherexample2, mode is TEST3")
	default:
		return errors.New("anotherexmaple2 river-node init error, unknown mode")
	}
	
	return nil
}

func (p *AnotherExmaple2)Run(){
	switch p.config.mode{
	case TEST2:
		go func(){
			for rawf := range p.config.rawChan{

				/*仅仅作为示范：*/
				p.config.eventChan<-define.ANOTHEREXAMPLE_TEST2

				rawf(append([]byte("测试2(TEST2)"), 0x12,0x33,0xff))

				/** []byte(p.uniqueId))是将一个string类型强制转换成[]byte
				 * 你不用担心原始string里的“汉字”是否会被丢弃
				 * 这类问题怎么说呢，显示会出问题，储存一般不会出问题
				 * 如果真遇到了相关问题，如fmt.Println打印出了乱码
				 * 首先去考虑下显示方式是否错了
				 * 比如你for{}了一个string
				 * 或者for range{}一个包含“汉字”的[]byte
				 */
				p.config.newChan<-append([]byte("test2"),[]byte(p.config.uniqueId)...)
			}
		}()
	default:
		p.config.eventChan<-define.ANOTHEREXAMPLE_ERR
	}
}



func NewAnotherExmaple2() river_node.NodeAbstract {
	return &AnotherExmaple2{}
}


func init() {
	river_node.Register(RIVER_NODE_NAME2, NewAnotherExmaple2)
}

