//此包的特殊之处在于，虽然错误次数无上限，但是所有过滤失败的内容都会返回%w类的错误而非%v
//作为一个很底层的river-node但是他往往都会与最上层的main层进行对接

//这是为了处理应对诸如“新usrio808发来连接后的首条信息这样的事件”

package river_node

import (
	"github.com/ziyouzy/logger"

	"fmt"
	"errors"
	"reflect"
	"bytes"
)



const BAITSFILTER_NODE_NAME = "baitsfilter过滤器"

type BaitsFilterConfig struct{
	UniqueId 		  			string	/*其所属上层数据通道(如Conn)的唯一识别标识*/
	Events 		  				chan EventAbs /*发送给主进程的信号队列，就像Qt的信号与槽*/
	Errors 			  			chan error

	Mode 			  			int /*define.KEEPHEAD或DROPHEAD*/
	Heads						[][]byte
	Len_max						int
	Len_min						int
	
	Raws 		      			chan []byte /*从主线程发来的信号队列，就像Qt的信号与槽*/

	News_KeepHead				chan []byte
	News_DropHead				chan []byte
}




func (p *BaitsFilterConfig)Name()string{
	return BAITSFILTER_NODE_NAME
}



type BaitsFilter struct{
	config 				*BaitsFilterConfig

	/*此包只用来作为纯粹的过滤操作，而对于Panich这样的逻辑应该留给HeartBeating去完成*/
	/*不过authcode与crc还是有必要设计出基于“失败次数”触发panich的业务逻辑的*/
	/*因为“次数超限”和“时间超限”是两种“本质不同”的意义体现*/
	/*对于BaitsFilter来说，“无上限原则”是他本应具备的属性*/
	//warpError_Panich	error
}

func (p *BaitsFilter)Name()string{
	return BAITSFILTER_NODE_NAME
}

func (p *BaitsFilter)Construct(BaitsFilterConfigAbs Config) error{
	if BaitsFilterConfigAbs.Name() != BAITSFILTER_NODE_NAME {
		return errors.New(
			fmt.Sprintf("[river-node type:%s] init error, config must BaitsFilterConfig",
			p.Name()))
	}

	v := reflect.ValueOf(BaitsFilterConfigAbs)
	c := v.Interface().(*BaitsFilterConfig)

	if c.UniqueId == "" {
		return errors.New(
			fmt.Sprintf("[river-node type:%s] init error, uniqueId is nil",
			p.Name()))
	}

	if c.Events == nil || c.Errors == nil || c.Raws == nil{
		return errors.New(
			fmt.Sprintf("[uid:%s] init error, Events or Errors or Raws is nil",
			c.UniqueId))
	}

	if c.News_KeepHead != nil || c.News_DropHead != nil{
		return errors.New(
			fmt.Sprintf("[uid:%s] init error, News_KeepHead and News_DropHead must nil",
			c.UniqueId))
	}

	if c.Mode != KEEPHEAD&&c.Mode != DROPHEAD {
		return errors.New(
			fmt.Sprintf("[uid:%s] init error, unknown mode",c.UniqueId)) 
	}

	if c.Len_min>c.Len_max{
		return errors.New(
			fmt.Sprintf("[uid:%s] init error, len_min > len_max!!!",c.UniqueId)) 
	}

	p.config = c	


	if p.config.Mode == KEEPHEAD{
		p.config.News_KeepHead = make(chan []byte)
	}else if p.config.Mode == DROPHEAD{
		p.config.News_DropHead = make(chan []byte)
	}

	return nil
}



