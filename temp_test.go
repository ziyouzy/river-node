package river_node

import (
	"fmt"
	"time"
	"testing"	
	"bytes"

	//"github.com/ziyouzy/go-authcode"

)


func TestTttt(t *testing.T) {
	// key :="key/salt"
	// dynamickeylen := 1
	// expiry :=2
	// ac :=go_authcode.New(key,dynamickeylen,expiry)
	// defer ac.CloseSafe()

	// for{
	// 	encode ,_ :=ac.Encode([]byte("1234567890abcdefaaaaaa//***/+++____"),"")
	// 	fmt.Println("encode_str:",string(encode))//[]byte
	// 	decode ,_ :=ac.Decode(nil,string(encode))    
	// 	fmt.Println("decode_str:",string(decode))//[]byte
	// 	time.Sleep(1*time.Second)
	// }


	baits := []byte{0x01,0x02,0x03,0x04,0x01,0x02,0x03,0x04,0x01,0x02,0x03,0x04,0x01,0x02,0x03,0x04}
	h := bytes.NewBuffer([]byte{})
	h.Write(baits)

	ch := make(chan []byte)
	go func(){
		b := <-ch
		fmt.Println("b:",b)
		time.Sleep(5*time.Second)
		fmt.Println("b:",b)
	}()
	go func(){
		testb := h.Bytes()
		fmt.Println("testb:",testb)
		ch <-testb
		time.Sleep(time.Second)
		baitsx :=[]byte{0x04,0x05,0xff}
		h.Reset()
		h.Write(baitsx)
		fmt.Println("testb:",testb)
	}() 

	select{}
}