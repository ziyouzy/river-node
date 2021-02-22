package river_node 

/*让signal也实现golang内置接口error，从而让他也变成一个error*/
func (p *signal)Error() (s string) {
	_, _, detail := p.Description() 
	s = fmt.Sprintf("river-node error: UniqueId: %s, Detail: %s",p.UniqueId, detail)
	//log(e)?
	return
}

func NewError(code int, uniqueid string)error{
	e :=&signal{
		UniqueId: 		uniqueid
			Code: 		code
	}
	return e	
}