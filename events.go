/** event结构类既可以作为river-node包数据传输所需的重要媒介"Event"，传入Events管道
 * 同时他也实现了Error方法，于是也可以作为错误传入Errors管道
 * 他内部的各个常量是不具有“正常/异常属性”的，就和physicalnode一样，不具有“正常/超限”属性
 * 而他被传入的管道是Events还是Errors才是决定他拥有“某种状态(正常/异常)”的时刻
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
	HEARTBREATING_FUSED 
	HEARTBREATING_TIMEOUT //一般为ERROR
	HEARTBREATING_TIMERLIMITED //一般为ERROR
	HEARTBREATING_RECOVERED
	HEARTBREATING_REACTIVEDESTRUCT
	HEARTBREATING_PROACTIVEDESTRUCT
	CRC_RUN
	CRC_FUSED 
	CRC_UPSIDEDOWN //一般为SIGNAL
	CRC_NOTPASS //一般为ERROR
	CRC_RECOVERED
	CRC_NULLTAIL
	CRC_REACTIVEDESTRUCT
	CRC_PROACTIVEDESTRUCT
	STAMPS_RUN
	STAMPS_REACTIVEDESTRUCT
	STAMPS_PROACTIVEDESTRUCT
	RAWSIMULATOR_RUN
	RAWSIMULATOR_REACTIVEDESTRUCT
	RAWSIMULATOR_PROACTIVEDESTRUCT
)


func NewEvent(code int, uniqueId string, dataToString string, commit string) Event{
	if uniqueid ==""&&code ==0 {
		return nil
	}

	eve :=&event{
		UniqueId: 	uniqueid,
		Code: 		code,
		Data:		dataToString,
		Commit:		commit,
	}
	return eve
}

type Event interface{
	Description()(string, int, string, string, string)
}



type event struct{
	UniqueId string
	Code 	 int
	Data	 string
	Commit 	 string
}


func (p *event)Description()(uniqueId string, code int, codeToString string, dataToString string, commit string){
	uniqueId = p.UniqueId;		dataToString = p.Data;		commit = p.Commit

	code = p.Code
	switch code{
	case RAWSIMULATOR_RUN:
		codeToString = "RAWSIMULATOR_RUN"
	case RAWSIMULATOR_REACTIVEDESTRUCT:
		codeToString = "RAWSIMULATOR_REACTIVEDESTRUCT"
	case RAWSIMULATOR_PROACTIVEDESTRUCT:
		codeToString = "RAWSIMULATOR_PROACTIVEDESTRUCT"

	case HEARTBREATING_RUN:
		codeToString = "HEARTBREATING_RUN" 
	case HEARTBREATING_RECOVERED:
		codeToString = "HEARTBREATING_RECOVERED"
	case HEARTBREATING_TIMEOUT:
		codeToString = "HEARTBREATING_TIMEOUT"
	case HEARTBREATING_TIMERLIMITED:
		codeToString = "HEARTBREATING_TIMERLIMITED"
	case HEARTBREATING_FUSED:
		codeToString = "HEARTBREATING_FUSED"
	case HEARTBREATING_REACTIVEDESTRUCT:
		codeToString = "HEARTBREATING_REACTIVEDESTRUCT"
	case HEARTBREATING_PROACTIVEDESTRUCT:
		codeToString = "HEARTBREATING_PROACTIVEDESTRUCT"

	case CRC_RUN:
		codeToString = "CRC_RUN"
	case CRC_UPSIDEDOWN:
		codeToString = "CRC_UPSIDEDOWN"
	case CRC_NOTPASS:
		codeToString = "CRC_NOTPASS"
	case CRC_RECOVERED:
		codeToString = "CRC_RECOVERED"
	case CRC_FUSED:
		codeToString = "CRC_FUSED"
	case CRC_REACTIVEDESTRUCT:
		codeToString = "CRC_REACTIVEDESTRUCT"
	case CRC_PROACTIVEDESTRUCT:
		codeToString = "CRC_PROACTIVEDESTRUCT"

	case STAMPS_RUN:
		codeToString = "STAMPS_RUN"
	case STAMPS_REACTIVEDESTRUCT:
		codeToString = "STAMPS_REACTIVEDESTRUCT"
	case STAMPS_PROACTIVEDESTRUCT:
		codeToString = "STAMPS_PROACTIVEDESTRUCT"

	case ANOTHEREXAMPLE_TEST1:
		codeToString ="ANOTHEREXAMPLE_TEST1"
	case ANOTHEREXAMPLE_TEST2:
		codeToString ="ANOTHEREXAMPLE_TEST2"
	case ANOTHEREXAMPLE_TEST3:
		codeToString ="ANOTHEREXAMPLE_TEST3"
	case ANOTHEREXAMPLE_ERR:
		codeToString ="ANOTHEREXAMPLE_ERR"
	}

	return 
}

