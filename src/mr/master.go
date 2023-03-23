package mr

import "log"
import "net"
import "os"
import "net/rpc"
import "net/http"

type MasterInterface interface {
	// TODO 
	RequestTask(args *RequestTaskArgs, reply *RequestTaskReply) error
	UpdateTaskStatus(args *UpdateTaskStatusArgs, reply *UpdateTaskStatusReply) error
	Example(args *ExampleArgs, reply *ExampleReply) error
}

type Master struct {
	TotalTaskStatus  TaskStatus
	MapTaskList      []Task
	ReduceTaskList   []Task
	CompleteTaskList []Task
	// TODO 
	// Your definitions here.
}

// Your code here -- RPC handlers for the worker to call.

func (m *Master) RequestTask(args *RequestTaskArgs, reply *RequestTaskReply) error {
	// TODO 
	return nil
}

func (m *Master) UpdateTaskStatus(args *UpdateTaskStatusArgs, reply *UpdateTaskStatusReply) error {
	// TODO 
	return nil
}

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (m *Master) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// start a thread that listens for RPCs from worker.go
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
func (m *Master) Done() bool {
	// TODO key point atom read?
	if m.TotalTaskStatus == TASK_STATUS_COMPLETE {
		return true
	}
	// TODO Your code here.
	return false
}

// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{
		TotalTaskStatus: TASK_STATUS_IN_PROGRESS,
	}

	// Your code here.
	// TODO go routine
	// TODO dispatchTask

	m.server()
	return &m
}

// server start
// worker request task
// ... woker and server will take tens of sec to complete the task
// worker and server keep the heartbeat check every 10 sec ===> process>10 sec? update the xxx
// worker reply the task status, master confirm
// 