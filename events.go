/** event结构类既可以作为river-node包数据传输所需的重要媒介"Event"，传入Events管道
 * 同时他也实现了Error方法，于是也可以作为错误传入Errors管道
 * 他内部的各个常量是不具有“正常/异常属性”的，就和physicalnode一样，不具有“正常/超限”属性
 * 而他被传入的管道是Events还是Errors才是决定他拥有“某种状态(正常/异常)”的时刻
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


func NewEvent(code int, uniqueId string, dataToString string, commit string) RN_event{
	if uniqueId =="" && code ==0 { return nil }

	if dataToString == "" {
		dataToString = "N/A"
	} else {
		commit = fmt.Sprintf("[Data not N/A] %s", commit)
	}

	if commit == "" {
		commit = "N/A"
	}

	eve :=&rn_event{
		UniqueId: 	uniqueId,
		Code: 		code,
		Data:		dataToString,
		Commit:		commit,
	}
	return eve
}

type RN_event interface{
	CodeString()string
	Description()(string,string,string,string)
}



type rn_event struct{
	UniqueId string
	Code 	 int
	Data	 string
	Commit 	 string
}


func (p *rn_event)CodeString()string{
	switch p.Code{
	case RAWSIMULATOR_RUN:
		return "RAWSIMULATOR_RUN"
	case RAWSIMULATOR_REACTIVEDESTRUCT:
		return "RAWSIMULATOR_REACTIVEDESTRUCT"
	case RAWSIMULATOR_PROACTIVEDESTRUCT:
		return "RAWSIMULATOR_PROACTIVEDESTRUCT"

	case HEARTBREATING_RUN:
		return "HEARTBREATING_RUN" 
	case HEARTBREATING_RECOVERED:
		return "HEARTBREATING_RECOVERED"
	case HEARTBREATING_TIMEOUT:
		return "HEARTBREATING_TIMEOUT"
	case HEARTBREATING_TIMERLIMITED:
		return "HEARTBREATING_TIMERLIMITED"
	case HEARTBREATING_FUSED:
		return "HEARTBREATING_FUSED"
	case HEARTBREATING_REACTIVEDESTRUCT:
		return "HEARTBREATING_REACTIVEDESTRUCT"
	case HEARTBREATING_PROACTIVEDESTRUCT:
		return "HEARTBREATING_PROACTIVEDESTRUCT"

	case CRC_RUN:
		return "CRC_RUN"
	case CRC_UPSIDEDOWN:
		return "CRC_UPSIDEDOWN"
	case CRC_NOTPASS:
		return "CRC_NOTPASS"
	case CRC_RECOVERED:
		return "CRC_RECOVERED"
	case CRC_FUSED:
		return "CRC_FUSED"
	case CRC_REACTIVEDESTRUCT:
		return "CRC_REACTIVEDESTRUCT"
	case CRC_PROACTIVEDESTRUCT:
		return "CRC_PROACTIVEDESTRUCT"

	case STAMPS_RUN:
		return "STAMPS_RUN"
	case STAMPS_REACTIVEDESTRUCT:
		return "STAMPS_REACTIVEDESTRUCT"
	case STAMPS_PROACTIVEDESTRUCT:
		return "STAMPS_PROACTIVEDESTRUCT"

	case ANOTHEREXAMPLE_TEST1:
		return "ANOTHEREXAMPLE_TEST1"
	case ANOTHEREXAMPLE_TEST2:
		return "ANOTHEREXAMPLE_TEST2"
	case ANOTHEREXAMPLE_TEST3:
		return "ANOTHEREXAMPLE_TEST3"
	case ANOTHEREXAMPLE_ERR:
		return "ANOTHEREXAMPLE_ERR"
	default:
		return "UNKNOWN_DEFINE"
	}
}


func (p *rn_event)Description()(string,string,string,string){
	return p.CodeString(), p.UniqueId, p.Data, p.Commit
}

