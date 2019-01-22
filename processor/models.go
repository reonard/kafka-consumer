package processor

type MonitorData struct {
	DeviceId     int64                    `json:"ID"`
	Customer     string                   `json:"Customer"`
	Project      string                   `json:"Project"`
	DeviceStatus int8                     `json:"status"`
	TimeStamp    string                   `json:"once"`
	Data         []map[string]interface{} `json:"ProbeData"`
}

type DeviceInfo struct {
	DeviceId   int64
	DeviceName string
	Customer   string
	Project    string
}
