package bytes


/*触发器*/
var Triggers = make(map[string]triggerAbstractFunc)
func TriggerRegister(Name string, F triggerAbstractFunc) {
	if Triggers[Name] != nil {
		//panic("logger: logger adapter " + adapterName + " already registered!")
	}
	
	if F == nil {
		//panic("logger: logger adapter " + adapterName + " is nil!")
	}

	Triggers[Name] = F
}


/*拦截器*/
var Interceptors = make(map[string]interceptorAbstractFunc)
func InterceptorRegister(Name string, F interceptorAbstractFunc) {
	if Interceptors[Name] != nil {
		//panic("logger: logger adapter " + adapterName + " already registered!")
	}
	
	if F == nil {
		//panic("logger: logger adapter " + adapterName + " is nil!")
	}

	Interceptors[Name] = F
}