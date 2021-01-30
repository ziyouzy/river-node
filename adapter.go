

/** 此包的职能边界是将存在的各种类型(字节切片、NodeDoAbs)
 * 与其所需的各类适配器(拦截器、过滤器、验证器、触发器等)
 * 进行合理的匹配、部署、预加载(类似)
 * 装配于各适配器的map中 
 * 装配的内容是函数类型，而非功能对象的结构体或接口
 * 这么做目的在于节省系统资源，实现了类似预加载的效果
 *
 * 当上层使用此包时，上层结构类内会用一个接口切片的字段承载从本包map拿到的对象进行真正的初始化
 * 同时因为其载体为切片而非map所以也可以控制切片内各个适配器的先后执行顺序
 * 
 * 而各个适配器的具体实现则需要独立的包来完成，如package heartbeating
 * 独立的包需要设计好实现了某个合适自己子门类的适配器接口
 * 并在进行初始化时先存入map实现预编译的效果
 * 最终实现上层的调用
 */


package zadapter


func BuildAdapter(Name string){
	switch Name{
	default:
		fmt.Println("Adapter Name:", Name)
	}
}

type AdapterAbstract interface {
	Name() string
	Init(config Config) error
}

type Config interface {
	Name() string
}