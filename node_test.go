package river_node

import (
    // "river-node/testpkg/testdatacreater"
    // "river-node/heartbeating"
    // "river-node/stamps"
    // "river-node/crc"
    // "river-node"
    "github.com/ziyouzy/logger"
    
 
    "fmt"
    "bytes"
    "time"
    "testing"	
    //"encoding/hex"
)
 

/*主线程是event与error的统一管理管道*/
var (
    Events  chan Event
    Errors  chan error
)


var (
    hbRaws                     chan struct{}
    crcRaws                    chan []byte
)

func TestInit(t *testing.T) {
    defer logger.Destory()

    Events  = make(chan Event)
    Errors  = make(chan error)
    eventRecriver(t)
    
//-----------

    hbRaws               = make(chan struct{})
    heartBeatingAbsf     := RegisteredNodes[HB_NODE_NAME]
    heartBeating         := heartBeatingAbsf()
    heartBeatingConfig   := &HeartBeatingConfig{

        UniqueId:                   "testPkg",
        Events:                     Events,
        Errors:                     Errors,

        TimeoutSec:                 3 * time.Second,
        Limit:                      3,
        Raws:                       hbRaws,

    }

    if err := heartBeating.Construct(heartBeatingConfig); err != nil {
        logger.Info("test heartbeating-river-node init fail:"+err.Error())
        panic("test heartbeating fail")
    } 

//-----------
    crcRaws              = make(chan []byte)
    crcAbsf              := RegisteredNodes[CRC_NODE_NAME]
    crc                  := crcAbsf()
    crcConfig            := &CRCConfig{

        UniqueId:                   "testPkg",
        Events:                     Events, /*发送给主进程的信号队列，就像Qt的信号与槽*/
        Errors:                     Errors,

        Mode:                       FILTER,
        Encoding:                   LITTLEENDDIAN,
        Limit_Filter:               20,
        StartIndex_Filter:          0,
        MinLen_Filter:              4, 

        Raws:                       crcRaws, /*从主线程发来的信号队列，就像Qt的信号与槽*/               
    }

    if err := crc.Construct(crcConfig); err != nil {
        logger.Info("test crc-river-node init fail:"+err.Error())
        panic("test crc fail")
    }

//----------

    stampsAbsf           := RegisteredNodes[STAMPS_NODE_NAME]
    stamps               := stampsAbsf()
    stampsConfig         := &StampsConfig{
       
        UniqueId:                   "testPkg",
        Events:                     Events,/*发送给主进程的信号队列，就像Qt的信号与槽*/
        Errors:                     Errors,

        AutoTimeStamp:              true,
        Mode:                       HEADS, 
        Breaking:                   []byte("/-/"), /*戳与数据间的分隔符，可以为nil*/
        Stamps:                     [][]byte{[]byte("city"),[]byte{0x01,0x00,0x01,0x00,},[]byte("name"),[]byte{0xff,},}, /*允许输入多个，会按顺序依次拼接*/          
       
        Raws:                       crcConfig.News_Filter,/*从主线程发来的信号队列，就像Qt的信号与槽*/
    } 

    if err := stamps.Construct(stampsConfig); err != nil {
        logger.Info("test stamps-river-node init fail:"+err.Error())
        panic("test stamps fail")
    }

//--------------------

    /**
     * 只要基于数据流动框架思路进行合理的设计，
     * 各个river-node的Init()与Run()执行的先后顺序并不会有什么先后要求
     */
    stamps.Run()
    crc.Run()
    heartBeating.Run()


	//  sourceTable = [][]byte{
	// 	[]byte{0x01, 0x02, 0x03, 0x04,},
	// 	[]byte{0x05, 0x06, 0x07, 0x08,},
	// 	[]byte{0x01, 0x04, 0x09, 0x13,},
	// 	[]byte{0xF1,0x05,0x00,0x00,0xFF,0x00,},
	// 	[]byte{0xF1,0x05,0x00,0x00,0x00,0x00,},}

	sourceTable := [][]byte{{0xF1,0x01,0x00,0x00,0x00,0x08, 0x29,0x3C},
                   {0xF1,0x02,0x00,0x20,0x00,0x08,0x6C,0xF6},}

    go func(){
        for i:=0;i<=len(sourceTable);i++{
            if i == len(sourceTable){
                i = 0
            }

            hbRaws <- struct{}{}
            crcRaws <- sourceTable[i]
            time.Sleep(time.Second)
        }
    }()

    go func(){
        for byteslice := range stampsConfig.News_Heads{
            //fmt.Println(byteslice)
            bl :=bytes.Split(byteslice,[]byte("/-/"))
            fmt.Println(bl)
            fmt.Println("fin of stamps news is:", StringTimeStamp(bl[0],true),string(bl[1]),fmt.Sprintf("%x",bl[2]),string(bl[3]),
            fmt.Sprintf("%x",bl[4]),fmt.Sprintf("%x",bl[5]))
        }
    }()
    
    select{}
}

 
/*对event与error进行统一回收和编排对应的触发事件*/
func eventRecriver(t *testing.T){
    go func(){
   	    for {
            select{
            case eve := <-Events:
                /*最重要的是，触发某个事件后，接下来去做什么*/
                fmt.Println("Recriver-event:",eve.CodeString())
                c, cs, uid, data, commit,time:=eve.Description() 
                fmt.Println("Recriver-event-details:", c, cs, uid, data, commit,time)
                fmt.Println("Recriver-event-toError:",eve.ToError().Error()) 
            case err := <-Errors:
                fmt.Println("Recriver-error:",err.Error())
                //实战中这里会进行日志的记录
            }
        }
    }()
}