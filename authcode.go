/*对进行river-node适配封装*/

package river_node

import (
	"github.com/ziyouzy/logger"

	"github.com/ziyouzy/go-authcode"

	//"crypto/md5"
	//"encoding/base64"
	//"encoding/hex"
	"fmt"
	//"strconv"
	//"strings"
	"errors"
	"reflect"
	//"time"
)



const AUTHCODE_RIVERNODE_NAME = "authcode"

type AuthCodeConfig struct{
	UniqueId 		  			string	/*其所属上层数据通道(如Conn)的唯一识别标识*/
	Events 		  				chan Event /*发送给主进程的信号队列，就像Qt的信号与槽*/
	Errors 			  			chan error

	Mode 			  			int /*define.ENCODE或DECODE*/
	
	Raws 		      			<-chan []byte /*从主线程发来的信号队列，就像Qt的信号与槽*/

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
	return AUTHCODE_RIVERNODE_NAME
}



type AuthCode struct{
	config 				*AuthCodeConfig

	authcode			*go_authcode.AuthCode
	countor 			int
	event_run 			Event
	event_fused 		Event

	stop				chan struct{}
}

func (p *AuthCode)Name()string{
	return AUTHCODE_RIVERNODE_NAME
}

func (p *AuthCode)Construct(AuthCodeConfigAbs Config) error{
	if AuthCodeConfigAbs.Name() != AUTHCODE_RIVERNODE_NAME {
		return errors.New("crc river_node init error, config must CRCConfig")
	}


	v := reflect.ValueOf(AuthCodeConfigAbs)
	c := v.Interface().(*AuthCodeConfig)


	if c.UniqueId == "" {
		return errors.New("authcode river_node init error, uniqueId is nil")
	}

	if c.Events == nil || c.Errors == nil || c.Raws == nil{
		return errors.New("authcode river_node init error, Events or Errors or Raws is nil")
	}

	if c.News_Encode != nil || c.News_Decode != nil{
		return errors.New("authcode river-node init error, News_Encode and News_Decode must nil")
	}

	if c.Mode != ENCODE&&c.Mode != DECODE {
		return errors.New("authcode river-node init error, unknown mode") 
	}

	if c.AuthCode_Key ==""{
		return errors.New("authcode river-node init error, Salt is nil") 
	}

	//AuthCode_ExpirySec与AuthCode_DynamicKeyLen都可以为0，但是不推荐
	
	if c.Mode ==DECODE && (c.Limit_Decode ==0){
		return errors.New("decode mode authcode river-node init error, Limit_Decode is nil") 
	}

	if c.Mode ==ENCODE && (c.Limit_Encode ==0){
		return errors.New("encode mode authcode river-node init error, Limit_Encode is nil") 
	}

	
	p.config = c

	p.authcode = go_authcode.New(p.config.AuthCode_Key, p.config.AuthCode_DynamicKeyLen, p.config.AuthCode_ExpirySec)

	p.event_fused = NewEvent(AUTHCODE_FUSED, c.UniqueId, "", "")

	modeStr := ""

	if p.config.Mode == ENCODE{
		modeStr ="authcode为加密模式，将加密后的数据注入News_Encode管道"
		p.event_run = NewEvent(AUTHCODE_RUN,p.config.UniqueId,"",fmt.Sprintf("authcode适配器开始运行，其UniqueId为%s, Mode为:%s",
			   		  p.config.UniqueId, modeStr))

		p.config.News_Encode		= make(chan []byte)

	}else if p.config.Mode == DECODE{
		modeStr ="authcode为解密模式，将解密后的数据注入News_Decode管道"
		p.event_run = NewEvent(AUTHCODE_RUN,p.config.UniqueId,"",fmt.Sprintf("authcode适配器开始运行，其UniqueId为%s, Mode为:%s",
					  p.config.UniqueId, modeStr))


		p.config.News_Decode 		= make(chan []byte)
	}

	p.stop 				= make(chan struct{})

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
					if !ok{ return }
					p.encode(raw)
				case <-p.stop:
					return
				}
			}
		}()
	case DECODE:
		go func(){
			defer p.reactiveDestruct()
			for {
				select{
				case raw, ok := <-p.config.Raws:
					if !ok{ return }
					p.decode(raw)
				case <-p.stop:
					return
				}
			}
		}()
	}
}


