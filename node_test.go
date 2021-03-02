package river_node

import (
    // "river-node/testpkg/testdatacreater"
    // "river-node/heartbeating"
    // "river-node/stamps"
    // "river-node/crc"
    // "river-node"
    "github.com/ziyouzy/logger"
    
 
    "fmt"
    //"bytes"
    "time"
    "testing"	
)
 

func TestNodes(t *testing.T) {
    defer logger.Destory()
//-----------
    Recriver(t)
//--------------------
    TestDataCreaterConnectCRC(t)
    CRCConnectStamps(t)
//-----------
    testDataCreaterAbsf   := RegisteredNodes[TEST_DATACREATER_RIVERNODE_NAME]
    testDataCreaterNode   := testDataCreaterAbsf()
    testDataCreaterConfig := &TestDataCreaterConfig{

        UniqueId:       "testPkg",
        Signals:        Signals,
        //Errors:         Errors,

        StepSec:		1 * time.Second,

 	    News: 		    TestDataCreaterNews,

    }


//-----------
   heartBeatingAbsf     := RegisteredNodes[HB_RIVERNODE_NAME]
   heartBeatingNode     := heartBeatingAbsf()
   heartBeatingConfig   := &HeartBeatingConfig{

       UniqueId:       "testPkg",
       Signals:        Signals,
       Errors:         Errors,

       TimeoutSec:     8 * time.Second,
       TimeoutLimit:   3,
       Raws:           HBRaws,
   }
 

//-----------
   crcAbsf              := RegisteredNodes[CRC_RIVERNODE_NAME]
   crcNode              := crcAbsf()
   crcConfig            := &CRCConfig{

       UniqueId:      "testPkg",
       Signals:       Signals, /*发送给主进程的信号队列，就像Qt的信号与槽*/
       Errors:        Errors,

       Mode:          NEWCHAN,
       IsBigEndian:   ISBIGENDDIAN,
       NotPassLimit:  20,
       Raws:          CRCRaws, /*从主线程发来的信号队列，就像Qt的信号与槽*/
        
	   PassNews:      CRCPassNews, /*校验通过切去掉校验码的新切片*/
       NotPassNews:   CRCNotPassNews, /*校验未通过的原始校验码*/
   } 


//----------
   stampsAbsf           := RegisteredNodes[STAMPS_RIVERNODE_NAME]
   stampsNode           := stampsAbsf()
   stampsNews            :=make(chan []byte)
   stampsConfig         := &StampsConfig{
       
       UniqueId:   "testPkg",
       Signals:    Signals,/*发送给主进程的信号队列，就像Qt的信号与槽*/
       Errors:     Errors,

       Mode:       HEADANDTAIL, 
       Breaking:   []byte("+"), /*戳与数据间的分隔符，可以为nil*/
       Stamps:     [][]byte{[]byte("city"),[]byte{0x01,0x00,0x01,0x00,},[]byte("name"),[]byte{0xff,},}, /*允许输入多个，会按顺序依次拼接*/          
       Raws:       StampsRaws,/*从主线程发来的信号队列，就像Qt的信号与槽*/
    
       News:       stampsNews,/*校验通过切去掉校验码的新切片*/
   } 

//--------------------

   if err := stampsNode.Init(stampsConfig); err != nil {
       logger.Info("test stamps-river-node init fail:"+err.Error())
   }else{
       stampsNode.Run()
   }
 
   if err := crcNode.Init(crcConfig); err != nil {
       logger.Info("test crc-river-node init fail:"+err.Error())
   }else{
       crcNode.Run()
   }

   if err := heartBeatingNode.Init(heartBeatingConfig); err != nil {
       logger.Info("test heartbeating-river-node init fail:"+err.Error())
   }else{
       heartBeatingNode.Run()
   }

   if err := testDataCreaterNode.Init(testDataCreaterConfig); err != nil {
       logger.Info("test testdatacreater-river-node init fail:"+err.Error())
   }else{
       //这里才是真正的数据注入操作，在这之前，无论如何去连接各个管道都不会担心“管道泄漏”问题
       testDataCreaterNode.Run()
   }

   //当前功能模块的末端输出流,地位等同於connector.go与逻辑上存在的creater.go内的函数或方法
   //是简单的管道连接与管道注入操作。与其对应的则是各个river-node，负责进行复杂的类似“事务”的符合操作
   go func(){
       for bytes := range stampsNews{
           fmt.Println("stampsNew：",string(bytes))
       }
   }()
      
   select{}
}


