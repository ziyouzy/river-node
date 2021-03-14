package river_node 

import(
	"fmt"
)

func NewError(code int, uniqueid string, dataToString string, commit string)error{
	if uniqueid ==""&&code ==0 {
		return nil
	}
	
	err :=&event{
		Code: 		code,
		UniqueId: 	uniqueid,
		Data:		dataToString,
		Commit:		commit,
	}

	return err	
}

/*让event也实现golang内置接口error，从而让他也变成一个error*/
func (p *event)Error() (s string) {
	_, _, codeToString, dataToString, commit  := p.Description()
	
	if commit ==""{
		s = fmt.Sprintf("river-node error: UniqueId: %s, Error: %s",
						p.UniqueId, conststring)
	}else{
		s = fmt.Sprintf("river-node error: UniqueId: %s, Error: %s, Commit: %s",
						p.UniqueId, conststring, commit)
	}
	return
}
