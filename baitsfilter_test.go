package river_node

import (
	"fmt"
	"time"
	"testing"	
	"github.com/ziyouzy/logger"
)

var (
	Events  chan Event
	Errors  chan error
)


var (
	raws		chan []byte 
)

func TestBaitsFilter(t *testing.T) {
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
	
	baitsFilterAbsf   := RegisteredNodes[BAITSFILTER_NODE_NAME]
	baitsFilter       := baitsFilterAbsf()
	baitsFilterConfig := &BaitsFilterConfig{
	
		UniqueId:                   "testPkg",
		Events:                     Events,
		Errors:                   	Errors,

		Mode: 			  			DROPHEAD,
		Heads:						[][]byte{[]byte("test1"),[]byte("test2"),[]byte("test3"),},
		Len_max:					100,
		Len_min:					0,
	
		Raws: 		      			raws,
	}
	
	if err := baitsFilter.Construct(baitsFilterConfig); err != nil {
		logger.Info("test baitsFilter-river-node init fail:"+err.Error())
		panic("baitsFilter fail")
	}else{
		go func(){
			for {
				baits := <-baitsFilterConfig.News_DropHead
				fmt.Println(fmt.Sprintf("bait:%x,string:%s",baits,string(baits)))
			}
		}()
	}

	baitsFilter.Run()

	go func(){
		for {
			raws <- []byte("test01abcdefghic")
			time.Sleep(time.Second)
			raws <- []byte("test11abcdefghic")
			time.Sleep(time.Second)
			raws <- []byte("test21abcdefghic")
			time.Sleep(time.Second)
			raws <- []byte("test31abcdefghic")
			time.Sleep(time.Second)
			raws <- []byte("test41abcdefghic")
			time.Sleep(time.Second)
			raws <- []byte("test51abcdefghic")
			time.Sleep(time.Second)
		}
	}()
    
   	select{}
}