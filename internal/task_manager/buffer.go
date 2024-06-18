package task_manager

import (
	pb "go-liteflow/pb"
	"sync"

	"google.golang.org/protobuf/proto"
)

// ever task need create a buffer space to cache input/output data
type Buffer struct {
	Mu       sync.Mutex
	Size     int //unit: byte
	TaskId   string
	Usage    int
	InQueue  chan *pb.Event            //serial obj data
	OutQueue map[string]chan *pb.Event //serial obj data(targetOpTaskId to output queue)

}

// data type enum
type DataType int

var InputData DataType = 1
var OutputData DataType = 2

// Buffer constructor
func NewBuffer(size int, taskId string) *Buffer {
	return &Buffer{
		Size:     size,
		TaskId:   taskId,
		InQueue:  make(chan *pb.Event, size),
		OutQueue: make(map[string]chan *pb.Event, 0),
	}
}

// set data to rel queue when has free space
func (b *Buffer) AddData(dataType DataType, data []*pb.Event) bool {
	switch dataType {
	case InputData:
		for _, d := range data {
			b.InQueue <- d
		}
	case OutputData:
		topId := data[0].TargetOpTaskId
		b.Mu.Lock()
		outQueue, exists := b.OutQueue[topId]
		if !exists {
			outQueue = make(chan *pb.Event, b.Size)
			b.OutQueue[topId] = outQueue
		}
		b.Mu.Unlock()
		for _, d := range data {
			outQueue <- d
		}
	default:
		return false
	}
	// when update usage need lock
	dataSize := 0
	for _, d := range data {
		dataSize += proto.Size(d)
	}
	b.Mu.Lock()
	defer b.Mu.Unlock()
	b.Usage += dataSize
	return true
}

// dequeue data from rel queue when the datas match the space and exec process success
func (b *Buffer) RemoveData(dataType DataType, size int, process func([]*pb.Event) error) error {
	//chose queue
	var queue *chan *pb.Event
	switch dataType {
	case InputData:
		queue = &b.InQueue
	case OutputData:
		d := b.OutQueue[b.TaskId]
		queue = &d
	default:
		return nil
	}
	//calculate remove data size
	curSize := 0
	data := make([]*pb.Event, 0)

loop:
	for {
		select {
		case d := <-*queue:
			dataSize := proto.Size(d)
			if curSize+dataSize >= size {
				break loop
			}
			curSize += dataSize
			data = append(data, d)
		default:
			// no data in chan, break loop
			break loop
		}
	}

	//push data，if has err then repush
	for {
		err := process(data)
		if err == nil {
			break
		}
	}
	//lock to update usage size
	b.Mu.Lock()
	defer b.Mu.Unlock()
	b.Usage -= curSize
	return nil
}
