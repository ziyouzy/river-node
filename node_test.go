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

    rawSimulatorAbsf   := RegisteredNodes[RAWSIMULATOR_RIVERNODE_NAME]
    rawSimulator       := rawSimulatorAbsf()
    rawSimulatorConfig := &RawSimulatorConfig{

        UniqueId:                   "testPkg",
        Events:                     Events,
        //Errors:                   Errors,

        StepSec:		            1 * time.Second,
    }

    if err := rawSimulator.Construct(rawSimulatorConfig); err != nil {
        logger.Info("test rawSimulator-river-node init fail:"+err.Error())
        panic("rawSimulator fail")
    }

//-----------

    hbRaws               = make(chan struct{})
    heartBeatingAbsf     := RegisteredNodes[HB_RIVERNODE_NAME]
    heartBeating         := heartBeatingAbsf()
    heartBeatingConfig   := &HeartBeatingConfig{

        UniqueId:                   "testPkg",
        Events:                     Events,
        Errors:                     Errors,

        TimeoutSec:                 3 * time.Second,
        TimeoutLimit:               3,
        Raws:                       hbRaws,

    }

    if err := heartBeating.Construct(heartBeatingConfig); err != nil {
        logger.Info("test heartbeating-river-node init fail:"+err.Error())
        panic("test heartbeating fail")
    } 

//-----------
    crcRaws              = make(chan []byte)
    crcAbsf              := RegisteredNodes[CRC_RIVERNODE_NAME]
    crc                  := crcAbsf()
    crcConfig            := &CRCConfig{

        UniqueId:                   "testPkg",
        Events:                     Events, /*发送给主进程的信号队列，就像Qt的信号与槽*/
        Errors:                     Errors,

        Mode:                       FILTER,
        Encoding:                   BIGENDDIAN,
        FilterNotPassLimit:         20,
        FilterStartIndex:           0,
        FilterMinLen:               4, 

        Raws:                       crcRaws, /*从主线程发来的信号队列，就像Qt的信号与槽*/               
    }

    if err := crc.Construct(crcConfig); err != nil {
        logger.Info("test crc-river-node init fail:"+err.Error())
        panic("test crc fail")
    }else{
        go func(){
            for bytes := range rawSimulatorConfig.News_ByteSlice{
                hbRaws <- struct{}{}
                crcRaws <- bytes
            }
        }()
    }

//----------

    stampsAbsf           := RegisteredNodes[STAMPS_RIVERNODE_NAME]
    stamps               := stampsAbsf()
    stampsConfig         := &StampsConfig{
       
        UniqueId:                   "testPkg",
        Events:                     Events,/*发送给主进程的信号队列，就像Qt的信号与槽*/
        Errors:                     Errors,

        Mode:                       HEADSANDTAILS, 
        Breaking:                   []byte("+"), /*戳与数据间的分隔符，可以为nil*/
        Stamps:                     [][]byte{[]byte("city"),[]byte{0x01,0x00,0x01,0x00,},[]byte("name"),[]byte{0xff,},}, /*允许输入多个，会按顺序依次拼接*/          
       
        Raws:                       crcConfig.News_Pass,/*从主线程发来的信号队列，就像Qt的信号与槽*/
    } 

    if err := stamps.Construct(stampsConfig); err != nil {
        logger.Info("test stamps-river-node init fail:"+err.Error())
        panic("test stamps fail")
    }else{
        go func(){
            for bytes := range stampsConfig.News{
                fmt.Println("fin_stampsNew：",string(bytes))
            }
        }()
    }

//--------------------




      
    stamps.Run()
    crc.Run()

    /**
     * 只要基于数据流动框架思路进行合理的设计，
     * 各个river-node的Init()与Run()执行的先后顺序并不会有什么先后要求
     * 还是那句话，需要执行析构的时候才会体现出会不会出现泄露的问题
     */
    rawSimulator.Run()

    heartBeating.Run()
    

    select{}
}

 
/*对event与error进行统一回收和编排对应的触发事件*/
func eventRecriver(t *testing.T){
    go func(){
   	    for {
            select{
            case eve := <-Events:
                /*最重要的是，触发某个事件后，接下来去做什么*/
                uniqueid, code, detail, _:= eve.Description()
                //fmt.Println("uniqueid, code, detail-",uniqueid, code, detail)
                switch code{
                case RAWSIMULATOR_RUN:
                    fmt.Println(uniqueid, "-detail:", detail)
                case RAWSIMULATOR_REACTIVEDESTRUCT:
                    fmt.Println(uniqueid, "-detail:", detail)
                case RAWSIMULATOR_PROACTIVEDESTRUCT:
                    fmt.Println(uniqueid, "-detail:", detail)

                case HEARTBREATING_RUN:
                    fmt.Println(uniqueid, "-detail:", detail)
                case HEARTBREATING_RECOVERED:
                    fmt.Println(uniqueid, "-detail:", detail)
                case HEARTBREATING_FUSED:
                    fmt.Println(uniqueid, "-detail:", detail)
                case HEARTBREATING_REACTIVEDESTRUCT:
                    fmt.Println(uniqueid, "-detail:", detail)
                case HEARTBREATING_PROACTIVEDESTRUCT:
                    fmt.Println(uniqueid, "-detail:", detail)

                case CRC_RUN:
                    fmt.Println(uniqueid, "-detail:", detail)
                case CRC_UPSIDEDOWN:
                    fmt.Println(uniqueid, "-detail:", detail)
                case CRC_RECOVERED:
                    fmt.Println(uniqueid, "-detail:", detail)
                case CRC_FUSED:
                    fmt.Println(uniqueid, "-detail:", detail)
                case CRC_REACTIVEDESTRUCT:
                    fmt.Println(uniqueid, "-detail:", detail)
                case CRC_PROACTIVEDESTRUCT:
                    fmt.Println(uniqueid, "-detail:", detail)
                
                case STAMPS_RUN:
                    fmt.Println(uniqueid, "-detail:", detail)
                case STAMPS_REACTIVEDESTRUCT:
                    fmt.Println(uniqueid, "-detail:", detail)
                case STAMPS_PROACTIVEDESTRUCT:
                    fmt.Println(uniqueid, "-detail:", detail)

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