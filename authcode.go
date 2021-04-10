/*对进行river-node适配封装*/

package river_node

import (
	"github.com/ziyouzy/logger"

	"github.com/ziyouzy/go-authcode"

	"fmt"
	"errors"
	"reflect"
)



const AUTHCODE_NODE_NAME = "authcode适配器"

type AuthCodeConfig struct{
	UniqueId 		  			string	/*其所属上层数据通道(如Conn)的唯一识别标识*/
	Events 		  				chan Event /*发送给主进程的信号队列，就像Qt的信号与槽*/
	Errors 			  			chan error

	Mode 			  			int /*define.ENCODE或DECODE*/
	
	Raws 		      			chan []byte /*从主线程发来的信号队列，就像Qt的信号与槽*/

	AuthCode_Key				string
	AuthCode_DynamicKeyLen		int
	AuthCode_ExpirySec			int

	News_Encode					chan []byte
	Limit_Encode				int
	//----------
	News_Decode					chan []byte
	Limit_Decode      			int

	//会使用"github.com/ziyouzy/go-authcode"
	//无论编码函数还是解码函数都可能返回error
	//因此需要同时拥有Limit_Encode与Limit_Decode这两个字段与其对应的功能
}




func (p *AuthCodeConfig)Name()string{
	return AUTHCODE_NODE_NAME
}



type AuthCode struct{
	config 				*AuthCodeConfig

	authcode			*go_authcode.AuthCode
	countor 			int
	event_run 			Event
	event_fused 		Event
}

func (p *AuthCode)Name()string{
	return AUTHCODE_NODE_NAME
}

func (p *AuthCode)Construct(AuthCodeConfigAbs Config) error{
	if AuthCodeConfigAbs.Name() != AUTHCODE_NODE_NAME {
		return errors.New(fmt.Sprintf("[%s] init error, config must AuthCodeConfig",p.Name()))
	}


	v := reflect.ValueOf(AuthCodeConfigAbs)
	c := v.Interface().(*AuthCodeConfig)


	if c.UniqueId == "" {
		return errors.New(fmt.Sprintf("[%s] init error, uniqueId is nil",p.Name()))
	}

	if c.Events == nil || c.Errors == nil || c.Raws == nil{
		return errors.New(fmt.Sprintf("[%s] init error, Events or Errors or Raws is nil",p.Name()))
	}

	if c.News_Encode != nil || c.News_Decode != nil{
		return errors.New(fmt.Sprintf("[%s] init error, News_Encode and News_Decode must nil",p.Name()))
	}

	if c.Mode != ENCODE&&c.Mode != DECODE {
		return errors.New(fmt.Sprintf("[%s] init error, unknown mode",p.Name())) 
	}

	if c.AuthCode_Key ==""{
		return errors.New(fmt.Sprintf("[%s] init error, Salt is nil",p.Name())) 
	}

	//AuthCode_ExpirySec与AuthCode_DynamicKeyLen都可以为0，但是不推荐
	
	if c.Mode ==DECODE && (c.Limit_Decode ==0){
		return errors.New(fmt.Sprintf("[decode %s] init error, Limit_Decode is nil",p.Name())) 
	}

	if c.Mode ==ENCODE && (c.Limit_Encode ==0){
		return errors.New(fmt.Sprintf("[encode %s] init error, Limit_Encode is nil",p.Name())) 
	}

	
	p.config = c

	p.authcode = go_authcode.New(p.config.AuthCode_Key, p.config.AuthCode_DynamicKeyLen, p.config.AuthCode_ExpirySec)

	p.event_fused = NewEvent(AUTHCODE_FUSED, c.UniqueId, "", "")

	modeStr := ""

	if p.config.Mode == ENCODE{
		modeStr ="authcode为加密模式，将加密后的数据注入News_Encode管道"
		p.event_run = NewEvent(AUTHCODE_RUN,p.config.UniqueId,"",fmt.Sprintf("[encode %s]开始运行，其UniqueId为%s, Mode为:%s",
			   		  p.Name(), p.config.UniqueId, modeStr))

		p.config.News_Encode		= make(chan []byte)

	}else if p.config.Mode == DECODE{
		modeStr ="authcode为解密模式，将解密后的数据注入News_Decode管道"
		p.event_run = NewEvent(AUTHCODE_RUN,p.config.UniqueId,"",fmt.Sprintf("[decode %s]开始运行，其UniqueId为%s, Mode为:%s",
					  p.Name(), p.config.UniqueId, modeStr))


		p.config.News_Decode 		= make(chan []byte)
	}

	return nil
}



