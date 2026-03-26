package delayqueue

type Task struct {
	TaskID     string `json:"task_id"`
	BizType    string `json:"biz_type"`
	Payload    []byte `json:"payload"`
	ExecuteAt  int64  `json:"execute_at"`
	RetryCount int    `json:"retry_count"`
	CreatedAt  int64  `json:"created_at"`
}
