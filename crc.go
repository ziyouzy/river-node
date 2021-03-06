package river_node

import (
	"github.com/ziyouzy/logger"

	"fmt"
	"bytes"
	"reflect"
	"errors"
	"encoding/binary"
)


const CRC_RIVERNODE_NAME = "crc"


type CRCConfig struct{
	UniqueId 		  			string	/*其所属上层数据通道(如Conn)的唯一识别标识*/
	Events 		  				chan Event /*发送给主进程的信号队列，就像Qt的信号与槽*/
	Errors 			  			chan error

	Mode 			  			int /*define.READONLY或define.NEWCHAN*/
	IsBigEndian 	  			bool
	NotPassLimit      			int
	
	Raws 		      			chan []byte /*从主线程发来的信号队列，就像Qt的信号与槽*/

	News_Pass 		  			chan []byte /*校验通过切去掉校验码的新切片*/
	News_NotPass 	  			chan []byte /*校验未通过的原始校验码*/
}




func (p *CRCConfig)Name()string{
	return CRC_RIVERNODE_NAME
}



type CRC struct{
	mbTable 			[]uint16
   	bytesHandler 		*bytes.Buffer
	config 				*CRCConfig

	countor 			int
	event_run 			Event
	event_upsidedown 	Event
	event_fused 		Event

	stop				chan struct{}
}

func (p *CRC)Name()string{
	return CRC_RIVERNODE_NAME
}

