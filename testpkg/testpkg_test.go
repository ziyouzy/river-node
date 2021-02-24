package testpkg

import (
    "river-node/heartbeating"
    "river-node/stamps"
    "river-node/crc"
    "river-node"
    "river-node/define"
    "river-node/logger"
    
 
    "fmt"
    "bytes"
    "time"
    "testing"	
)
 
/*此测试函数在实战中等同于一个zconn的初始化操作
 * 因此当zconn被销毁时这里的hbRaws，stampsRaws，stampsNews，需要预先被销毁
 * nodeHeartBeating,nodeStamps,nodeCRC也需要预先被销毁
 * 他们这两个我自定义的复杂数据类型地位是平级的
 */
 func TestNodes(t *testing.T) {
    defer logger.Flush()
    
    /*此管道的作用是测试信号的生成*/
    testBytesSenderCh := make(chan []byte)

    /*主线程是signal与error的统一管理管道*/
    signals           := make(chan Signal)
    errors            := make(chan error)

    /*心跳包的事件注入管道*/
    hbRaws            := make(chan struct{})

    /*CRC校验包的事件注入管道*/
    crcRaws           := make(chan []byte)
    /*CRC校验包会生成的新数据流*/
    crcPassNews       := make(chan []byte)
    crcNotPassNews    := make(chan []byte)

    /*印章包的事件注入管道*/
    stampsRaws        := make(chan []byte)
    /*印章包会生成的新数据流*/
    stampsNews        := make(chan []byte)
     
 
    heartBeatingAbsf    := river_node.Nodes[heartbeating.RIVER_NODE_NAME]
    stampsAbsf          := river_node.Nodes[stamps.RIVER_NODE_NAME]
    crcAbsf             := river_node.Nodes[crc.RIVER_NODE_NAME]
 
    nodeHeartBeating     := heartBeatingAbsf()
    nodeStampsAbsNode    := stampsAbsf()
    nodeCRC              := crcAbsf()
 
 
    /** 拿到对象类后，内部字段很多都是空的，所以需要进行初始化       
	 * 下面只是简单的演示一下
     */

/*----------在实战中，每一个zconn会分别包含一个heartbreat，
            一个crc，一个stamps，于是当zconn被销毁时，也就需要先销毁他们三个
            */
    heartBeatingConfig := &heartbeating.HeartBeatingConfig{
        UniqueId:   "testPkg",
         Timeout:   8 * time.Second,
         Signals:   signals,
          Errors:   errors,
            Raws:   hbRaws,
    }
 
    if err := nodeHeartBeating.Init(heartBeatingConfig); err == nil {
        logger.Info("test river-node init success")
    }
 
    nodeHeartBeating.Run()
//----------
    stampsConfig := &stamps.StampsConfig{
        UniqueId:   "testPkg",
	
        /** 分为三种，HEAD、TAIL、HEADANDTAIL
         * 当是HEADANDTAIL模式，切len(stamp)>1时
         * 按照如下顺序拼接：
         * stamp1+raw(首部);raw+stamp2(尾部);stamp3+raw(首部);raw+stamp4(尾部);stamp5+raw(首部);....
         * 这样的首尾交替规律的拼接方式
         */
            Mode:   stamps.HEADANDTAIL, 
        Breaking:   []byte("+"), /*戳与数据间的分隔符，可以为nil*/
          Stamps:   [][]byte{[]byte("city"),[]byte{0x01,0x00,0x01,0x00,},[]byte("name"),[]byte{0xff,},}, /*允许输入多个，会按顺序依次拼接*/
         Signals:   signals,/*发送给主进程的信号队列，就像Qt的信号与槽*/
          Errors:   errors,
            Raws:   stampRaws,/*从主线程发来的信号队列，就像Qt的信号与槽*/
            News:   stampNews,/*校验通过切去掉校验码的新切片*/
    } 

    if err := nodeStamps.Init(stampsConfig); err == nil {
        logger.Info("test river-node init success")
    }
 
    nodeStamps.Run()
//-----------
    crcConfig := &crc.CRCConfig{
        UniqueId:   "testPkg",

	    /** 有两种模式：define.READONLY和define.NEWCHAN
	    * 前者只判定当前字节序列是否能通过CRC校验
	    * 后者则校验后生成新的数据管道 
	    */

	        Mode:   crc.NEWCHAN,
	 IsBigEndian:   crc.ISBIGENDDIAN,
         Signals:   signals, /*发送给主进程的信号队列，就像Qt的信号与槽*/
          Errors:   errors,
	        Raws:   crcRaws, /*从主线程发来的信号队列，就像Qt的信号与槽*/
	    PassNews:   crcPassNews, /*校验通过切去掉校验码的新切片*/
     NotPassNews:   crcNotPassNews, /*校验未通过的原始校验码*/
    } 

    if err := NodeCRC.Init(crcConfig); err == nil {
        logger.Info("test river-node init success")
    }

    NodeCRC.Run()
//--------------------

 
	/*对signal与error进行统一回收和编排对应的触发事件*/
    go func(){
        /** 最后再统一思考关闭的操作吧
         * defer close(signals)
         * defer close(errors)
         */
        for {
            select{
            case sig := <-signals:
                /*最重要的是，触发某个事件后，接下来去做什么*/
                uniqueid, code, detail := sig.Description()
                switch code{
                case HEARTBREATING_RUN:
                    fmt.Println(uniqueid, "-detail:", detail)
                case HEARTBREATING_REBUILD:
                    fmt.Println(uniqueid, "-detail:", detail)
                case HEARTBREATING_NORMAL:
                    fmt.Println(uniqueid, "-detail:", detail)
                //case HEARTBREATING_TIMEOUT:
                    //timeout为Errors
                //case HEARTBREATING_RECOVERED:
                    //似乎暂时不需要处理所谓的“恢复”，到是也可以用日志记录一下
                case HEARTBREATING_PREPAREDESTORY:
                    fmt.Println(uniqueid, "-detail:", detail)
                    //detail中或许会包含的内容是“"心跳包连续多(5)次超时无响应，因此断开当前客户端连接"”
                    fmt.Println("可以从信号中获取被剔除ZConn的uid，从而基于uid进行后续的收尾工作")
                
                case CRC_RUN:
                    fmt.Println("CRC校验适配器被激活")
                case CRC_NORMAL:
                    fmt.Println("CRC成功校验某字节组")
                case CRC_UPSIDEDOWN:
                    fmt.Println("CRC校验检测出某字节组的校验码大小端反了")
                case CRC_REVERSEENDIAN:
                    fmt.Println("CRC校验适配器已自动将大小端进行了翻转")
                case CRC_DROPZCONN:
                    fmt.Println("可以从信号中获取被剔除ZConn的uid，从而基于uid进行后续的收尾工作")
                case STAMPS_RUN:
                    fmt.Println("STAMPS适配器被激活")

                case ANOTHEREXAMPLE_TEST1:
                    fmt.Println("ANOTHEREXAMPLE_TEST1测试成功")
                case ANOTHEREXAMPLE_TEST2:
                    fmt.Println("ANOTHEREXAMPLE_TEST2测试成功")
                case ANOTHEREXAMPLE_TEST3:
                    fmt.Println("ANOTHEREXAMPLE_TEST3测试成功")
                case ANOTHEREXAMPLE_ERR:
                    fmt.Println("ANOTHEREXAMPLE_ERR测试成功")
                default:
                    fmt.Println("未知的适配器返回了未知的信号类型这里不过多进行演示，"+
                                "详细的演示会在river-node/test包内进行")
                }			
                


            case err := <-errors:
                fmt.Println(err.Error())
                //实战中这里会进行日志的记录

            }
        }
    }()
 

    go func(){
        defer close(testBytesSenderCh)
        for i := 1;i < 20;i++{
            testBytesSenderCh <- []byte{0x01, 0x02, 0x03,}
            if i < 10{
                time.Sleep(time.Second)
            }else{
                time.Sleep(60*time.Second)
                i = 1
                fmt.Println("ReStart")
            }
        }
    }()
 

    go func(){
        defer close(hbRaws)
        defer close(stampsRaws)
        //defer close(mainTestCRCRawCh)
        for bytes := range testBytesSenderCh{
            fmt.Println("bytes is", bytes)
 
            hbRaws <- struct{}{}
            stampsRaws <- bytes	

            /** 事情要一件一件完成
            
             * 而且CRC校验包不会直接拦截一个“当前所在管道的”事件
             * (就好比这里的位置)
             * 而是借助自身管道，筛选出有效数据
             */
            //mainTestCRCRawCh <- bytes		
        }
    }()
 
 
    /** 如下线程用来实现各个适配器所生成新数据的数据调度工作
     * 要明确全局只会有1个signal回收管道，这个回收管道所在携程就是主携程，设计时要时刻考虑、遵循单向调用链模式
      
     * 之后的思路，每个适配器都会有独立的数据注入管道
     * 都是通过各自的config对象类实现“管道的对接”与“新事件的生成与注入”
 
     * 不过根据实际需求，有的适配器会拥有新数据管道，有的则没有
     
     * 而所有“新的数据管道”，根据业务需求，都会通过和上方携程同样的for range{}
     * 实现新的“数据处理节点”
 
     * 下方携程就是个大致的例子
     */

    go func(){
        defer close(mainTestCRCRawinCh)
        for bytesWithStamps := range mainTestStampsNewoutCh{
           fmt.Println("bytesWithStamps is", bytesWithStamps)
           mainTestCRCRawinCh <- bytesWithStamps
        }
    }()

    go func(){
        //defer 暂时似乎到了尽头 
        for bytesPassCRC := range mainTestCRCNewoutCh{
            fmt.Println("bytesPassCRC is：", bytes.Split(bytesPassCRC,[]byte("+")))
        }
    }()

    go func(){
        //defer 暂时似乎到了尽头 
        for bytesNotpassCRC := range mainTestCRCErroutCh{
           fmt.Println("bytesNotPassCRC is：", bytes.Split(bytesNotpassCRC,[]byte("+")))
        }
    }()
 
 
         /** 如何才算是一次完整的数据流“分流”事件呢？
          * 首先数据分流是数据处理的一种形式，或者说是分型
          * 数据处理只有两种形式，分流与截流
          * 从某个数据管道拿到事件后（for range拿到某一个切片）
          * 可能会做的一些事，包括但不限于：
          * 1.原始数据流截断
             （直接让数据消失，等同于 _ =xxx的效果
              他不属于数据分流，而是节流的一种
              “数据分流”是“数据处理”的一种分型）
              
          * 2.原始数据流转化为一个新数据流(新流只有一个也算分流)
          * 3.原始数据流转化为多个新数据流
             （转化的方式可以基于元数据流数据
              也可以是直接引入新的流）
 
          * 4.其他的数据截流处理方式：
            如fmt.Println()输出
            或最后一次不需创建新管道的数据转换+录入数据库
            等
 
          * 5.数据合流(聚合，无论是借助缓存实现的还是不借助缓存实现的)
          
          * 而在这里进行根本就不属于上面任何一种，因为上一次数据处理（分流）已经完成了
          * 这里的for range其实是已经开始了新一轮的数据处理（分流）
          
          * 其实可以概括成，每个for range{}都是一个独立的“数据处理节点”
          */

 
     //这里的测试并没有实现CRC适配器与Stamps包的任何对象，所以测试到此未知
 
     //复杂些的，会同时包含数据处理与传递的测试实例请看package river-node/test所包含的内容
     
    select{}
 }