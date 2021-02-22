/** signal结构类既可以作为river-node包数据传输所需的重要媒介"Signal"，传入Signals管道
 * 同时他也实现了Error方法，于是也可以作为错误传入Errors管道
 * 他内部的各个常量是不具有“正常/异常属性”的，就和physicalnode一样，不具有“正常/超限”属性
 * 而他被传入的管道是Signals还是Errors才是决定他拥有“某种状态(正常/异常)”的时刻
 */

package river_node

import(
	"fmt"
)

const (
	ANOTHEREXAMPLE_TEST1 = 999999
	ANOTHEREXAMPLE_TEST2 = 999998
	ANOTHEREXAMPLE_TEST3 = 999997
	ANOTHEREXAMPLE_ERR = 999996
)


const (
	HEARTBREATING_RUN = iota
	HEARTBREATING_REBUILD
	HEARTBREATING_NORMAL 
	HEARTBREATING_TIMEOUT //ERROR
	HEARTBREATING_DROPCONN
	CRC_RUN
	CRC_NORMAL
	CRC_UPSIDEDOWN
	CRC_REVERSEENDIAN
	CRC_NOTPASS //ERROR
	CRC_DROPCONN
	STAMPS_RUN
)


func NewSignal(code int, uniqueid string) Signal{
	s :=&signal{
		UniqueId: 		uniqueid
			Code: 		code
	}
	return s
}

type Signal interface{
	Description()(string, int, string)
}

type signal struct{
	UniqueId string
	Code int
}


func (p *signal)Description()(uniqueid string, code int, str string){
	if p.UniqueId =="" { return }

	uniqueid = p.UniqueId;    code = p.Code

	switch code{
	case HEARTBREATING_INIT:
		str ="HEARTBREATING_INIT" 
	case HEARTBREATING_REBUILD:
		str ="HEARTBREATING_REBUILD"
	case HEARTBREATING_NORMAL:
		str ="HEARTBREATING_NORMAL"
	case HEARTBREATING_TIMEOUT:
		str ="HEARTBREATING_TIMEOUT"
	case CRC_INIT:
		str ="CRC_INIT"
	case CRC_NORMAL:
		str ="CRC_NORMAL"
	case CRC_UPSIDEDOWN:
		str ="CRC_UPSIDEDOWN"
	case CRC_NOTPASS:
		str ="CRC_NOTPASS"
	case STAMPS_INIT:
		str ="STAMPS_INIT"
	case ANOTHEREXAMPLE_TEST1:
		str ="ANOTHEREXAMPLE_TEST1"
	case ANOTHEREXAMPLE_TEST2:
		str ="ANOTHEREXAMPLE_TEST2"
	case ANOTHEREXAMPLE_TEST3:
		str ="ANOTHEREXAMPLE_TEST3"
	case ANOTHEREXAMPLE_ERR:
		str ="ANOTHEREXAMPLE_ERR"
	}

	return 
}

