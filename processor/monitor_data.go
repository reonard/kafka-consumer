package processor

import "time"


type MonitorData struct {
	Id string
	TimeStamp time.Time
	Data map[string]interface{}
}


func Save(){

}