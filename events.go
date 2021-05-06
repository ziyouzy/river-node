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
	//拿hb举例最重要的是HEARTBREATING_TIMERLIMITED，如果main层检测到这个信号后必然会显式析构其所属的river
	//而不是说这个hb直接去析构其所属的river
	//main会先通过uid锁定问题river，再去析构他
	//hb做的仅仅是发送此信号给main层，而不会去做其他的事(如析构自身或者析构其所在的river)	
	HEARTBREATING_RUN = iota
	HEARTBREATING_FUSED //很重要，需要告诉给系统这个错误，或许他应该改名为FAIL，让他的语义代表一种节点的“panic”即可
	HEARTBREATING_TIMEOUT //为返回给系统看的ERROR，只需要计入日志
	HEARTBREATING_TIMERLIMITED //为返回给系统看的ERROR，只需要计入日志，但是他会连带触发HEARTBREATING_FUSED
	HEARTBREATING_RECOVERED //这是成功，理论上不需要告诉系统成功只需要告诉系统失败，同时这个东西也没必要告诉用户，所以他肯能会被砍掉了
	HEARTBREATING_REACTIVE_DESTRUCT //这其实也是返回给系统的成功，系统不需要，但是对于用户，river层的这种析构事件才会有些意义
	HEARTBREATING_PROACTIVE_DESTRUCT //不存在这个
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
	BAITSFILTER_FUSED
	BAITSFILTER_RUN
	BAITSFILTER_HEADUNDEFINE
	BAITSFILTER_LENAUTHFAIL
	BAITSFILTER_REACTIVE_DESTRUCT
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

	case BAITSFILTER_FUSED:
		return "BAITSFILTER_FUSED"
	case BAITSFILTER_RUN:
		return "BAITSFILTER_RUN"
	case BAITSFILTER_HEADUNDEFINE:
		return "BAITSFILTER_HEADUNDEFINE"
	case BAITSFILTER_LENAUTHFAIL:
		return "BAITSFILTER_LENAUTHFAIL"
	case BAITSFILTER_REACTIVE_DESTRUCT:
		return "BAITSFILTER_REACTIVE_DESTRUCT"

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


