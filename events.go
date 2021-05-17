//river-node层不再需要events，只需要给我自己看的logs
//但是本层不仅仅会丢弃Events，也会丢弃Logs了
//这些Err传到到了“很接近用户”的层之后，如果有必要，RNerrors可以进而转化成面向用户的events
//同时在他的上层(表面上是river层实际则往往是main层)，也可以将某些有必要的Err，通过logger进行记录
//这也都会与"断开某个river"一样，或者是录入mysql一样，变成上层被动响应的一种分支
//等设计到了上述这步，才会是这也是plan for fail，not success的具体设计体现
package river_node 

import(
	"fmt"
	"time"
	"strings"
)


const (
	HEARTBREATING_RUN =iota
	HEARTBREATING_PANICH 
	HEARTBREATING_TIMEOUT
	HEARTBREATING_TIMERLIMITED
	HEARTBREATING_RECOVERED
	HEARTBREATING_REACTIVE_DESTRUCT
	CRC_RUN
	CRC_PANICH 
	CRC_UPSIDEDOWN
	CRC_NOTPASS
	CRC_RECOVERED
	CRC_CHECKFAIL
	CRC_REACTIVE_DESTRUCT
	STAMPS_RUN
	STAMPS_REACTIVE_DESTRUCT
	AUTHCODE_RUN
	AUTHCODE_PANICH
	AUTHCODE_ENCODE_FAIL
	AUTHCODE_ENCODE_RECOVERED
	AUTHCODE_DECODE_FAIL
	AUTHCODE_DECODE_RECOVERED
	AUTHCODE_REACTIVE_DESTRUCT
	BAITSFILTER_RUN
	BAITSFILTER_HEADAUTHFAIL
	BAITSFILTER_LENAUTHFAIL
	BAITSFILTER_REACTIVE_DESTRUCT
)

type EventAbs interface{
	CodeString()string
	CodeInt() int
	Description()(int, string,string,string,error,string,int64)
	ParentRiverUID() string

	Error() string
	Unwrap() error
}

func NewEvent(code int, uniqueId string, 
	commentRaw string, commentError error, commentText string) EventAbs{

	//code默认为0

	if uniqueId =="" { return nil }

	if commentRaw == "" {commentRaw = "commentRaw is nil"}

	//if commentError ==nil {...}

	if commentText == "" {commentText = "commentText is nil"}


	return &Event{
		UniqueId: 		uniqueId,
		Code: 			code,

		CommentRaw:		commentRaw,
		CommentError:	commentError,
		CommentText:	commentText,

		TimeStamp:		time.Now().UnixNano(),
	}
	
}

//似乎需要添加error字段，并实现Unwarp方法，从而让Event能够“功能完整的包裹"一个error
type Event struct{
	UniqueId 		string
	Code 	 		int

	CommentRaw		string
	CommentError	error
	CommentText		string
	
	TimeStamp		int64
}

func (p *Event)CodeInt()int{
	return p.Code
}

func (p *Event)ParentRiverUID() string{
	//strings.SplitN("a,b,c", ",", 2) // ["a", "b,c"]
    //strings.SplitN("a,b,c,d", ",", 1) // ["a,b,c", "d"]
	return strings.SplitN(p.UniqueId,":",1)[0]
}

func (p *Event)CodeString()string{
	switch p.Code{
	case HEARTBREATING_RUN:
		return "HEARTBREATING_RUN"
	case HEARTBREATING_PANICH:
		return "HEARTBREATING_PANICH"
	case HEARTBREATING_TIMEOUT:
		return "HEARTBREATING_TIMEOUT"
	case HEARTBREATING_TIMERLIMITED:
		return "HEARTBREATING_TIMERLIMITED"
	case HEARTBREATING_RECOVERED:
		return "HEARTBREATING_RECOVERED"
	case HEARTBREATING_REACTIVE_DESTRUCT:
		return "HEARTBREATING_REACTIVE_DESTRUCT"

	case CRC_RUN:
		return "CRC_RUN"
	case CRC_PANICH: 
		return "CRC_PANICH"
	case CRC_UPSIDEDOWN:
		return "CRC_UPSIDEDOWN"
	case CRC_NOTPASS:
		return "CRC_NOTPASS"
	case CRC_CHECKFAIL:
		return "CRC_CHECKFAIL"
	case CRC_RECOVERED:
		return "CRC_RECOVERED"
	case CRC_REACTIVE_DESTRUCT:
		return "CRC_REACTIVE_DESTRUCT"

	case AUTHCODE_RUN:
		return "AUTHCODE_RUN"
	case AUTHCODE_PANICH:
		return "AUTHCODE_PANICH"
	case AUTHCODE_ENCODE_FAIL:
		return "AUTHCODE_ENCODE_FAIL"
	case AUTHCODE_DECODE_FAIL:
		return "AUTHCODE_DECODE_FAIL"
	case AUTHCODE_RECOVERED:
		return "AUTHCODE_RECOVERED"
	case AUTHCODE_REACTIVE_DESTRUCT:
		return "AUTHCODE_REACTIVE_DESTRUCT"

	case STAMPS_RUN:
		return "STAMPS_RUN"
	case STAMPS_REACTIVE_DESTRUCT:
		return "STAMPS_REACTIVE_DESTRUCT"

	case BAITSFILTER_RUN:
		return "BAITSFILTER_RUN"
	case BAITSFILTER_HEADAUTHFAIL:
		return "BAITSFILTER_HEADAUTHFAIL"
	case BAITSFILTER_LENAUTHFAIL:
		return "BAITSFILTER_LENAUTHFAIL"
	case BAITSFILTER_REACTIVE_DESTRUCT:
		return "BAITSFILTER_REACTIVE_DESTRUCT"

	default:
		return "UNKNOWN_DEFINE"
	}
}

func (p *Event)Description()(int, string, string, 
	string, error, string, int64){

		return p.Code, p.CodeString(), p.UniqueId, 
			p.CommentRaw, p.CommentError, p.CommentText, p.TimeStamp
}

func (p *Event)Error() string {

	c, cStr, uid, rawStr, err, textStr, t:= p.Description()

	errStr :="commentError is nil";	if err !=nil{errStr =err.Error()}

	return fmt.Sprintf("[RIVER_NODE_ERROR] CodeInt: %d, CodeString: %s, UniqueId: %s, "+
		"CommentRawString: %s, CommentErrorString: %s, CommentTextString: %s,  "+
		"[Time: %s]", c, cStr, uid, rawStr, errStr, textStr, 
		time.Unix(0, t).Format("2006-01-02 15:04:05.000000000"))

}


func (p *Event) Unwrap() error {
	return p.CommentError
}