func (p *CRC)Construct(CRCConfigAbs Config) error{
	if CRCConfigAbs.Name() != CRC_RIVERNODE_NAME {
		return errors.New("crc river_node init error, config must CRCConfig")
	}


	v := reflect.ValueOf(CRCConfigAbs)
	c := v.Interface().(*CRCConfig)


	if c.UniqueId == "" {
		return errors.New("crc river_node init error, uniqueId is nil")
	}

	if c.Events == nil || c.Errors == nil{
		return errors.New("crc river_node init error, Events or Errors is nil")
	}

	if c.Raws == nil || c.NotPassLimit ==0{
		return errors.New("crc river_node init error, Raws or NotPassLimit is nil")
	}

	if c.Mode != NEWCHAN&&c.Mode != READONLY {
		return errors.New("crc river-node init error, unknown mode") 
	}


	if c.Mode ==NEWCHAN && (c.News_Pass == nil || c.News_NotPass == nil){
		return errors.New("newchan mode crc river-node init error, "+
		             "PassNews or NotPassNews or  is nil") 
	}

	if c.Mode ==READONLY && (c.News_Pass != nil || c.News_NotPass != nil){
		return errors.New("readonly mode crc river-node init error, "+
		             "PassNews or NotPassNews or  is not nil") 
	}

	
	p.config = c

	p.mbTable = []uint16{
		0X0000, 0XC0C1, 0XC181, 0X0140, 0XC301, 0X03C0, 0X0280, 0XC241,
		0XC601, 0X06C0, 0X0780, 0XC741, 0X0500, 0XC5C1, 0XC481, 0X0440,
		0XCC01, 0X0CC0, 0X0D80, 0XCD41, 0X0F00, 0XCFC1, 0XCE81, 0X0E40,
		0X0A00, 0XCAC1, 0XCB81, 0X0B40, 0XC901, 0X09C0, 0X0880, 0XC841,
		0XD801, 0X18C0, 0X1980, 0XD941, 0X1B00, 0XDBC1, 0XDA81, 0X1A40,
		0X1E00, 0XDEC1, 0XDF81, 0X1F40, 0XDD01, 0X1DC0, 0X1C80, 0XDC41,
		0X1400, 0XD4C1, 0XD581, 0X1540, 0XD701, 0X17C0, 0X1680, 0XD641,
		0XD201, 0X12C0, 0X1380, 0XD341, 0X1100, 0XD1C1, 0XD081, 0X1040,
		0XF001, 0X30C0, 0X3180, 0XF141, 0X3300, 0XF3C1, 0XF281, 0X3240,
		0X3600, 0XF6C1, 0XF781, 0X3740, 0XF501, 0X35C0, 0X3480, 0XF441,
		0X3C00, 0XFCC1, 0XFD81, 0X3D40, 0XFF01, 0X3FC0, 0X3E80, 0XFE41,
		0XFA01, 0X3AC0, 0X3B80, 0XFB41, 0X3900, 0XF9C1, 0XF881, 0X3840,
		0X2800, 0XE8C1, 0XE981, 0X2940, 0XEB01, 0X2BC0, 0X2A80, 0XEA41,
		0XEE01, 0X2EC0, 0X2F80, 0XEF41, 0X2D00, 0XEDC1, 0XEC81, 0X2C40,
		0XE401, 0X24C0, 0X2580, 0XE541, 0X2700, 0XE7C1, 0XE681, 0X2640,
		0X2200, 0XE2C1, 0XE381, 0X2340, 0XE101, 0X21C0, 0X2080, 0XE041,
		0XA001, 0X60C0, 0X6180, 0XA141, 0X6300, 0XA3C1, 0XA281, 0X6240,
		0X6600, 0XA6C1, 0XA781, 0X6740, 0XA501, 0X65C0, 0X6480, 0XA441,
		0X6C00, 0XACC1, 0XAD81, 0X6D40, 0XAF01, 0X6FC0, 0X6E80, 0XAE41,
		0XAA01, 0X6AC0, 0X6B80, 0XAB41, 0X6900, 0XA9C1, 0XA881, 0X6840,
		0X7800, 0XB8C1, 0XB981, 0X7940, 0XBB01, 0X7BC0, 0X7A80, 0XBA41,
		0XBE01, 0X7EC0, 0X7F80, 0XBF41, 0X7D00, 0XBDC1, 0XBC81, 0X7C40,
		0XB401, 0X74C0, 0X7580, 0XB541, 0X7700, 0XB7C1, 0XB681, 0X7640,
		0X7200, 0XB2C1, 0XB381, 0X7340, 0XB101, 0X71C0, 0X7080, 0XB041,
		0X5000, 0X90C1, 0X9181, 0X5140, 0X9301, 0X53C0, 0X5280, 0X9241,
		0X9601, 0X56C0, 0X5780, 0X9741, 0X5500, 0X95C1, 0X9481, 0X5440,
		0X9C01, 0X5CC0, 0X5D80, 0X9D41, 0X5F00, 0X9FC1, 0X9E81, 0X5E40,
		0X5A00, 0X9AC1, 0X9B81, 0X5B40, 0X9901, 0X59C0, 0X5880, 0X9841,
		0X8801, 0X48C0, 0X4980, 0X8941, 0X4B00, 0X8BC1, 0X8A81, 0X4A40,
		0X4E00, 0X8EC1, 0X8F81, 0X4F40, 0X8D01, 0X4DC0, 0X4C80, 0X8C41,
		0X4400, 0X84C1, 0X8581, 0X4540, 0X8701, 0X47C0, 0X4680, 0X8641,
		0X8201, 0X42C0, 0X4380, 0X8341, 0X4100, 0X81C1, 0X8081, 0X4040}

	p.bytesHandler = bytes.NewBuffer([]byte{})

	p.event_upsidedown = NewEvent(CRC_UPSIDEDOWN, c.UniqueId, "")
	p.event_fused 	  	= NewEvent(CRC_FUSED, c.UniqueId, "")


	endianStr := ""
	modeStr   := ""

	if p.config.IsBigEndian{
		endianStr ="大端模式"
	}else{
		endianStr ="小端模式"
	}

	if p.config.Mode == READONLY{
		modeStr ="只判断是否校验通过，只将结果通过Events管道返回给上层(READONLY)"
	}else if p.config.Mode == NEWCHAN{
		modeStr ="不仅仅判断是否校验通过，并为通过与未通过校验的数据分别创建新的管道(NEWCHAN)"
	}
	
	p.event_run = NewEvent(CRC_RUN,p.config.UniqueId,
		fmt.Sprintf("CRC校验适配器开始运行，其UniqueId为%s, 最大校验失败次数为%d, Mode为:%s,"+
		   "大小端模式为:%s",p.config.UniqueId, p.config.NotPassLimit, modeStr, endianStr))


	p.stop =make(chan struct{})

	return nil
}


