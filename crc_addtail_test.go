package river_node

import (
    "github.com/ziyouzy/logger"
    
 
    "fmt"
    "encoding/hex"
    "time"
    "testing"	
)
 

/*主线程是event与error的统一管理管道*/
var (
    Events  chan Event
    Errors  chan error
)


var (
    rawSimulatorNews_ByteSlice    chan []byte

    crcNews_AddTail               chan []byte   
)

func TestInit(t *testing.T) {
    defer logger.Destory()

    Events  = make(chan Event)
    Errors  = make(chan error)
    
	go func(){
		for{
			select {
			case eve := <-Events:
				fmt.Printf("TestCRCTail event:")
				fmt.Println(eve.Description())
				 
			case err := <-Errors:
				fmt.Printf("TestCRCTail error:")
				fmt.Println(err.Error())
			}
		}
	}()

    rawSimulatorNews_ByteSlice   = make(chan []byte)
    rawSimulatorAbsf   := RegisteredNodes[RAWSIMULATOR_RIVERNODE_NAME]
    rawSimulator       := rawSimulatorAbsf()
    rawSimulatorConfig := &RawSimulatorConfig{

        UniqueId:                   "testPkg",
        Events:                     Events,
        //Errors:                   Errors,

        StepSec:		            1 * time.Second,

 	    News_ByteSlice: 		    rawSimulatorNews_ByteSlice,

    } 

//-----------
	crcNews_AddTail      = make(chan []byte)

    crcAbsf              := RegisteredNodes[CRC_RIVERNODE_NAME]
    crc                  := crcAbsf()
    crcConfig            := &CRCConfig{

        UniqueId:                   "testPkg",
        Events:                     Events, /*发送给主进程的信号队列，就像Qt的信号与槽*/
        Errors:                     Errors,

        Mode:                       ADDTAIL,
        Encoding:                   BIGENDDIAN,
        //FilterNotPassLimit:         20,
        //FilterStartIndex:           4,
        Raws:                       rawSimulatorNews_ByteSlice, /*从主线程发来的信号队列，就像Qt的信号与槽*/
        
	    News_AddTail:            	crcNews_AddTail,
    } 

//-----------

	if err := rawSimulator.Construct(rawSimulatorConfig); err != nil {
		logger.Info("test rawSimulator-river-node init fail:"+err.Error())
		panic("rawSimulator fail")
	}

	if err := crc.Construct(crcConfig); err != nil {
		logger.Info("test crc-river-node init fail:"+err.Error())
		panic("test crc fail")
	}

//----------

    go func(){
        for byteslice := range crcNews_AddTail{
            fmt.Println("with tail：",hex.EncodeToString(byteslice))
        }
    }()
      
    rawSimulator.Run()
    crc.Run()

    select{}
}

