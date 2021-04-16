package river_node

import (
	"fmt"
	"time"
	"testing"	
	"github.com/ziyouzy/logger"
	
	"github.com/ziyouzy/go-authcode"
)

var (
	Events  chan Event
	Errors  chan error
)


var (
	raws		chan []byte 
)

func TestAuthCode(t *testing.T) {
	defer logger.Destory()
	
	Events  = make(chan Event)
	Errors  = make(chan error)

	raws	= make(chan []byte)
		
	go func(){
		for {
			select{
		 	case eve := <-Events:
				/*最重要的是，触发某个事件后，接下来去做什么*/
				fmt.Println("Recriver-event:",eve.CodeString())
				c, cs, uid, data, commit, t:=eve.Description() 
				fmt.Println("Recriver-event-details:", c, cs, uid, data, commit, t)
				fmt.Println("Recriver-event-toError:",eve.ToError().Error()) 
		 	case err := <-Errors:
				fmt.Println("Recriver-error:",err.Error())
				//实战中这里会进行日志的记录
		 	}
	 	}
	}()
	//-----------
	
	authCodeAbsf   := RegisteredNodes[AUTHCODE_NODE_NAME]
	authCode       := authCodeAbsf()
	authCodeConfig := &AuthCodeConfig{
	
		UniqueId:                   "testPkg",
		Events:                     Events,
		Errors:                   	Errors,

		Mode: 			  			ENCODE,
	
		Raws: 		      			raws,

		AuthCode_Key:				"testkey",
		AuthCode_DynamicKeyLen:		4,
		AuthCode_ExpirySec:			2,

		Limit_Encode:				3,
		//----------
		//News_Decode					chan []byte
		//Limit_Decode      			int
	}
	
	if err := authCode.Construct(authCodeConfig); err != nil {
		logger.Info("test authCode-river-node init fail:"+err.Error())
		panic("authcode fail")
	}else{
		go func(){
			ac := go_authcode.New(authCodeConfig.AuthCode_Key,authCodeConfig.AuthCode_DynamicKeyLen,authCodeConfig.AuthCode_ExpirySec)
			for {
				bait := <-authCodeConfig.News_Encode
				fmt.Println("bait:",string(bait))
				if bait_new ,err := ac.Decode(nil,string(bait));err ==nil{
					fmt.Println(string(bait_new))
				}
			}
		}()
	}

	authCode.Run()

	go func(){
		
		for {
			raws <- []byte("jkdhfskhfkfh12131313--==///***++")
			time.Sleep(time.Second)
		}
	}()
    
   	select{}
}