func (p *AuthCode)Run(){

	p.config.Events <-p.event_run

	switch p.config.Mode{
	case ENCODE:
		go func(){
			defer p.reactiveDestruct()
			for {
				select{
				case raw, ok := <-p.config.Raws:
					if !ok{
						return
					}else{
						p.encode(raw)
					}
				}
			}
		}()
	case DECODE:
		go func(){
			defer p.reactiveDestruct()
			for {
				select{
				case raw, ok := <-p.config.Raws:
					if !ok{
						return
					}else{
						p.decode(raw)
					}
				}
			}
		}()
	}
}


//被动 - reactive
//被动析构是检测到Raws被上层关闭后的响应式析构操作
func (p *AuthCode)reactiveDestruct(){
	p.config.Events <-NewEvent(AUTHCODE_REACTIVE_DESTRUCT,p.config.UniqueId,"",fmt.Sprintf("[%s]触发了隐式析构方法",p.Name()))

	//析构数据源
	p.authcode.CloseSafe()

	switch p.config.Mode{
	case ENCODE:
		close(p.config.News_Encode)
	case DECODE:
		close(p.config.News_Decode)
	}
}



func NewAuthCode() NodeAbstract {
	return &AuthCode{}
}


func init() {
	Register(AUTHCODE_NODE_NAME, NewAuthCode)
	logger.Info(fmt.Sprintf("预加载完成，[%s]已预加载至package river_node.RNodes结构内",AUTHCODE_NODE_NAME))
}

/*------------以下是所需的功能方法-------------*/
func (p *AuthCode)encode(baits []byte){
	if res ,err := p.authcode.Encode(baits,""); err !=nil{
		p.countor++

		if p.countor > p.config.Limit_Encode{
			p.config.Errors <-NewError(AUTHCODE_ENCODE_FAIL, p.config.UniqueId, fmt.Sprintf("%x",baits),
							fmt.Sprintf("[%s]连续%d次加密失败，已超过系统设定的最大次数，系统设定的最大连续失败"+
								"次数为%d,[报错内容为%s]",p.Name(),p.countor,p.config.Limit_Encode,err.Error()))
			p.countor =0
			p.config.Events <-p.event_fused
		}else{
			p.config.Errors <-NewError(AUTHCODE_ENCODE_FAIL, p.config.UniqueId, fmt.Sprintf("%x",baits),
							fmt.Sprintf("[%s]连续第%d次加密失败，当前系统设定的最大连续失败次数为%d,[报错内容为%s]",
				   				p.Name(),p.countor, p.config.Limit_Encode,err.Error()))
		}
		
	}else{
		if p.countor !=0{
			p.config.Events <-NewEvent(AUTHCODE_ENCODE_RECOVERED,p.config.UniqueId,"", fmt.Sprintf("[%s]已从第%d次加密失败中恢复，"+
							"当前系统设定的最大失败次数为%d",p.Name(),p.countor,p.config.Limit_Encode))
			p.countor =0
		}
		p.config.News_Encode <- res
	}
}

func (p *AuthCode)decode(baits []byte){
	if res ,err := p.authcode.Decode(baits,""); err !=nil{
		p.countor++

		if p.countor > p.config.Limit_Decode{
			p.config.Errors <-NewError(AUTHCODE_DECODE_FAIL, p.config.UniqueId, fmt.Sprintf("%x",baits),
							fmt.Sprintf("[%s]连续%d次解密失败，已超过系统设定的最大次数，系统设定的最大连续失败"+
								"次数为%d,[报错内容为%s]",p.Name(), p.countor,p.config.Limit_Decode,err.Error()))
			p.countor =0
			p.config.Events <-p.event_fused
		}else{
			p.config.Errors <-NewError(AUTHCODE_DECODE_FAIL, p.config.UniqueId, fmt.Sprintf("%x",baits),
							fmt.Sprintf("[%s]连续%d次解密失败，当前系统设定的最大连续失败次数为%d,[报错内容为%s]",
				   				p.Name(), p.countor, p.config.Limit_Decode,err.Error()))
		}
		
	}else{
		if p.countor !=0{
			p.config.Events <-NewEvent(AUTHCODE_DECODE_RECOVERED,p.config.UniqueId,"", fmt.Sprintf("[%s]已从第%d次解密失败中恢复，"+
							"当前系统设定的最大失败次数为%d",p.Name(),p.countor,p.config.Limit_Decode))
			p.countor =0
		}
		p.config.News_Decode <- res
	}
}