func (p *AuthCode)ProactiveDestruct(){
	p.config.Events <-NewEvent(AUTHCODE_PROACTIVE_DESTRUCT,p.config.UniqueId,"","注意，由于某些原因authcode包主动调用了显式析构方法")
	p.stop<-struct{}{}	
}

//被动 - reactive
//被动析构是检测到Raws被上层关闭后的响应式析构操作
func (p *AuthCode)reactiveDestruct(){
	p.config.Events <-NewEvent(AUTHCODE_REACTIVE_DESTRUCT,p.config.UniqueId,"","authcode包触发了隐式析构方法")

	//析构数据源
	close(p.stop)

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
	Register(AUTHCODE_RIVERNODE_NAME, NewAuthCode)
	logger.Info("预加载完成，AuthCode适配器已预加载至package river_node.RNodes结构内")
}

/*------------以下是所需的功能方法-------------*/
func (p *AuthCode)encode(baits []byte){
	if res ,err := p.authcode.Encode(baits,""); err !=nil{
		p.countor++

		if p.countor > p.config.Limit_Encode{
			p.config.Errors <-NewError(AUTHCODE_ENCODE_FAIL, p.config.UniqueId, err.Error(),
							fmt.Sprintf("AUTHCODE加密连续%d次失败，已超过系统设定的最大次数，系统设定的最大连续失败"+
							"次数为%d",p.countor,p.config.Limit_Encode))
			p.countor =0
			p.config.Events <-p.event_fused
		}else{
			p.config.Errors <-NewError(AUTHCODE_ENCODE_FAIL, p.config.UniqueId, err.Error(),
				fmt.Sprintf("连续第%d次AUTHCODE加密失败，当前系统设定的最大连续失败次数为%d",
				   p.countor, p.config.Limit_Encode))
		}
		
	}else{
		if p.countor !=0{
			p.config.Events <-NewEvent(AUTHCODE_ENCODE_RECOVERED,p.config.UniqueId,"", fmt.Sprintf("已从第%d次AUTHCODE加密失败中恢复，"+
							"当前系统设定的最大失败次数为%d",p.countor,p.config.Limit_Encode))
			p.countor =0
		}
		p.config.News_Encode <- res
	}
}

func (p *AuthCode)decode(baits []byte){
	if res ,err := p.authcode.Decode(baits,""); err !=nil{
		p.countor++

		if p.countor > p.config.Limit_Decode{
			p.config.Errors <-NewError(AUTHCODE_DECODE_FAIL, p.config.UniqueId, err.Error(),
							fmt.Sprintf("AUTHCODE解密连续%d次失败，已超过系统设定的最大次数，系统设定的最大连续失败"+
							"次数为%d",p.countor,p.config.Limit_Decode))
			p.countor =0
			p.config.Events <-p.event_fused
		}else{
			p.config.Errors <-NewError(AUTHCODE_DECODE_FAIL, p.config.UniqueId, err.Error(),
				fmt.Sprintf("连续第%d次AUTHCODE解密失败，当前系统设定的最大连续失败次数为%d",
				   p.countor, p.config.Limit_Decode))
		}
		
	}else{
		if p.countor !=0{
			p.config.Events <-NewEvent(AUTHCODE_DECODE_RECOVERED,p.config.UniqueId,"", fmt.Sprintf("已从第%d次AUTHCODE解密失败中恢复，"+
							"当前系统设定的最大失败次数为%d",p.countor,p.config.Limit_Decode))
			p.countor =0
		}
		p.config.News_Decode <- res
	}
}
