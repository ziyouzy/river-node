/** event结构类既可以作为river-node包数据传输所需的重要媒介"Event"，传入Events管道
 * 同时他也实现了Error方法，于是也可以作为错误传入Errors管道
 * 他内部的各个常量是不具有“正常/异常属性”的，就和physicalnode一样，不具有“正常/超限”属性
 * 而他被传入的管道是Events还是Errors才是决定他拥有“某种状态(正常/异常)”的时刻
 */

package river_node

import(
	"fmt"
	"time"
)


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
	CodeInt() int
	Description()(int, string,string,string,string,int64)
	ToError() error
}

//有可能会存在将一个纯粹的error转化为这里的error的情况
//比如我自己所设计的phyicalnode包，以及authcode包，他们在运行是会返回原生的error数据类型
//而将他们river-node化后，必然需要处理这些error，禁止把这些error等同于第三个参数
//再或者说，无论error还是event的第三个字段只能放入出问题的数据，err的具体内容是属于commit参数的一部分
func NewEvent(code int, uniqueId string, raw string, commit string) Event{
	if uniqueId =="" && code ==0 { return nil }

	if raw == "" {
		raw = "N/A"
	} else {
		commit = fmt.Sprintf("[Data not N/A] %s", commit)
	}

	if commit == "" {
		commit = "N/A"
	}

	e :=&eve{
		UniqueId: 	uniqueId,
		Code: 		code,
		Data:		raw,
		Commit:		commit,

		TimeStamp:	time.Now().UnixNano(),
	}
	return e
}


type eve struct{
	UniqueId 	string
	Code 	 	int
	Data	 	string
	Commit 	 	string

	TimeStamp	int64
}

func (p *eve)CodeInt()int{
	return p.Code
}

func (p *eve)CodeString()string{
	switch p.Code{
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

	default:
		return "UNKNOWN_DEFINE"
	}
}

func (p *eve)Description()(int, string, string, string, string, int64){
	return p.Code, p.CodeString(), p.UniqueId, p.Data, p.Commit, p.TimeStamp
}

func (p *eve)ToError()error{
	return &err{p}
}


