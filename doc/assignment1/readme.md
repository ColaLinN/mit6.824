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


Task: complete the master and the workers
1. One master and multiple workers: 
   - One master process, and one or more worker processes executing in parallel, run them all on a single machine. 
2. The **workers** will talk to the master via RPC. 
3. Each worker process will 
   1. ask the master for a task, 
   2. read the task's input from one or more files, 
   3. execute the task, 
   4. and write the task's output to one or more files. 
4. Fault detection: The master should notice if a worker hasn't completed its task in a reasonable amount of time (for this lab, use ten seconds), and give the same task to a different worker.


DON'T CHANGE: The "main" routines for the master and worker are in 
- main/mrmaster.go 
- and main/mrworker.go; 
  
CHANGE: You should put your implementation in 
- mr/master.go, 
- mr/worker.go, 
- and mr/rpc.go.

running commands
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

Exception Error case:
1. network error:
- 2023/03/23 03:04:52 dialing:dial unix /var/tmp/824-mr-501: connect: connection refused
exit status 1
2. go plugin error
- 2023/03/23 03:04:37 cannot load plugin wc.so