package river_node

import (
	"github.com/ziyouzy/logger"

	"fmt"
	"bytes"
	"reflect"
	"errors"
	"encoding/binary"
	"encoding/hex"
)


const CRC_NODE_NAME = "crc校验适配器"


type CRCConfig struct{
	UniqueId 		  			string	/*其所属上层数据通道(如Conn)的唯一识别标识*/
	Events 		  				chan Event /*发送给主进程的信号队列，就像Qt的信号与槽*/
	Errors 			  			chan error

	Mode 			  			int /*define.FILTER或ADDTAIL*/
	Encoding 	  			    bool
	
	Raws 		      			chan []byte /*从主线程发来的信号队列，就像Qt的信号与槽*/

	News_AddTail				chan []byte
	//----------
	StartIndex_Filter			int
	Limit_Filter      			int
	MinLen_Filter               int
	News_Filter 		  		chan []byte /*校验通过切去掉校验码的新切片*/

}




func (p *CRCConfig)Name()string{
	return CRC_NODE_NAME
}



type CRC struct{
	mbTable 			[]uint16
   	bytesHandler 		*bytes.Buffer
	config 				*CRCConfig

	countor 			int
	event_run 			Event
	event_fused 		Event
}

func (p *CRC)Name()string{
	return CRC_NODE_NAME
}

func (p *CRC)Construct(CRCConfigAbs Config) error{
	if CRCConfigAbs.Name() != CRC_NODE_NAME {
		return errors.New(fmt.Sprintf("[river-node type:%s] init error, config must CRCConfig", p.Name()))
	}


	v := reflect.ValueOf(CRCConfigAbs)
	c := v.Interface().(*CRCConfig)


	if c.UniqueId == "" {
		return errors.New(fmt.Sprintf("[river-node type:%s] init error, uniqueId is nil", p.Name()))
	}

	if c.Events == nil || c.Errors == nil || c.Raws == nil{
		return errors.New(fmt.Sprintf("[uid:%s] init error, Events or Errors or Raws is nil", p.config.UniqueId))
	}

	if c.News_Filter != nil || c.News_AddTail != nil{
		return errors.New(fmt.Sprintf("[uid:%s] init error, News_Filter or News_AddTail is not nil", p.config.UniqueId))
	}


	if c.Mode != FILTER&&c.Mode != ADDTAIL {
		return errors.New(fmt.Sprintf("[uid:%s] init error, unknown mode", p.config.UniqueId)) 
	}

	
	if c.Mode ==FILTER && (c.Limit_Filter ==0 || c.MinLen_Filter ==0){
		return errors.New(fmt.Sprintf("[filter-uid:%s] init error, Limit_Filter or MinLen_Filter is nil", p.config.UniqueId)) 
	}

	/*FilterStartIndex允许为0且默认为0*/
	
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

	p.bytesHandler 		= bytes.NewBuffer([]byte{})

	p.event_fused 	  	= NewEvent(CRC_FUSED, c.UniqueId, "", "")


	endianStr := ""
	modeStr   := ""

	if p.config.Encoding ==BIGENDDIAN{
		endianStr ="大端模式"
	}else{
		endianStr ="小端模式"
	}
	
	if p.config.Mode == FILTER{
		modeStr ="crc校验模式，将通过的结果注入News_Filter管道，不通过的进行转化并注入Errors管道"
		p.event_run = NewEvent(CRC_RUN,p.config.UniqueId,"",
			fmt.Sprintf("[uid:%s;mode:%s]开始运行，最大校验失败次数为%d,大小端模式为:%s,校验起始下标为:%d",
			p.config.UniqueId, modeStr, p.config.Limit_Filter, endianStr,p.config.StartIndex_Filter))

		p.config.News_Filter = make(chan []byte)

	}else if p.config.Mode == ADDTAIL{
		modeStr ="追加crc校验码模式，将追加后生成的modbus码注入News_AddTail管道"
		p.event_run = NewEvent(CRC_RUN,p.config.UniqueId,"",
			fmt.Sprintf("[uid:%s;mode:%s]开始运行，大小端模式为:%s", 
			p.config.UniqueId, modeStr, endianStr))

		p.config.News_AddTail = make(chan []byte)
	}

	return nil
}


func (p *CRC)Run(){

	p.config.Events <-p.event_run

	switch p.config.Mode{
	case FILTER:
		go func(){
			defer p.reactiveDestruct()
			for {
				select{
				case mb, ok := <-p.config.Raws:
					if !ok{
						return
					}else{
						p.filter(mb)
					}
				}
			}
		}()
	case ADDTAIL:
		go func(){
			defer p.reactiveDestruct()
			for {
				select{
				case raw, ok := <-p.config.Raws:
					if !ok{
						return
					}else{
						p.addTail(raw)
					}
				}
			}
		}()
	}

}


//被动 - reactive
//被动析构是检测到Raws被上层关闭后的响应式析构操作
func (p *CRC)reactiveDestruct(){
	p.config.Events <-NewEvent(CRC_REACTIVE_DESTRUCT,p.config.UniqueId,"",
					fmt.Sprintf("[uid:%s]触发了隐式析构方法", p.config.UniqueId))

	//析构数据源

	switch p.config.Mode{
	case FILTER:
		close(p.config.News_Filter)
	case ADDTAIL:
		close(p.config.News_AddTail)
	}
}



