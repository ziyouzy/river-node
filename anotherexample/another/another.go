package another

import(
	"fmt"
)


type AnotherStruct struct{
	TimeStamp int64
	Tip string

	TestSl []float64

	RawSl []byte
}


/*-------------------*/


type AnotherFunc func([]byte)

func testFunc(sl []byte){
	fmt.Println("this is the test of zadapter/anotherexample/another/"+
	            "struct of AnoterFunc/testFunc(), sl is:",sl)
}