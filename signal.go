/** signal结构类既可以作为river-node包数据传输所需的重要媒介"Signal"，传入Signals管道
 * 同时他也实现了Error方法，于是也可以作为错误传入Errors管道
 * 他内部的各个常量是不具有“正常/异常属性”的，就和physicalnode一样，不具有“正常/超限”属性
 * 而他被传入的管道是Signals还是Errors才是决定他拥有“某种状态(正常/异常)”的时刻
 */

package river_node

import(
//	"fmt"
)

const (
	ANOTHEREXAMPLE_TEST1 = 999999
	ANOTHEREXAMPLE_TEST2 = 999998
	ANOTHEREXAMPLE_TEST3 = 999997
	ANOTHEREXAMPLE_ERR = 999996
)

//package river-node占用编号范围是0~199
//已知目前usrio-808占用编码范围200~219
const (
	HEARTBREATING_RUN = iota
	HEARTBREATING_REBUILD
	HEARTBREATING_NORMAL 
	HEARTBREATING_TIMEOUT //一般为ERROR
	HEARTBREATING_TIMERLIMITED //一般为ERROR
	HEARTBREATING_RECOVERED
	HEARTBREATING_PANIC
	CRC_RUN
	CRC_NORMAL
	CRC_UPSIDEDOWN //一般为SIGNAL
	CRC_NOTPASS //一般为ERROR
	CRC_RECOVERED
	CRC_PANIC
	STAMPS_RUN
	TESTDATACREATER_RUN
)


func NewSignal(code int, uniqueid string, commit string) Signal{
	if uniqueid ==""&&code ==0 {
		//fmt.error
		return nil
	}

	s :=&signal{
		UniqueId: 	uniqueid,
		Code: 		code,
		Commit:		commit,
	}
	return s
}

type Signal interface{
	Description()(string, int, string, string)
}



type signal struct{
	UniqueId string
	Code 	 int
	Commit 	 string
}


func (p *signal)Description()(uniqueid string, code int, conststring string, commit string){
	uniqueid = p.UniqueId;	code = p.Code;	commit =p.Commit

	switch code{
	case TESTDATACREATER_RUN:
		conststring ="TESTDATACREATER_RUN"
	case HEARTBREATING_RUN:
		conststring ="HEARTBREATING_RUN" 
	case HEARTBREATING_REBUILD:
		conststring ="HEARTBREATING_REBUILD"
	case HEARTBREATING_NORMAL:
		conststring ="HEARTBREATING_NORMAL"
	case HEARTBREATING_RECOVERED:
		conststring ="HEARTBREATING_RECOVERED"
	case HEARTBREATING_TIMEOUT:
		conststring ="HEARTBREATING_TIMEOUT"
	case HEARTBREATING_TIMERLIMITED:
		conststring ="HEARTBREATING_TIMERLIMITED"
	case HEARTBREATING_PANIC:
		conststring ="HEARTBREATING_PANIC"

	case CRC_RUN:
		conststring ="CRC_RUN"
	case CRC_NORMAL:
		conststring ="CRC_NORMAL"
	case CRC_UPSIDEDOWN:
		conststring ="CRC_UPSIDEDOWN"
	case CRC_NOTPASS:
		conststring ="CRC_NOTPASS"
	case CRC_RECOVERED:
		conststring ="CRC_RECOVERED"
	case CRC_PANIC:
		conststring ="CRC_PANIC"

	case STAMPS_RUN:
		conststring ="STAMPS_RUN"

	case ANOTHEREXAMPLE_TEST1:
		conststring ="ANOTHEREXAMPLE_TEST1"
	case ANOTHEREXAMPLE_TEST2:
		conststring ="ANOTHEREXAMPLE_TEST2"
	case ANOTHEREXAMPLE_TEST3:
		conststring ="ANOTHEREXAMPLE_TEST3"
	case ANOTHEREXAMPLE_ERR:
		conststring ="ANOTHEREXAMPLE_ERR"
	}

	return 
}

