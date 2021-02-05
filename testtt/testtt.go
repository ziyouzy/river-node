package testtt

import(
	"zadapter"
	_ "zadapter/heartbreating"

	"fmt"
)


func testRun(){
	fmt.Println("test1")
	zadapter.NewLogger()
	zadapter.Logger_Info("test Log Info")

	fmt.Println(len(zadapter.Adapters))
}