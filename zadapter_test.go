package zadapter

import (
	"fmt"
	"time"
	"testing"	
)

func TestZadapter(t *testing.T) {
	NewLogger()
	Logger_Info("test Log Info")


	testRawCH := make(chan []byte)

	mainTestSignalReceiverCh :=make(chan int)
	mainTestSlotSenderCh :=make(chan struct{})

	
	heartBeatingAbs := Adapters[HEARTBEATING_ADAPTER_NAME]()
	heartBeatingConfig := &HeartBeatingConfig{
		timeout : 8 * time.Second,
		uniqueId : "testCh",
		signalChan : mainTestSignalReceiverCh,
		slotChan : mainTestSlotSenderCh,
	}
	if err := heartBeatingAbs.Init(heartBeatingConfig); err == nil {
		Logger_Info("test adapter init success")
	}

	heartBeatingAbs.Run()


	go func(){
		defer close(testRawCH)
		for {
			INCH <- []byte{0x01, 0x02, 0x03,}
			time.Sleep(time.Second)
		}
	}()

	go func(){
		defer close(mainTestSlotSenderCh)
		for bytes := range testRawCH{
			fmt.Println("bytes is", bytes)
			mainTestSlotSenderCh <- struct{}{}
		}
	}()

	go func(){
		defer close(mainTestSignalReceiverCh)
		for signal := range mainTestSignalReceiverCh{
			switch signal{
			case HEARTBREATING_NORMAL:
				fmt.Println("signal:", "HEARTBREATING_NORMAL")
			case HEARTBREATING_TIMEOUT:
				fmt.Println("signal:", "HEARTBREATING_TIMEOUT")
			}
			
		}
	}()
	
	select{}
}