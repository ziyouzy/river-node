package testpkg

import (
    "river-node/testpkg/testdatacreater"
    "river-node/heartbeating"
    "river-node/stamps"
    "river-node/crc"
    "river-node"
    "github.com/ziyouzy/logger"
    
 
    "fmt"
    //"bytes"
    "time"
    "testing"	
)
 

func TestNodes(t *testing.T) {
    defer logger.Destory()
//-----------
    Recriver()
//--------------------
    TestDataCreaterConnectCRC()
    CRCConnectStamps()
//-----------
    testDataCreaterAbsf   := river_node.RegisteredNodes[testdatacreater.RIVER_NODE_NAME]
    testDataCreaterNode   := testDataCreaterAbsf()
    testDataCreaterConfig := &testdatacreater.TestDataCreaterConfig{

        UniqueId:       "testPkg",
        Signals:        Signals,
        //Errors:         Errors,

        StepSec:		1 * time.Second,

 	    News: 		    TestDataCreaterNews,

    }


//-----------
   heartBeatingAbsf     := river_node.RegisteredNodes[heartbeating.RIVER_NODE_NAME]
   heartBeatingNode     := heartBeatingAbsf()
   heartBeatingConfig   := &heartbeating.HeartBeatingConfig{

       UniqueId:       "testPkg",
       Signals:        Signals,
       Errors:         Errors,

       TimeoutSec:     8 * time.Second,
       TimeoutLimit:   3,
       Raws:           HBRaws,
   }
 

//-----------
   crcAbsf              := river_node.RegisteredNodes[crc.RIVER_NODE_NAME]
   crcNode              := crcAbsf()
   crcConfig            := &crc.CRCConfig{

       UniqueId:      "testPkg",
       Signals:       Signals, /*发送给主进程的信号队列，就像Qt的信号与槽*/
       Errors:        Errors,

       Mode:          crc.NEWCHAN,
       IsBigEndian:   crc.ISBIGENDDIAN,
       NotPassLimit:  20,
       Raws:          CRCRaws, /*从主线程发来的信号队列，就像Qt的信号与槽*/
        
	   PassNews:      CRCPassNews, /*校验通过切去掉校验码的新切片*/
       NotPassNews:   CRCNotPassNews, /*校验未通过的原始校验码*/
   } 


//----------
   stampsAbsf           := river_node.RegisteredNodes[stamps.RIVER_NODE_NAME]
   stampsNode           := stampsAbsf()
   stampsNews            :=make(chan []byte)
   stampsConfig         := &stamps.StampsConfig{
       
       UniqueId:   "testPkg",
       Signals:    Signals,/*发送给主进程的信号队列，就像Qt的信号与槽*/
       Errors:     Errors,

       Mode:       stamps.HEADANDTAIL, 
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