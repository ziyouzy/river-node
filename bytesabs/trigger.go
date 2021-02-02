/*触发器*/    /*如心跳包*/
package bytesabs

import(
    "github.com/ziyouzy/heartbreating"
)

type triggerAbstractFunc func() TriggerAbstract

/** 触发器会被装配在拦截器之后
 * ToDo方法不会有任何返回值
 * 所有的逻辑都会在触发器所属的模块实现
 */

type TriggerAbstract interface {
	Name() string
	Init(config Config) error
	ToDo()
}


/** 由于通信管道的存在似乎就无法为其设计初始化默认值的操作了
 * 但这个函数也是必须的，因为上层一定会遍历各个map
 * 从而识别并确认都有哪些已经注册并在册的预编译适配器
 */

func NewHeartbBreating() TriggerAbstract {
	return &HeartBeating{}
}


func init() {
	Register(HEARTBEATING_ADAPTER_NAME, NewHeartbBreating)
}