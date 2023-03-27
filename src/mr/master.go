package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

type MasterInterface interface {
	RequestTask(args *RequestTaskArgs, reply *RequestTaskReply) error
	UpdateTaskStatus(args *UpdateTaskStatusArgs, reply *UpdateTaskStatusReply) error
	Example(args *ExampleArgs, reply *ExampleReply) error
	Done() bool
}

type Master struct {
	masterStatus   MasterStatus
	mapTaskList    MasterTaskList
	reduceTaskList MasterTaskList
}

func (m *Master) checkTaskStatus(t *MasterTask) {
	select {
	case <-t.WaitChan:
		log.Println(fmt.Sprintf("taskType %d, taskID %d back to wait status", t.TaskType, t.TaskID))
		t.WaitChan <- struct{}{}
	case <-t.CompleteChan:
		log.Println(fmt.Sprintf("taskType %d, taskID %d is completed recently", t.TaskType, t.TaskID))
	case <-time.After(time.Second * time.Duration(10)):
		log.Println(fmt.Sprintf("taskType %d, taskID %d is not completed within 10 sec, let's rollback it", t.TaskType, t.TaskID))
		select {
		case <-t.RunningChan:
			t.WaitChan <- struct{}{}
		case <-time.After(time.Second * time.Duration(1)):
			log.Println("task is moved to wait or completed in the timeframe of 10 to 11sec")
		}
	}
}

func (m *Master) RequestTask(args *RequestTaskArgs, reply *RequestTaskReply) error {
	log.Println("[RequestTask API]")
	log.Println(fmt.Sprint("args", args))
	log.Println(fmt.Sprint("reply", reply))

	for idx, mapTask := range m.mapTaskList.TaskList {
		select {
		case <-mapTask.WaitChan:
			log.Println(fmt.Sprintf("dispatch %d map task to worker", idx))
			mapTask.RunningChan <- struct{}{}

			reply.Task = WorkerTask{
				TaskID:            mapTask.TaskID,
				TaskType:          mapTask.TaskType,
				AllocatedFileName: mapTask.AllocatedFileName,
			}
			reply.nReduce = len(m.reduceTaskList.TaskList)

			go m.checkTaskStatus(&mapTask)
		case <-time.After(time.Millisecond):
		}
	}
	for idx, reduceTask := range m.reduceTaskList.TaskList {
		select {
		case <-reduceTask.WaitChan:
			log.Println(fmt.Sprintf("dispatch %d reduce task to worker", idx))
			reduceTask.RunningChan <- struct{}{}

			reply.Task = WorkerTask{
				TaskID:            reduceTask.TaskID,
				TaskType:          reduceTask.TaskType,
				AllocatedFileName: reduceTask.AllocatedFileName,
			}
			reply.nReduce = len(m.reduceTaskList.TaskList)

			go m.checkTaskStatus(&reduceTask)
		case <-time.After(time.Millisecond):
		}
	}
	log.Println(fmt.Sprintf("currently "))
	return nil
}

