package river_node 

import(
	"fmt"
)

func NewError(code int, uniqueid string, commit string)error{
	if uniqueid ==""&&code ==0 {
		return nil
	}
	
	err :=&event{
		UniqueId: 	uniqueid,
		Code: 		code,
		Commit:		commit,
	}
	return err	
}

/*让event也实现golang内置接口error，从而让他也变成一个error*/
func (p *event)Error() (s string) {
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
