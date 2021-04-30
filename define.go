package river_node

const (
	TIMER_STOPBEFOREEXPIRE = true
	TIMER_STOPAFTEREXPIRE = false
)

//crc
const (
	BIGENDDIAN = true
	LITTLEENDDIAN = false
)

const (
	FILTER = iota
	ADDTAIL
)
//crc end

//stamps
const (
	HEADS = iota
	TAILS
	HEADSANDTAILS
)
//stamps end


//authcode
const (
	ENCODE = iota
	DECODE
)
//authcode end

//baitsfilter
const (
	KEEPHEAD = iota
	DROPHEAD
)
//baitsfilter