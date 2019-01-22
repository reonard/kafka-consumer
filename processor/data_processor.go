package processor

import (
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/patrickmn/go-cache"
	"kafka-consumer/db"
	"os"
	"os/signal"
	"strconv"
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

func (w *Worker) UpsertData(collection string, query map[string]interface{}, set interface{}, insert interface{}) {

	fmt.Printf("Worker %d update %s Data \n", w.workerId, collection)

	s := db.GetSession()
	defer s.Close()
	c := s.DB("pilot").C(collection)

	update := bson.M{"$set": set}

	if _, err := c.Upsert(query, update); err != nil {
		fmt.Println(err)
		return
	}

}

func (w *Worker) updateLastMonData(monData *MonitorData) {

	w.UpsertData("last_metric",
		map[string]interface{}{"deviceid": monData.DeviceId},
		monData, nil)

}

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

		var deviceInfo DeviceInfo
		var err error

		deviceCacheInfo, exists := db.InMemoryCache.Get(strconv.Itoa(int(dt.DeviceId)))
		if !exists {
			err, deviceInfo = GetDeviceInfo(dt.DeviceId)
			if err != nil {
				db.InMemoryCache.Set(strconv.Itoa(int(dt.DeviceId)), deviceInfo, cache.DefaultExpiration)
			}
		} else {
			deviceInfo = deviceCacheInfo.(DeviceInfo)
		}

		dt.Project = deviceInfo.Project
		dt.Customer = deviceInfo.Customer

		w.updateLastMonData(dt)
		w.saveMonData(dt)

		unixTime, err := strconv.Atoi(dt.TimeStamp)
		if err != nil {
			fmt.Println("Atoi错误")
			fmt.Println(unixTime)
			continue
		}
		// 更新设备状态
		err = db.ExecuteUpdate(
			"UPDATE device SET device_status = ?, status_time = FROM_UNIXTIME(?) WHERE device_id = ? and status_time < FROM_UNIXTIME(?)",
			dt.DeviceStatus, unixTime/1000, dt.DeviceId, unixTime/1000)

		if err != nil {
			fmt.Println(err)
		}
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

func NewProcessor(maxWorker int, maxBuffer int) MonitorDataProcessor {

	p := MonitorDataProcessor{
		monData: make(chan []*MonitorData),
		signal:  make(chan os.Signal, 1),
	}

	signal.Notify(p.signal, os.Interrupt)

	for i := 1; i <= maxWorker; i++ {
		p.workers = append(p.workers,
			&Worker{
				workerId:      i,
				maxDataBuffer: maxBuffer,
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
