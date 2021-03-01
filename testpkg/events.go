package testpkg

import (
    "river-node"

    "fmt"
)



/*主线程是signal与error的统一管理管道*/
var (
    Signals chan river_node.Signal
    Errors  chan error
)

 
/*对signal与error进行统一回收和编排对应的触发事件*/
func Recriver(){
    Signals = make(chan river_node.Signal)
    Errors  = make(chan error)
    go func(){
   	    for {
            select{
            case sig := <-Signals:
                /*最重要的是，触发某个事件后，接下来去做什么*/
                uniqueid, code, detail, _:= sig.Description()
                switch code{
                case river_node.TESTDATACREATER_RUN:
                    fmt.Println(uniqueid, "-detail:", detail)
                case river_node.HEARTBREATING_RUN:
                    fmt.Println(uniqueid, "-detail:", detail)
                case river_node.HEARTBREATING_REBUILD:
                    fmt.Println(uniqueid, "-detail:", detail)
                case river_node.HEARTBREATING_NORMAL:
                    fmt.Println(uniqueid, "-detail:", detail)
                //case river_node.HEARTBREATING_TIMEOUT:
                    //timeout为Errors
                //case river_node.HEARTBREATING_TIMERLIMITED:
                //case river_node.HEARTBREATING_RECOVERED:
                    //似乎暂时不需要处理所谓的“恢复”，到是也可以用日志记录一下
                case river_node.HEARTBREATING_PANIC:
                    fmt.Println(uniqueid, "-detail:", detail)
                    //detail中或许会包含的内容是“"心跳包连续多(5)次超时无响应，因此断开当前客户端连接"”
                    fmt.Println("可以从信号中获取被剔除的uid，从而基于uid进行后续的收尾工作")
                
                case river_node.CRC_RUN:
                    fmt.Println(uniqueid, "-detail:", detail)
                case river_node.CRC_NORMAL:
                    fmt.Println("CRC成功校验某字节组")
                case river_node.CRC_UPSIDEDOWN:
                    fmt.Println("CRC校验检测出某字节组的校验码大小端反了")
                //case river_node.CRC_NOTPASS:
                //case river_node.CRC_RECOVERED:
                case river_node.CRC_PANIC:
                    fmt.Println("可以从信号中获取被剔除ZConn的uid，从而基于uid进行后续的收尾工作")
                
                case river_node.STAMPS_RUN:
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