func (m *Master) UpdateTaskStatus(args *UpdateTaskStatusArgs, reply *UpdateTaskStatusReply) error {
	log.Println("[UpdateTaskStatus API]")
	log.Println(fmt.Sprint("args", args))
	log.Println(fmt.Sprint("reply", reply))

	TaskType := args.Task.TaskType
	task := new(MasterTask)
	if TaskType == TASK_TYPE_REDUCE {
		task = &m.mapTaskList.TaskList[args.Task.TaskID]
	} else {
		task = &m.reduceTaskList.TaskList[args.Task.TaskID]
	}

	switch args.TaskStatus {
	case TASK_STATUS_IN_PROGRESS:
		select {
		case <-task.WaitChan:
			reply.msg = "heartbeat fail, task alr backed to wait status"
			task.WaitChan <- struct{}{}
		case <-task.RunningChan:
			reply.msg = "heartbeat successfully, task is running"
			task.RunningChan <- struct{}{}
		case <-task.CompleteChan:
			reply.msg = "heartbeat fail, task already completed"
			task.CompleteChan <- struct{}{}
		case <-time.After(time.Second & time.Duration(2)):
			reply.msg = "no available chan after 2 sec, nothing to do"
		}
	case TASK_STATUS_COMPLETE:
		select {
		case <-task.WaitChan:
			reply.msg = "complete fail, task alr backed to wait status"
			task.WaitChan <- struct{}{}
		case <-task.RunningChan:
			reply.msg = "complete successfully"
			// TODO: dirty job, update file
			task.CompleteChan <- struct{}{}
		case <-task.CompleteChan:
			reply.msg = "task already completed by another task, will not proceed these ouput files"
			task.CompleteChan <- struct{}{}
		case <-time.After(time.Second & time.Duration(2)):
			reply.msg = "no available chan after 2 sec, nothing to do"
		}
	case TASK_STATUS_FAIL:
		select {
		case <-task.WaitChan:
			reply.msg = "late fail inform, task alr backed to wait status"
			task.WaitChan <- struct{}{}
		case <-task.RunningChan:
			reply.msg = "got that inform, rollback task status to wait"
			task.WaitChan <- struct{}{}
		case <-task.CompleteChan:
			reply.msg = "late fail inform, task already completed by another task"
			task.CompleteChan <- struct{}{}
		case <-time.After(time.Second & time.Duration(2)):
			reply.msg = "no available chan after 2 sec, nothing to do"
		}
	}
	log.Println(reply.msg)
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
// if the entire job ha s finished.
func (m *Master) Done() bool {
	if m.masterStatus.mu.TryLock() {
		defer m.masterStatus.mu.Unlock()
		if m.masterStatus.TaskStatus == TASK_STATUS_COMPLETE {
			log.Println("task completed")
			return true
		}
	}
	log.Println("task haven't completed")
	return false
}

// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeMaster(filenames []string, nReduce int) *Master {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	m := Master{
		masterStatus: MasterStatus{
			TaskStatus: TASK_STATUS_IN_PROGRESS,
		},
		mapTaskList: MasterTaskList{
			TaskList: make([]MasterTask, 0),
		},
		reduceTaskList: MasterTaskList{
			TaskList: make([]MasterTask, 0),
		},
	}

	for idx, filename := range filenames {
		newMapTask := MasterTask{
			TaskID:            idx,
			TaskType:          TASK_TYPE_MAP,
			AllocatedFileName: []string{filename},
			CompleteChan:      make(chan struct{}, 1),
			RunningChan:       make(chan struct{}, 1),
			WaitChan:          make(chan struct{}, 1),
		}
		newMapTask.WaitChan <- struct{}{}
		m.mapTaskList.TaskList = append(m.mapTaskList.TaskList, newMapTask)
	}

	for idx := 0; idx < nReduce; idx++ {
		reduceTaskFilenames := make([]string, 0)
		for x := 0; x < len(filenames); x++ {
			reduceTaskFilenames = append(
				reduceTaskFilenames,
				GetMasterTaskOuputFilenames(TASK_TYPE_REDUCE, x, idx),
			)
		}

		newReduceTask := MasterTask{
			TaskID:            idx,
			TaskType:          TASK_TYPE_REDUCE,
			AllocatedFileName: reduceTaskFilenames,
			CompleteChan:      make(chan struct{}, 1),
			RunningChan:       make(chan struct{}, 1),
			WaitChan:          make(chan struct{}, 1),
		}
		newReduceTask.WaitChan <- struct{}{}
		m.reduceTaskList.TaskList = append(m.reduceTaskList.TaskList, newReduceTask)
	}

	m.server()
	return &m
}

// server start
// worker request task
// ... woker and server will take tens of sec to complete the task
// worker and server keep the heartbeat check every 10 sec ===> process>10 sec? update the xxx
// worker reply the task status, master confirm
