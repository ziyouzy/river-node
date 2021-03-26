package river_node

import(
	"runtime"
	"bytes"
	"strconv"
	"time"
	"encoding/binary"
)


func goID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}


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

func StringTimeStamp(res []byte, isNano bool)string{
	if isNano {
		return time.Unix(0, int64(binary.BigEndian.Uint64(res))).Format("2006-01-02 15:04:05.000000000")
	}else{
		return time.Unix(int64(binary.BigEndian.Uint64(res)), 0).Format("2006-01-02 15:04:05")
	}
}