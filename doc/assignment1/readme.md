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