package processor

import (
	"fmt"
	"kafka-consumer/db"
	. "kafka-consumer/lib"
	"os"
	"os/signal"
	"strconv"
	"testing"
	"time"
)

func TestNewProcessor(t *testing.T) {

	InitCfg()

	db.InitCache()

	s := db.InitMongoDB("127.0.0.1:27017")
	defer s.Close()

	m := db.InitMySQLDB(AppCfg.MySqlURL)
	defer m.Close()

	p := NewProcessor(1, 1)
	p.Run()

	p.AddData(&MonitorData{DeviceId: 1, TimeStamp: strconv.Itoa(int(time.Now().Unix())*1000 - 50000)})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	//go p.Wait()
	fmt.Println("Listen2")
	for {
		select {
		case <-signals:
			fmt.Println("Notify Processor")
			//p.signal<-sig
			return
		}
	}

}
