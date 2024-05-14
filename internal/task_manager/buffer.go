package task_manager

import (
	"sync"
)

// ever task need create a buffer space to cache input/output data
type Buffer struct {
	Mu       sync.Mutex
	Size     int //unit: byte
	TaskId    string
	Usage    int
	InQueue  [][]byte //serial obj data
	OutQueue [][]byte //serial obj data
}

// data type enum
type DataType int;
var InputData DataType = 1;
var OutputData DataType = 2;

// Buffer constructor
func NewBuffer(size int, taskId string) *Buffer {
	return &Buffer{
		Size:     size,
		TaskId:    taskId,
		InQueue:  make([][]byte, 0),
		OutQueue: make([][]byte, 0),
	}
}


// set data to rel queue when has free space
func (b *Buffer) AddData(dataType DataType, data [][]byte) bool {
    b.Mu.Lock()
    defer b.Mu.Unlock()
	dataSize := 0
    for _,d := range data {
        dataSize += len(d)
    }
    if b.Usage+dataSize > b.Size { 
        return false
    }
    switch dataType {
    case InputData:
        b.InQueue = append(b.InQueue, data...)
    case OutputData:
        b.OutQueue = append(b.OutQueue, data...)
    default:
        return false
    }
    b.Usage += dataSize
    return true
}

// dequeue data from rel queue when the datas match the space and exec process success
func (b *Buffer) RemoveData(dataType DataType, size int,process func([][]byte) error ) error {
    b.Mu.Lock()
    defer b.Mu.Unlock()

	//chose queue
    var queue *[][]byte
    switch dataType {
    case InputData:
        queue = &b.InQueue
    case OutputData:
        queue = &b.OutQueue
    default:
        return nil
    }
	//calculate remove data size
	curSize := 0
    var removeCount int
    data := make([][]byte, 0)
    for _, d := range *queue {
        if curSize+len(d) > size {
            break
        }
        data = append(data, d)
        curSize += len(d)
        removeCount++
    }
    err := process(data)
    if err != nil {
        return err
    }
    *queue = (*queue)[removeCount:]
    b.Usage -= curSize
    return nil
}
