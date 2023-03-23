# mit6.824

# assignment 1

```
export GO111MODULE="off"

$ cd ~/6.824
$ cd src/main
$ go build -buildmode=plugin ../mrapps/wc.go
$ rm mr-out*
$ go run mrsequential.go wc.so pg*.txt
$ more mr-out-0
A 509
ABOUT 2
ACT 8
...

```

file under /src/main/

1. mrsequential.go
   - maps and reduces one at a time
2. mrapps/wc.go
   - Word Count
3. mrapps/indexer.go
   - text indexer

### Task: complete the master and the workers

1. One master and multiple workers: 
   - One master process, and one or more worker processes executing in parallel, run them all on a single machine. 
2. The **workers** will talk to the master via RPC. 
3. Each worker process will 
   1. ask the master for a task, 
   2. read the task's input from one or more files, 
   3. execute the task, 
   4. and write the task's output to one or more files. 
4. Fault detection: The master should notice if a worker hasn't completed its task in a reasonable amount of time (for this lab, use ten seconds), and give the same task to a different worker.

### Extra Rules

1. Divide intermediate keys
- The map phase should divide the intermediate keys into buckets for nReduce reduce tasks
2. X'th Reduce task -> `mr-out-X`. 
- The worker implementation should put the output of the X'th reduce task in the file mr-out-X
3. `"%v %v"` output format for the return amount
- A mr-out-X file should contain one line per Reduce function output.
4. only modify these files: mr/worker.go, mr/master.go, and mr/rpc.go
5. put intermediate Map output in files in the current directory
6. implement Done() to exit
7. When the job is completely finished, the worker processes should exit.

### Hints

1. [key point] intermediate file name: `mr-X-Y`, X Map tasks, and Y Reduce tasks
2. use encoding/json to encode the keyValue data
3. [key point] use ihash(key) % NReduce to choose the reduce task id
4. [key point] [atom read] master, as a server, will be concurrent, don't forget to lock shared data.
5. Use Go's race detector, with go build -race and go run -race. Can refer to ./test-mr.sh
6. [key point] Workers will sometimes need to wait, e.g. reduces can't start until the last map has finished
7. To test crash recovery, you can use the mrapps/crash.go application plugin.
8. Master cannot distingusish between crashed worker and slowly alive wokrers. 
- The best you can do is have the master wait for some amount of time, and then give up and re-issue the task to a different worker.
8. use ioutil.TempFile to create a temporary file and os.Rename to atomically rename it once it is completely written.
9. the outputs of test-mr.sh are in sub-directory `mr-tmp`


DON'T CHANGE: The "main" routines for the master and worker are in 
- main/mrmaster.go 
- and main/mrworker.go; 
  

CHANGE: You should put your implementation in 
- mr/master.go, 
- mr/worker.go, 
- and mr/rpc.go.

### Exception Error case:

1. network error:

- 2023/03/23 03:04:52 dialing:dial unix /var/tmp/824-mr-501: connect: connection refused
  exit status 1

2. go plugin error

- 2023/03/23 03:04:37 cannot load plugin wc.so
- Root Cause: wc.so is using the keyValue struct in ./mr/rpc.go. 
- Therefore, if you change anything in the mr/ directory, you will probably have to re-build any MapReduce plugins you use, with something like go build -buildmode=plugin ../mrapps/wc.go

### Commands

```
rm mr-out-*
go run mrmaster.go pg-*.txt
go run mrworker.go wc.so
```

the result should match this
```
cat mr-out-* | sort | more
```

test script
1. check the output
2. check master and workers run in parrallel
3. recovers from workers that crashed while running task
4. expects to see output in files named mr-out-X, one for each reduce task. 
```
sh ./test-mr.sh
```



# file name convention

intermediate files will be: mr-X-Y

X is the number of Map tasks, Y is the number of Reduce tasks.

Let's say 2 map tasks, and 4 reduce tasks



Y is getting by ihash(key) % NReduce 

For map task 1

1. mr-0-0
2. mr-0-1
3. mr-0-2
4. mr-0-3

For map task 2

1. mr-1-0
2. mr-1-1
3. mr-1-2
4. mr-1-3



For reduce taks 0, worker will read the following file 

1. mr-0-0
2. mr-1-0

ouput is `mr-out-0`.



For reduce taks 1, worker will read the following file 

1. mr-0-1
2. mr-1-1

ouput is `mr-out-1`.
