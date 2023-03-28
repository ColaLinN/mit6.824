package mr

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"
)

type WorkerInterface interface {
	CallRequestTask(args *RequestTaskArgs) *RequestTaskReply
	CallUpdateTaskStatus(args *UpdateTaskStatusArgs) *UpdateTaskStatusReply
	CallExample()
}

type WorkerImpl struct {
	mapf    func(string, string) []KeyValue
	reducef func(string, []string) string
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(
	mapf func(string, string) []KeyValue,
	reducef func(string, []string) string,
) {
	// uncomment to send the Example RPC to the master.
	// CallExample()

	w := WorkerImpl{
		mapf:    mapf,
		reducef: reducef,
	}

	for {
		taskReply, err := w.CallRequestTask()
		if err != nil {
			log.Println(err.Error())
			return
		}
		task := &taskReply.Task
		allocatedFilenames := task.AllocatedFileName

		switch taskReply.Task.TaskType {
		case TASK_TYPE_MAP:
			outputFilenames, err := w.Map(allocatedFilenames, task.TaskID, taskReply.NReduce)
			if err != nil {
				updateReply, err := w.CallUpdateTaskStatus(task, TASK_STATUS_FAIL, nil)
				if err != nil {
					log.Println(err.Error())
					return
				}
				log.Println(updateReply)
			} else {
				updateReply, err := w.CallUpdateTaskStatus(task, TASK_STATUS_COMPLETE, outputFilenames)
				if err != nil {
					log.Println(err.Error())
					return
				}
				log.Println(updateReply)
			}
		case TASK_TYPE_REDUCE:
			outputFilenames, err := w.Reduce(allocatedFilenames, task.TaskID)
			if err != nil {
				updateReply, err := w.CallUpdateTaskStatus(task, TASK_STATUS_FAIL, nil)
				if err != nil {
					log.Println(err.Error())
					return
				}
				log.Println(updateReply)
			} else {
				updateReply, err := w.CallUpdateTaskStatus(task, TASK_STATUS_COMPLETE, outputFilenames)
				if err != nil {
					log.Println(err.Error())
					return
				}
				log.Println(updateReply)
			}
		case TASK_TYPE_EXIT:
			return
		case TASK_TYPE_WAIT:
			log.Println("wait for a while")
			time.Sleep(time.Second * time.Duration(2))
		}
	}
}

func (w *WorkerImpl) CallRequestTask() (*RequestTaskReply, error) {
	args := RequestTaskArgs{}
	reply := RequestTaskReply{}

	if ok := call("Master.RequestTask", &args, &reply); !ok {
		return nil, errors.New("request fail")
	}

	log.Println("TaskType", TaskType_NameMap[reply.Task.TaskType])
	log.Println("TaskID", reply.Task.TaskID)
	return &reply, nil
}

// TODO: heartbeat
func (w *WorkerImpl) CallUpdateTaskStatus(task *WorkerTask, status TaskStatus, ouputFilenames []string) (*UpdateTaskStatusReply, error) {
	args := UpdateTaskStatusArgs{
		Task:               *task,
		TaskStatus:         status,
		OutputTempFileName: ouputFilenames,
	}
	reply := UpdateTaskStatusReply{}

	log.Println("TaskStatus", TaskStatus_NameMap[args.TaskStatus])

	if ok := call("Master.UpdateTaskStatus", &args, &reply); !ok {
		return nil, errors.New("request fail")
	}

	return &reply, nil
}

func (w *WorkerImpl) Map(filenames []string, mapID int, nReduce int) ([]string, error) {
	// read each input file,
	// pass it to Map,
	// accumulate the intermediate Map output.
	//
	intermediateKVArrays := make([][]KeyValue, 10)
	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("cannot open %v", filename)
		}
		content, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatalf("cannot read %v", filename)
		}
		file.Close()
		keyValueArray := w.mapf(filename, string(content))
		for _, kv := range keyValueArray {
			reduceID := ihash(kv.Key) % nReduce
			intermediateKVArrays[reduceID] = append(intermediateKVArrays[reduceID], kv)
		}
	}

	for reduceID := 0; reduceID < nReduce; reduceID++ {
		tempfilename := fmt.Sprintf("temp.mr-%d-%d.*", mapID, reduceID)
		tmpfile, err := ioutil.TempFile("", tempfilename)
		if err != nil {
			return []string{}, err
		}
		defer tmpfile.Close()

		enc := json.NewEncoder(tmpfile)
		for _, kv := range intermediateKVArrays[reduceID] {
			err := enc.Encode(&kv)
			if err != nil {
				return []string{}, err
			}
		}

		outputFileName := fmt.Sprintf("mr-%d-%d", mapID, reduceID)
		if err := os.Rename(tmpfile.Name(), outputFileName); err != nil {
			return []string{}, err
		}
	}

	return []string{}, nil
}

func (w *WorkerImpl) Reduce(filenames []string, reduceID int) ([]string, error) {
	intermediateKVArray := []KeyValue{}
	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("cannot open %v", filename)
		}
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			intermediateKVArray = append(intermediateKVArray, kv)
		}
	}

	sort.Sort(ByKey(intermediateKVArray))

	tempfilename := fmt.Sprintf("temp.mr-out-%d.*", reduceID)
	tmpfile, err := ioutil.TempFile("", tempfilename)
	if err != nil {
		return []string{}, err
	}
	defer tmpfile.Close()

	//
	// call Reduce on each distinct key in intermediate[],
	// and print the result to mr-out-0.
	//
	i := 0
	for i < len(intermediateKVArray) {
		j := i + 1
		for j < len(intermediateKVArray) && intermediateKVArray[j].Key == intermediateKVArray[i].Key {
			j++
		}

		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediateKVArray[k].Value)
		}
		output := w.reducef(intermediateKVArray[i].Key, values)

		fmt.Fprintf(tmpfile, "%v %v\n", intermediateKVArray[i].Key, output)

		i = j
	}

	outputFileName := fmt.Sprintf("mr-out-%d", reduceID)
	if err := os.Rename(tmpfile.Name(), outputFileName); err != nil {
		return []string{}, err
	}
	return []string{outputFileName}, nil
}

// example function to show how to make an RPC call to the master.
//
// the RPC argument and reply types are defined in rpc.go.
func (w *WorkerImpl) CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	call("Master.Example", &args, &reply)

	// reply.Y should be 100.
	fmt.Printf("reply.Y %v\n", reply.Y)
}

// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	log.Println("call API", rpcname, "args", args)
	err = c.Call(rpcname, args, reply)
	if err != nil {
		fmt.Println(err)
		return false
	}

	log.Println("call API", rpcname, "reply", reply)
	return true
}
