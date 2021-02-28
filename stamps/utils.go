package stamps

import(
	"time"
	"encoding/binary"
)
func NanoTimeStamp()[]byte{
	timestamp :=make([]byte,8)
	binary.BigEndian.PutUint64(timestamp, uint64(time.Now().UnixNano()))
	return timestamp
}

func TimeStamp()[]byte{
	timestamp :=make([]byte,4)
	binary.BigEndian.PutUint64(timestamp, uint64(time.Now().Unix()))
	return timestamp
}