package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"


// example to show how to declare the arguments
// and reply for an RPC.
//
// sample
type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.
type RequestTaskArgs struct{}

type RequestTaskReply struct {
	Task    WorkerTask
	nReduce int
}

type UpdateTaskStatusArgs struct {
	Task               WorkerTask
	TaskStatus         TaskStatus
	OutputTempFileName []string // TODO: how should we name the tempFile?
}

type UpdateTaskStatusReply struct {
	msg string
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
