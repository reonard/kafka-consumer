package processor

import (
	"fmt"
	"kafka-consumer/db"
	"os"
	"os/signal"
	"sync"
)

const (
	STATUS_NORMAL = iota
	STATUS_WARN
	STATUS_ALARM
	STATUS_OFFLINE
	STATUS_ERROR
)

var wg sync.WaitGroup

type Worker struct {
	workerId      int
	maxDataBuffer int
	bulkData      []interface{}
	signals       chan os.Signal
	monData       chan []*MonitorData
	error         chan error
}

func (w *Worker) flushData(collection string, data ...interface{}) {

	fmt.Printf("Worker %d Flushing %s Data \n", w.workerId, collection)

	s := db.GetSession()
	defer s.Close()
	c := s.DB("pilot").C(collection)

	if err := c.Insert(data...); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf(" %d %s Data Flushed \n", len(data), collection)

}

//func (w *Worker) saveAlarmData(monData *MonitorData) {
//
//	w.flushData("alarm_data", monData)
//}

func (w *Worker) saveMonData(monData *MonitorData) {

	w.bulkData = append(w.bulkData, monData)

	// flush bulk data
	if len(w.bulkData) >= w.maxDataBuffer {

		w.flushData("metric_data", w.bulkData...)

		w.bulkData = make([]interface{}, 0, w.maxDataBuffer)
	}

}

func (w *Worker) processData(data ...*MonitorData) {

	for _, dt := range data {

		w.saveMonData(dt)

		//if dt.DeviceStatus == STATUS_ALARM {
		//	w.saveAlarmData(dt)
		//}
	}
}

func (w *Worker) Process() {

	fmt.Printf("Worker %d start processing \n", w.workerId)

Loop:
	for {
		select {

		case dt, ok := <-w.monData:
			if ok {
				w.processData(dt...)
			}
		case <-w.signals:
			fmt.Printf("Worker %d Received Signal, Flushing...\n", w.workerId)
			w.flushData("metric_data", w.bulkData...)
			break Loop
		}
	}

	fmt.Printf("Worker %d Done\n", w.workerId)
	wg.Done()
}

type MonitorDataProcessor struct {
	monData chan []*MonitorData
	workers []*Worker
	signal  chan os.Signal
}

func NewProcessor(maxWorker int) MonitorDataProcessor {

	p := MonitorDataProcessor{
		monData: make(chan []*MonitorData),
		signal:  make(chan os.Signal, 1),
	}

	signal.Notify(p.signal, os.Interrupt)

	for i := 1; i <= maxWorker; i++ {
		p.workers = append(p.workers,
			&Worker{
				workerId:      i,
				maxDataBuffer: 1,
				monData:       p.monData,
				signals:       make(chan os.Signal, 1)})
	}

	return p
}

func (processor *MonitorDataProcessor) AddData(data ...*MonitorData) {

	processor.monData <- data
}

func (processor *MonitorDataProcessor) Run() {

	for _, w := range processor.workers {
		wg.Add(1)
		go w.Process()
	}

	go func() {

		select {
		case sig := <-processor.signal:
			fmt.Println("Stopping Workers")
			for _, w := range processor.workers {
				w.signals <- sig
			}
			return
		}
	}()

}

func (processor *MonitorDataProcessor) Wait() {

	wg.Wait()
}