func (p *BaitsFilter)Run(){
	fmt.Println("!!!!!baitsfilter_1!!!!!!")
	modeStr := ""
	fmt.Println("!!!!!baitsfilter_1!!!!!!")
	if p.config.Mode == KEEPHEAD{
		fmt.Println("!!!!!baitsfilter_KEEPHEAD_2!!!!!!")
		modeStr ="baitsfilter所适配的模式将保留用来判定/识别的head，数据会注入News_KeepHead管道"
		p.config.Events <-NewEvent(
			BAITSFILTER_RUN,p.config.UniqueId,"",nil,
			fmt.Sprintf("[mode:%s]节点开始运行",modeStr))
	}else if p.config.Mode == DROPHEAD{
		fmt.Println("!!!!!baitsfilter_DROPHEAD_2!!!!!!")
		modeStr ="baitsfilter所适配的模式将丢弃用来判定/识别的head，数据会注入News_DropHead管道"
		p.config.Events <-NewEvent(
			BAITSFILTER_RUN,p.config.UniqueId,"",nil,
			fmt.Sprintf("[mode:%s]节点开始运行", modeStr))
	}
	fmt.Println("!!!!!baitsfilter_3!!!!!!")

	switch p.config.Mode{
	case KEEPHEAD:
		go func(){
			defer p.reactiveDestruct()
			for {
				select{
				case raw, ok := <-p.config.Raws:
					if !ok{
						return
					}else{
						p.keepHead(raw)
					}
				}
			}
		}()
	case DROPHEAD:
		go func(){
			defer p.reactiveDestruct()
			for {
				select{
				case raw, ok := <-p.config.Raws:
					if !ok{
						return
					}else{
						p.dropHead(raw)
					}
				}
			}
		}()
	}
}



//被动 - reactive
//被动析构是检测到Raws被上层关闭后的响应式析构操作
func (p *BaitsFilter)reactiveDestruct(){
	//析构数据源
	switch p.config.Mode{
	case KEEPHEAD:
		close(p.config.News_KeepHead)
	case DROPHEAD:
		close(p.config.News_DropHead)
	}

	p.config.Events <-NewEvent(
		BAITSFILTER_REACTIVE_DESTRUCT,p.config.UniqueId,"",nil,
		"触发了隐式析构方法")
}



func NewBaitsFilter() NodeAbstract {
	return &BaitsFilter{}
}


func init() {
	Register(BAITSFILTER_NODE_NAME, NewBaitsFilter)
	logger.Info(
		fmt.Sprintf("预加载完成，[river-node type:%s]已预加载至package river_node.RNodes结构内",
		BAITSFILTER_NODE_NAME))
}

/*------------以下是所需的功能方法-------------*/
func (p *BaitsFilter)keepHead(baits []byte){
	if !p.lenAuth(baits){return}

	for _, head := range p.config.Heads{
		//子字符串首次出现的位置，没有则返回-1，有则从零开始汇报发现的起始位置
		if bytes.Index(baits,head)==0{
			p.config.News_KeepHead <-baits
			return
		}
	}

	p.config.Errors <-fmt.Errorf(
		"%w",NewEvent(BAITSFILTER_HEADAUTHFAIL, p.config.UniqueId, fmt.Sprintf("%x",baits),nil,
		"发现了报头未知的Baits"))

}

func (p *BaitsFilter)dropHead(baits []byte){
	if !p.lenAuth(baits){return}

	for _, head := range p.config.Heads{
		//子字符串首次出现的位置，没有则返回-1，有则从零开始汇报发现的起始位置
		if bytes.Index(baits, head)==0{
			//纯粹的删除左侧某些字段
			baits =baits[len(head):]
			p.config.News_DropHead <-baits
			return
		}
	}

	p.config.Errors <-fmt.Errorf(
		"%w",NewEvent(BAITSFILTER_HEADAUTHFAIL, p.config.UniqueId, fmt.Sprintf("%x",baits),nil,
		"发现了报头未知的Baits"))
}

//注意，判断是的整体baits的长度，而不是head的长度
func (p *BaitsFilter)lenAuth(baits []byte)bool{
	l :=len(baits)
	if p.config.Len_max ==0&&p.config.Len_min ==0{return true}

	if p.config.Len_max >= l&&p.config.Len_min <= l{
		return true
	}else{
		p.config.Errors <-fmt.Errorf(
			"%w", NewEvent(BAITSFILTER_LENAUTHFAIL, p.config.UniqueId, fmt.Sprintf("%x", baits), 
			nil, fmt.Sprintf("发现了不符合长度标准的baits，当前设定的最小长度为%d,"+
					"最大长度为%d,然而baits长度为%d",p.config.Len_min, p.config.Len_max, l)))

		return false
	}
}