var (
    TestDataCreaterNews chan []byte
    HBRaws              chan struct{}
    CRCRaws             chan []byte
)
	
func TestDataCreaterConnectCRC(t *testing.T){
    TestDataCreaterNews = make(chan []byte)
    HBRaws              = make(chan struct{})
    CRCRaws             = make(chan []byte)
	go func(){
        for bytes := range TestDataCreaterNews{
            fmt.Println("source bytes is", bytes)
            HBRaws <- struct{}{}
            CRCRaws <- bytes	
        }
    }()
}

var (
    CRCPassNews    chan []byte     
    CRCNotPassNews chan []byte     
    StampsRaws     chan []byte    
)

func CRCConnectStamps(t *testing.T){
    CRCPassNews         = make(chan []byte)
    CRCNotPassNews      = make(chan []byte)
    StampsRaws          = make(chan []byte)
	go func(){
        for bytes := range CRCPassNews{
            StampsRaws <- bytes	
        }
	}()
	
	go func(){
        for bytes := range CRCNotPassNews{
            StampsRaws <- bytes	
        }
    }()
}



/*主线程是signal与error的统一管理管道*/
var (
    Signals chan Signal
    Errors  chan error
)

 
/*对signal与error进行统一回收和编排对应的触发事件*/
func Recriver(t *testing.T){
    Signals = make(chan Signal)
    Errors  = make(chan error)
    go func(){
   	    for {
            select{
            case sig := <-Signals:
                /*最重要的是，触发某个事件后，接下来去做什么*/
                fmt.Println("sig-",sig)
                uniqueid, code, detail, _:= sig.Description()
                fmt.Println("uniqueid, code, detail-",uniqueid, code, detail)
                switch code{
                case TESTDATACREATER_RUN:
                    fmt.Println(uniqueid, "-detail:", detail)
                case HEARTBREATING_RUN:
                    fmt.Println(uniqueid, "-detail:", detail)
                case HEARTBREATING_REBUILD:
                    fmt.Println(uniqueid, "-detail:", detail)
                case HEARTBREATING_NORMAL:
                    fmt.Println(uniqueid, "-detail:", detail)
                //case HEARTBREATING_TIMEOUT:
                    //timeout为Errors
                //case HEARTBREATING_TIMERLIMITED:
                //case HEARTBREATING_RECOVERED:
                    //似乎暂时不需要处理所谓的“恢复”，到是也可以用日志记录一下
                case HEARTBREATING_PANIC:
                    fmt.Println(uniqueid, "-detail:", detail)
                    //detail中或许会包含的内容是“"心跳包连续多(5)次超时无响应，因此断开当前客户端连接"”
                    fmt.Println("可以从信号中获取被剔除的uid，从而基于uid进行后续的收尾工作")
                
                case CRC_RUN:
                    fmt.Println(uniqueid, "-detail:", detail)
                case CRC_NORMAL:
                    fmt.Println("CRC成功校验某字节组")
                case CRC_UPSIDEDOWN:
                    fmt.Println("CRC校验检测出某字节组的校验码大小端反了")
                //case CRC_NOTPASS:
                //case CRC_RECOVERED:
                case CRC_PANIC:
                    fmt.Println("可以从信号中获取被剔除ZConn的uid，从而基于uid进行后续的收尾工作")
                
                case STAMPS_RUN:
                    fmt.Println("STAMPS适配器被激活")

                default:
                    fmt.Println("未知的适配器返回了未知的信号类型这里不过多进行演示，"+
                                "详细的演示会在river-node/test包内进行")
                }			
                

            case err := <-Errors:
                fmt.Println(err.Error())
                //实战中这里会进行日志的记录
            }
        }
    }()
}