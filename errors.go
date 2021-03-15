package river_node 

import(
	"fmt"
)

func NewError(code int, uniqueId string, dataToString string, commit string)error{
	return &rn_error{ NewEvent(code, uniqueId, dataToString, commit) }	
}

type rn_error struct{
	RN_event
} 

/*让event也实现golang内置接口error，从而让他也变成一个error*/
func (p *rn_error)Error() string {
	c, cs, uid, data, commit := p.Description()
	return fmt.Sprintf("[ERROR] Code: %d, CodeString: %s, UniqueId: %s, DataString: %s, "+
			  "Commit: %s", c, cs, uid, data, commit)
}
