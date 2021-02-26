package river_node 

import(
	"fmt"
)

func NewError(code int, uniqueid string, commit string)error{
	if uniqueid ==""&&code ==0 {
		//fmt.error
		return nil
	}
	
	e :=&signal{
		UniqueId: 	uniqueid,
		Code: 		code,
		Commit:		commit,
	}
	return e	
}

/*让signal也实现golang内置接口error，从而让他也变成一个error*/
func (p *signal)Error() (s string) {
	_, _, conststring, commit  := p.Description()
	
	if commit ==""{
		s = fmt.Sprintf("river-node error: UniqueId: %s, Error: %s",
						p.UniqueId, conststring)
	}else{
		s = fmt.Sprintf("river-node error: UniqueId: %s, Error: %s, Commit: %s",
						p.UniqueId, conststring, commit)
	}
	return
}
