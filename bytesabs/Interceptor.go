/*拦截器*/    /*如CRC校验*/
package bytesabs

type interceptorAbstractFunc func() InterceptorAbstract


/*ToDo方法的返回值代表了是否需要被拦截*/
type InterceptorAbstract interface {
	Name() string
	Init(config Config) error
	ToDo([]byte) bool
}