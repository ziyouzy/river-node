package testpkg

import(
    "fmt"
)

var (
    TestDataCreaterNews chan []byte
    HBRaws              chan struct{}
    CRCRaws             chan []byte
)
	
func TestDataCreaterConnectCRC(){
    TestDataCreaterNews = make(chan []byte)
    HBRaws              = make(chan struct{})
    CRCRaws             = make(chan []byte)
	go func(){
        for bytes := range TestDataCreaterNews{
            fmt.Println("source bytes is", bytes)
            HBRaws <- struct{}{}
            CRCRaws <- bytes	
        }
    }()
}

var (
    CRCPassNews    chan []byte     
    CRCNotPassNews chan []byte     
    StampsRaws     chan []byte    
)

func CRCConnectStamps(){
    CRCPassNews         = make(chan []byte)
    CRCNotPassNews      = make(chan []byte)
    StampsRaws          = make(chan []byte)
	go func(){
        for bytes := range CRCPassNews{
            StampsRaws <- bytes	
        }
	}()
	
	go func(){
        for bytes := range CRCNotPassNews{
            StampsRaws <- bytes	
        }
    }()
}