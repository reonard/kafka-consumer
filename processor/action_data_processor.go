package processor

import (
	"fmt"
	"kafka-consumer/db"
	"os"
	"os/signal"
	"strconv"
	"sync"
)

const (
	STATUS_PENDING = iota
	STATUS_PUSHED
	STATUS_ACK
)

var actionWg sync.WaitGroup

type ActionWorker struct {
	workerId int
	signals  chan os.Signal
	actData  chan []*ActionData
	error    chan error
}

func (w *ActionWorker) processData(data ...*ActionData) {

	for _, dt := range data {

		var err error

		unixTime, err := strconv.Atoi(dt.TimeStamp)
		if err != nil {
			fmt.Println("Atoi错误")
			fmt.Println(unixTime)
			continue
		}
		// 更新下发任务状态
		err = db.ExecuteSQL(
			"UPDATE t_issue_status SET issue_status = ?, success = ?, status_time = FROM_UNIXTIME(?) WHERE id = ? and device_id= ?",
			STATUS_ACK, dt.Success, unixTime/1000, dt.MsgID, dt.DeviceId)

		if err != nil {
			fmt.Println(err)
		}
	}
}

func (w *ActionWorker) Process() {

	fmt.Printf("Worker %d start processing \n", w.workerId)

Loop:
	for {
		select {

		case dt, ok := <-w.actData:
			if ok {
				w.processData(dt...)
			}
		case <-w.signals:
			fmt.Printf("Worker %d Received Signal, Flushing...\n", w.workerId)
			break Loop
		}
	}

	fmt.Printf("ActWorker %d Done\n", w.workerId)
	actionWg.Done()
}

type ActionDataProcessor struct {
	monData chan []*ActionData
	workers []*ActionWorker
	signal  chan os.Signal
}

func NewActionProcessor(maxWorker int, maxBuffer int) ActionDataProcessor {

	p := ActionDataProcessor{
		monData: make(chan []*ActionData),
		signal:  make(chan os.Signal, 1),
	}

	signal.Notify(p.signal, os.Interrupt)

	for i := 1; i <= maxWorker; i++ {
		p.workers = append(p.workers,
			&ActionWorker{
				workerId: i,
				actData:  p.monData,
				signals:  make(chan os.Signal, 1)})
	}

	return p
}

func (processor *ActionDataProcessor) AddData(data ...*ActionData) {

	processor.monData <- data
}

func (processor *ActionDataProcessor) Run() {

	for _, w := range processor.workers {
		actionWg.Add(1)
		go w.Process()
	}

	go func() {

		select {
		case sig := <-processor.signal:
			fmt.Println("Stopping ActWorkers")
			for _, w := range processor.workers {
				w.signals <- sig
			}
			return
		}
	}()

}

func (processor *ActionDataProcessor) Wait() {

	actionWg.Wait()
}
