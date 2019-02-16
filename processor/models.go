package processor

type MonitorData struct {
	DeviceId     int64                    `json:"ID"`
	Project      int64                    `json:"Project"`
	DeviceStatus int8                     `json:"status"`
	TimeStamp    string                   `json:"once"`
	Data         []map[string]interface{} `json:"ProbeData"`
}

type ActionData struct {
	DeviceId     int64                    `json:"ID"`
	MsgID        int64                    `json:"MsgID"`
	TimeStamp    string                   `json:"once"`
	Success      bool                     `json:"Success"`
	DeviceSecret string                   `json:"Sign,omitempty"`
	ProbeConfig  []map[string]interface{} `json:"ProbeConfig"`
}

type DeviceInfo struct {
	DeviceId   int64
	DeviceName string
	Project    int64
}
