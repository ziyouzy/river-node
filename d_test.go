package zadapter

import (
	"fmt"
	"time"
	"testing"
)


func TestTd(t *testing.T) {
	fmt.Println("test")
	var td time.Duration
	fmt.Println(td)

	if td == (0 * time.Second) {
		fmt.Println("true")
	}else{
		fmt.Println("false")
	}
}