func NewCRC() NodeAbstract {
	return &CRC{}
}


func init() {
	Register(CRC_NODE_NAME, NewCRC)
	logger.Info(fmt.Sprintf("预加载完成，[river-node type:%s]已预加载至package river_node.RNodes结构内", 
		CRC_NODE_NAME))
}


func (p *CRC)filter(mb []byte){
	if len(mb) < p.config.MinLen_Filter{
		p.config.Errors <-NewError(CRC_NOTPASS, p.config.UniqueId, hex.EncodeToString(mb),
			fmt.Sprintf("[uid:%s]待验证的modbus码不足所设定的最少位数：%d位，",
			p.config.UniqueId, p.config.MinLen_Filter))
		return
	}

	raw,crc := p.midModbus(mb) 

	if bytes.Equal(p.checkCRC16(raw[p.config.StartIndex_Filter:], p.config.Encoding), crc){
		if p.countor != 0{
			p.countor = 0
			p.config.Events <-NewEvent(CRC_RECOVERED, p.config.UniqueId, "",
				fmt.Sprintf("[uid:%s]已从第%d次CRC校验失败中恢复，当前系统设定的最大失败次数为%d",
				p.config.UniqueId, p.countor, p.config.Limit_Filter))
		}

		p.bytesHandler.Reset()
		p.bytesHandler.Write(raw)//通过crc校验后，再阉割掉后两位校验位
		p.config.News_Filter <-append([]byte{},p.bytesHandler.Bytes()...)

	}else if bytes.Equal(p.checkCRC16(raw[p.config.StartIndex_Filter:], !p.config.Encoding),crc){
		if p.countor != 0{
			p.config.Events <-NewEvent(CRC_RECOVERED,p.config.UniqueId,hex.EncodeToString(raw),
				fmt.Sprintf("[uid:%s]已从第%d次CRC校验失败中恢复，当前系统设定的最大失败次数为%d,但是当前"+
				"这一字节数组存在大小端颠倒的问题",p.config.UniqueId, p.countor, p.config.Limit_Filter))
			p.countor = 0
		}

		p.config.Events <-NewEvent(CRC_UPSIDEDOWN, p.config.UniqueId, hex.EncodeToString(raw),"")

		p.bytesHandler.Reset()
		p.bytesHandler.Write(raw)//通过crc校验后，再阉割掉后两位校验位
		p.config.News_Filter <-append([]byte{},p.bytesHandler.Bytes()...)

	}else if p.countor < p.config.Limit_Filter{
		p.countor++
		p.config.Errors <-NewError(CRC_NOTPASS, p.config.UniqueId, hex.EncodeToString(mb),
				fmt.Sprintf("[uid:%s]连续第%d次CRC校验失败，当前系统设定的最大连续失败次数为%d",
				p.config.UniqueId, p.countor, p.config.Limit_Filter))

		// 未通过校验的数据(错误数据)不再放入任何管道
	}else{
		p.config.Errors <-NewError(CRC_NOTPASS, p.config.UniqueId, hex.EncodeToString(mb),
				fmt.Sprintf("[uid:%s]CRC验证连续%d次失败，已超过系统设定的最大次数，系统设定的最大连续失败"+
				"次数为%d",p.config.UniqueId, p.countor,p.config.Limit_Filter))
		p.countor =0

		// 未通过校验的数据(错误数据)不再放入任何管道)

		p.config.Events <-p.event_fused
	}
}

//为传入的字节切片末尾添加校验位(ADDTAIL)
func (p *CRC)addTail(raw []byte){
	var tail []byte
	if tail = p.checkCRC16(raw, p.config.Encoding); tail ==nil{
		p.config.Errors <-NewError(CRC_NULLTAIL, p.config.UniqueId, hex.EncodeToString(raw), 
			fmt.Sprintf("[uid:%s](ADDTAIL模式)生成了空的校验码,问题校验码", p.config.UniqueId))
		return
	}

	raw = append(raw,tail...)
	p.bytesHandler.Reset()
	p.bytesHandler.Write(raw)
	p.config.News_AddTail <- append([]byte{},p.bytesHandler.Bytes()...)
}


/*------------以下是所需的功能方法-------------*/



//拆分需要验证的modbus码，拆分后会共享原始切片(mb)的底层数组
//不可以对两个返回值做任何直接的修改操作
func (p *CRC)midModbus(mb []byte) ([]byte, []byte){
	l :=len(mb)
	return mb[:l-2], mb[l-2:]
}


//计算出某个字节数组的crc16校验码
func (p *CRC)checkCRC16(data []byte, Endian bool) []byte {
	var crc16 uint16 = 0xffff

	for _, v := range data {
		n := uint8(uint16(v)^crc16)
		crc16 >>= 8
		crc16 ^= p.mbTable[n]
	}
	
	
	if Endian == BIGENDDIAN{
		p.bytesHandler.Reset()
		binary.Write(p.bytesHandler, binary.BigEndian, &crc16)
	}else{
		p.bytesHandler.Reset()
		binary.Write(p.bytesHandler, binary.LittleEndian, &crc16)
	}
	return append([]byte{},p.bytesHandler.Bytes()...)
}