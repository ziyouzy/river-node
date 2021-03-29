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
	HEARTBREATING_REACTIVE_DESTRUCT
	HEARTBREATING_PROACTIVE_DESTRUCT
	CRC_RUN
	CRC_FUSED 
	CRC_UPSIDEDOWN //一般为SIGNAL
	CRC_NOTPASS //一般为ERROR
	CRC_RECOVERED
	CRC_NULLTAIL
	CRC_REACTIVE_DESTRUCT
	CRC_PROACTIVE_DESTRUCT
	STAMPS_RUN
	STAMPS_REACTIVE_DESTRUCT
	STAMPS_PROACTIVE_DESTRUCT
	RAWSIMULATOR_RUN
	RAWSIMULATOR_REACTIVE_DESTRUCT
	RAWSIMULATOR_PROACTIVE_DESTRUCT
	AUTHCODE_RUN
	AUTHCODE_FUSED
	AUTHCODE_ENCODE_FAIL
	AUTHCODE_DECODE_FAIL
	AUTHCODE_ENCODE_RECOVERED
	AUTHCODE_DECODE_RECOVERED
	AUTHCODE_PROACTIVE_DESTRUCT
	AUTHCODE_REACTIVE_DESTRUCT
)

type Event interface{
	CodeString()string
	Code() int
	Description()(int, string,string,string,string)
	ToError() error
}

func NewEvent(code int, uniqueId string, dataToString string, commit string) Event{
	if uniqueId =="" && code ==0 { return nil }

	if dataToString == "" {
		dataToString = "N/A"
	} else {
		commit = fmt.Sprintf("[Data not N/A] %s", commit)
	}

	if commit == "" {
		commit = "N/A"
	}

	e :=&eve{
		uniqueId: 	uniqueId,
		code: 		code,
		data:		dataToString,
		commit:		commit,
	}
	return e
}


type eve struct{
	uniqueId string
	code 	 int
	data	 string
	commit 	 string
}

func (p *eve)Code()int{
	return p.code
}

func (p *eve)CodeString()string{
	switch p.Code(){
	case RAWSIMULATOR_RUN:
		return "RAWSIMULATOR_RUN"
	case RAWSIMULATOR_REACTIVE_DESTRUCT:
		return "RAWSIMULATOR_REACTIVE_DESTRUCT"
	case RAWSIMULATOR_PROACTIVE_DESTRUCT:
		return "RAWSIMULATOR_PROACTIVE_DESTRUCT"

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
	case HEARTBREATING_REACTIVE_DESTRUCT:
		return "HEARTBREATING_REACTIVE_DESTRUCT"
	case HEARTBREATING_PROACTIVE_DESTRUCT:
		return "HEARTBREATING_PROACTIVE_DESTRUCT"

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
	case CRC_REACTIVE_DESTRUCT:
		return "CRC_REACTIVE_DESTRUCT"
	case CRC_PROACTIVE_DESTRUCT:
		return "CRC_PROACTIVE_DESTRUCT"

	case STAMPS_RUN:
		return "STAMPS_RUN"
	case STAMPS_REACTIVE_DESTRUCT:
		return "STAMPS_REACTIVE_DESTRUCT"
	case STAMPS_PROACTIVE_DESTRUCT:
		return "STAMPS_PROACTIVE_DESTRUCT"

	case AUTHCODE_RUN:
		return "AUTHCODE_RUN"
	case AUTHCODE_ENCODE_FAIL:
		return "AUTHCODE_ENCODE_FAIL"
	case AUTHCODE_DECODE_FAIL:
		return "AUTHCODE_DECODE_FAIL"
	case AUTHCODE_ENCODE_RECOVERED:
		return "AUTHCODE_ENCODE_RECOVERED"
	case AUTHCODE_DECODE_RECOVERED:
		return "AUTHCODE_DECODE_RECOVERED"
	case AUTHCODE_PROACTIVE_DESTRUCT:
		return "AUTHCODE_PROACTIVE_DESTRUCT"
	case AUTHCODE_REACTIVE_DESTRUCT:
		return "AUTHCODE_REACTIVE_DESTRUCT"
	case AUTHCODE_FUSED:
		return "AUTHCODE_FUSED"

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

func (p *eve)Description()(int, string, string, string, string){
	return p.Code(), p.CodeString(), p.uniqueId, p.data, p.commit
}

func (p *eve)ToError()error{
	return &err{p}
}

//------------


type err struct{
	Event
} 

func (p *err)Error() string {
	c, cs, uid, data, commit := p.Event.Description()
	return fmt.Sprintf("[ERROR] Code: %d, CodeString: %s, UniqueId: %s, DataString: %s, "+
			  "Commit: %s", c, cs, uid, data, commit)
}