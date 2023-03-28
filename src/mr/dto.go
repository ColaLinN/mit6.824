package mr

import (
	"fmt"
	"sync"
)

func GetMasterTaskOuputFilenames(taskType TaskType, mapTaskID int, reduceTaskID int) string {
	switch taskType {
	case TASK_TYPE_MAP:
		return fmt.Sprintf("mr-%d-%d", mapTaskID, reduceTaskID)
	case TASK_TYPE_REDUCE:
		return fmt.Sprintf("mr-out-%d", reduceTaskID)
	default:
		return ""
	}
}

type TaskType int

const (
	TASK_TYPE_EXIT TaskType = iota // used to inform worker only, either return this while all the tasks done, or can say the task is done while fail to contact master
	TASK_TYPE_MAP
	TASK_TYPE_REDUCE
	TASK_TYPE_WAIT // used to inform worker only, let worker be idle for a while
)

var TaskType_NameMap = map[TaskType]string{
	TASK_TYPE_EXIT:   "EXIT",
	TASK_TYPE_MAP:    "MAP",
	TASK_TYPE_REDUCE: "REDUCE",
	TASK_TYPE_WAIT:   "WAIT",
}

type TaskStatus int

const (
	TASK_STATUS_IN_PROGRESS TaskStatus = iota // will not be used, 10 sec, go routine, heart beat checking
	TASK_STATUS_COMPLETE
	TASK_STATUS_FAIL
)

var TaskStatus_NameMap = map[TaskStatus]string{
	TASK_STATUS_IN_PROGRESS: "IN_PROGRESS",
	TASK_STATUS_COMPLETE:    "COMPLETE",
	TASK_STATUS_FAIL:        "FAIL",
}

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

type WorkerTask struct {
	TaskID            int
	TaskType          TaskType
	AllocatedFileName []string
}

type MasterTask struct {
	TaskID            int
	TaskType          TaskType
	AllocatedFileName []string
	CompleteChan      chan struct{}
	RunningChan       chan struct{}
	WaitChan          chan struct{}
}

type MasterTaskList struct {
	TaskList []MasterTask
}

type MasterStatus struct {
	CompletedMapTaskNumber    int
	CompletedReduceTaskNumber int
	TaskStatus                TaskStatus
	mu                        sync.Mutex
}
