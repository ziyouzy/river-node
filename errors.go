 package river_node 


//只是进行了一下简单的封装,error不存在时间字段，时间以log日志所生成的时间为准
func NewError(code int, uniqueId string, raw string, commit string)error{
	return NewEvent(code,uniqueId,raw,commit).ToError()
}


