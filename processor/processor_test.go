package processor

import (
	"fmt"
	"os"
	"os/signal"
	"testing"
)

func TestNewProcessor(t *testing.T) {

	//s:=db.InitDB("127.0.0.1:27017")
	//defer s.Close()
	//
	//p:= NewProcessor(1)
	//p.Run()

	//p.AddData(&MonitorData{Id:"1",TimeStamp:time.Now()})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

    //go p.Wait()
	fmt.Println("Listen2")
    for{
		select {
		case <-signals:
			fmt.Println("Notify Processor")
			//p.signal<-sig
			return
		}
	}


}