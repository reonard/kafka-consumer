package processor

type MonitorData struct {
	DeviceId     int64                    `json:"ID"`
	Project      int64                    `json:"Project"`
	DeviceStatus int8                     `json:"status"`
	TimeStamp    string                   `json:"once"`
	Data         []map[string]interface{} `json:"ProbeData"`
}

type DeviceInfo struct {
	DeviceId   int64
	DeviceName string
	Project    int64
}
