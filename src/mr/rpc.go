package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type TaskType int

const (
	TASK_MAP    TaskType = 0
	TASK_REDUCE TaskType = 1
)

type TaskStatus int

const (
	TASK_STATUS_IN_PROGRESS TaskStatus = 0 // 10 sec, go routine, heart beat checking
	TASK_STATUS_COMPLETE    TaskStatus = 1
	TASK_STATUS_FAIL        TaskStatus = 2
)

type Task struct {
	TaskName          string
	TaskType          TaskType
	AllocatedFileName string
	OutputFileName    string
}

type RequestTaskArgs struct{}

type RequestTaskReply struct {
	Task Task
}

type UpdateTaskStatusArgs struct {
	Task       Task
	TaskStatus TaskStatus
}

type UpdateTaskStatusReply struct{}

// sample
type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