func (p *CRC)Run(){

	p.config.Events <- p.event_run


	/** 
	 * 如下为Run()的核心，也就是注入数据的操作环节
	 * 或者说是对数据源的使用环节
	 * 当Raws被关闭后，至少系统中的数据流动框架会立刻回归到Construct()时的状态
	 */
	switch p.config.Mode{
	case READONLY:
		go func(){
			defer p.reactiveDestruct()
			for {
				select{
				case mb, ok := <-p.config.Raws:
					if !ok{ return }
					if !p.readOnlyCheck(mb){ return }
				case <-p.stop:
					return
				}
			}
		}()
	case NEWCHAN:
		go func(){
			defer p.reactiveDestruct()
			for {
				select{
				case mb, ok := <-p.config.Raws:
					if !ok{ return }
					if !p.newChanCheck(mb){ return }
				case <-p.stop:
					return
				}
			}
		}()
	}
}

/**
 * 基于：
 * “无论怎样都不应该在消费端关闭channel,因为在消费端无法判断生产端是否还会向通道中发送元素值”
 * 的设计思路，ProactiveDestruct()的存在是不合理的
 */
//主动 - proactive
//主动析构并不会主动触发Raws管道的关闭，具体请参考heartbeating适配器
//但是也一定要慎重使用
func (p *CRC)ProactiveDestruct(){
	p.config.Events <-NewEvent(CRC_PROACTIVEDESTRUCT,p.config.UniqueId,
	 "注意，由于某些原因CRC校验包主动调用了显式析构方法")
	/**
	 * 析构数据源,析构后Run()内部会自动触发下面的reactiveDestruct()方法
	 */
	p.stop<-struct{}{}
	
}

//被动 - reactive
//被动析构是检测到Raws被上层关闭后的响应式析构操作
func (p *CRC)reactiveDestruct(){
	p.config.Events <-NewEvent(CRC_REACTIVEDESTRUCT,p.config.UniqueId,"CRC校验包触发了隐式析构方法")

	//析构数据源
	close(p.stop)
	switch p.config.Mode{
	case READONLY:
	case NEWCHAN:
		close(p.config.News_NotPass)
		close(p.config.News_Pass)
	}
}



func NewCRC() NodeAbstract {
	return &CRC{}
}


func init() {
	Register(CRC_RIVERNODE_NAME, NewCRC)
	logger.Info("预加载完成，CRC校验适配器已预加载至package river_node.RNodes结构内")
}





/*------------以下是所需的功能方法-------------*/



//验证一个需要验证的modbus码
func (p *CRC)readOnlyCheck(mb []byte) bool { 

	raw,crc := p.midModbus(mb) 

	if bytes.Equal(p.checkCRC16(raw, p.config.IsBigEndian), crc){
		if p.countor != 0{ 
			p.countor = 0
			p.config.Events <- NewEvent(CRC_RECOVERED,p.config.UniqueId,
					fmt.Sprintf("已从第%d次CRC校验失败中恢复，当前系统设定的最大失败次数为%d",
					   p.countor,p.config.NotPassLimit))
		}

		return true
	}else if bytes.Equal(p.checkCRC16(raw, !p.config.IsBigEndian),crc){
		if p.countor != 0{
			p.countor = 0
			p.config.Events <-NewEvent(CRC_RECOVERED,p.config.UniqueId,
					fmt.Sprintf("已从第%d次CRC校验失败中恢复，当前系统设定的最大失败次数为%d,"+
					   "但是当前这一字节数组存在大小端颠倒的问题",p.countor,p.config.NotPassLimit))
		}

		//无论countor是否为0都会返回upsidedown信号
		p.config.Events <- p.event_upsidedown
		return true

	}else if p.countor < p.config.NotPassLimit{
		p.countor++
		p.config.Errors <-NewError(CRC_NOTPASS,p.config.UniqueId, 
				fmt.Sprintf("连续第%d次CRC校验失败，当前系统设定的最大连续失败次数为%d",
				   p.countor,p.config.NotPassLimit))
		return true
	}else{
		p.config.Errors <- NewError(CRC_FUSED,p.config.UniqueId, 
				fmt.Sprintf("CRC验证连续%d次失败，已超过系统设定的最大次数，系统设定的最大连续失败"+
				   "次数为%d",p.countor,p.config.NotPassLimit))
		p.countor =0
		p.config.Events <- p.event_fused
		return false
	}
}

