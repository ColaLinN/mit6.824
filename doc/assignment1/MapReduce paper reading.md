

意思是Reduce Woker读到他被assign的所有文件后，他会像mrsequential.go里一样，把所有的文件读到内存里，按key(word)排序，然后遍历计算每个word的count

When a reduce worker has read all intermediate data, it sorts it by the

The sorting is needed because typically many different keys map to the same reduce task