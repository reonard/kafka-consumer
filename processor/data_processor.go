package processor

import (
	"fmt"
	"kafka-consumer/db"
	"os"
	"os/signal"
	"sync"
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

func (w *Worker) flushData() {

	fmt.Printf("Worker %d Flushing Data \n", w.workerId)
	s := db.GetSession()
	defer s.Close()
	c := s.DB("pilot").C("device1")

	if len(w.bulkData) == 0 {
		fmt.Println("No data need to flushed")
		return
	}

	if err := c.Insert(w.bulkData...); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf(" %d Data Flushed \n", len(w.bulkData))
}

func (w *Worker) saveData(data ...*MonitorData) {

	for _, dt := range data {
		w.bulkData = append(w.bulkData, dt)
	}

	if len(w.bulkData) >= w.maxDataBuffer {
		w.flushData()
		w.bulkData = make([]interface{}, 0, w.maxDataBuffer)
	}
}

func (w *Worker) Process() {

	fmt.Printf("Worker %d start processing \n", w.workerId)

Loop:
	for {
		select {
		case dt, ok := <-w.monData:
			if ok {
				w.saveData(dt...)
			}
		case <-w.signals:
			fmt.Printf("Worker %d Received Signal, Flushing...\n", w.workerId)
			w.flushData()
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