//验证并生成一个全新的切片变量并放入newchan管道中，这个全新的切片的底层数组也是全新的
func (p *CRC)newChanCheck(mb []byte) bool{
	raw,crc := p.midModbus(mb) 

	if bytes.Equal(p.checkCRC16(raw, p.config.IsBigEndian), crc){
		if p.countor != 0{
			p.countor = 0
			p.config.Events <- NewEvent(CRC_RECOVERED,p.config.UniqueId,
					fmt.Sprintf("已从第%d次CRC校验失败中恢复，当前系统设定的最大失败次数为%d",
					   p.countor,p.config.NotPassLimit))
		}

		p.bytesHandler.Reset()
		p.bytesHandler.Write(raw)//通过crc校验则阉割掉后两位校验位
		p.config.News_Pass <-p.bytesHandler.Bytes()

		return true

	}else if bytes.Equal(p.checkCRC16(raw, !p.config.IsBigEndian),crc){
		if p.countor != 0{
			p.countor = 0
			p.config.Events <-NewEvent(CRC_RECOVERED,p.config.UniqueId,
					fmt.Sprintf("已从第%d次CRC校验失败中恢复，当前系统设定的最大失败次数为%d,"+
					   "但是当前这一字节数组存在大小端颠倒的问题",p.countor,p.config.NotPassLimit))
		}

		p.config.Events <- p.event_upsidedown

		p.bytesHandler.Reset()
		p.bytesHandler.Write(raw)//通过crc校验则阉割掉后两位校验位
		p.config.News_NotPass <-p.bytesHandler.Bytes()

		return true

	}else if p.countor < p.config.NotPassLimit{
		p.countor++
		p.config.Errors <-NewError(CRC_NOTPASS,p.config.UniqueId, 
				fmt.Sprintf("连续第%d次CRC校验失败，当前系统设定的最大连续失败次数为%d",p.countor,
				   p.config.NotPassLimit))

		p.bytesHandler.Reset()
		p.bytesHandler.Write(mb)//未通过crc校验则保留后两位校验位
		p.config.News_NotPass <-p.bytesHandler.Bytes()

		return true
	}else{
		p.config.Errors <-NewError(CRC_FUSED,p.config.UniqueId, 
				fmt.Sprintf("CRC验证连续%d次失败，已超过系统设定的最大次数，系统设定的最大连续失败"+
				   "次数为%d",p.countor,p.config.NotPassLimit))
		p.countor =0

		p.bytesHandler.Reset()
		p.bytesHandler.Write(mb)//未通过crc校验则保留后两位校验位
		p.config.News_NotPass <- p.bytesHandler.Bytes()

		p.config.Events <- p.event_fused
		return false
	}
}

//拆分需要验证的modbus码，拆分后会共享原始切片(mb)的底层数组
//不可以对两个返回值做任何直接的修改操作
func (p *CRC)midModbus(mb []byte) ([]byte, []byte){
	l :=len(mb)
	return mb[:l-2], mb[l-2:]
}


//计算出某个字节数组的crc16校验码
func (p *CRC)checkCRC16(data []byte, isBigEndian bool) []byte {
	var crc16 uint16 = 0xffff

	for _, v := range data {
		n := uint8(uint16(v)^crc16)
		crc16 >>= 8
		crc16 ^= p.mbTable[n]
	}

	
	if isBigEndian{
		p.bytesHandler.Reset()
		binary.Write(p.bytesHandler, binary.BigEndian, &crc16)
	}else{
		p.bytesHandler.Reset()
		binary.Write(p.bytesHandler, binary.LittleEndian, &crc16)
	}
	return p.bytesHandler.Bytes()
}