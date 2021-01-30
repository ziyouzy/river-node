/*触发器*/    /*如心跳包*/
package bytesabs

type triggerAbstractFunc func() TriggerAbstract

/** 触发器会被装配在拦截器之后
 * ToDo方法不会有任何返回值
 * 所有的逻辑都会在触发器所属的模块实现
 */

type TriggerAbstract interface {
	Name() string
	Init(config Config) error
	ToDo([]byte)
}