//river是node的外层，river的自定义属性强于node，因此只设计了一个接口

package river_node


type RiverAbstract interface {
	Name() string
	Run()
	Destruct()   
}
