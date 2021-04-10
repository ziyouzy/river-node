package river_node 

import(
	"fmt"
	"time"
)

//只是进行了一下简单的封装,error不存在时间字段，时间以log日志所生成的时间为准
func NewError(code int, uniqueId string, raw string, commit string)error{
	return NewEvent(code,uniqueId,raw,commit).ToError()
}


type err struct{
	Event
} 

func (p *err)Error() string {
	c, cs, uid, data, commit, t:= p.Event.Description()
	return fmt.Sprintf("[ERROR] Code: %d, CodeString: %s, UniqueId: %s, DataString: %s, "+
			  "Commit: %s, Time: %s", c, cs, uid, data, commit, time.Unix(0, t).Format("2006-01-02 15:04:05.000000000"))
}