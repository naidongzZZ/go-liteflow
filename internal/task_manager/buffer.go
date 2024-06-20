package task_manager

import (
	pb "go-liteflow/pb"
	"log/slog"
	"reflect"
	"sync"
	"time"

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
		var outQueue chan *pb.Event
		var exist bool
		for {
			outQueue, exist = b.OutQueue[topId]
			if exist {
				break
			} else {
				time.Sleep(5 * time.Second)
			}
		}

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

func (b *Buffer) UpdateUsageWhenOutput(size int) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	b.Usage -= size
	slog.Info("buffer state", slog.Any("usage", b.Usage))
}

func (b *Buffer) ListenInputQueueAndHandle(process func(*pb.Event) error) error {
	queue := &b.InQueue
	for data := range *queue {
		for {
			err := process(data)
			if err == nil {
				//handle success
				b.UpdateUsageWhenOutput(proto.Size(data))
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}

	return nil
}

func (b *Buffer) LinstenOutputQueueAndPush(clientMap map[string]*pb.Core_EventChannelClient) error {
	outputQueues := b.OutQueue
	cases := make([]reflect.SelectCase, 0, len(outputQueues))
	keys := make([]string, 0, len(outputQueues))
	for key, ch := range outputQueues {
		cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)})
		keys = append(keys, key)
	}
	for len(cases) > 0 {
		// listen output queue data
		idx, recv, ok := reflect.Select(cases)
		if !ok {
			// channal has closed, remove rel channel
			cases = append(cases[:idx], cases[idx+1:]...)
			keys = append(keys[:idx], keys[idx+1:]...)
			continue
		}
		event := recv.Interface().(*pb.Event)
		// slog.Info(fmt.Sprintf("Received from output channel %s: %v\n", keys[idx], event))
		//push data
		client := *clientMap[keys[idx]]
		for {
			er := client.Send(NewSingleEventReq(event.Data, b.TaskId, keys[idx]))
			if er != nil {
				slog.Error("data out push error")
			} else {
				resp, err := client.Recv()
				if err != nil || resp.EventType != pb.EventType_ACK || string(resp.Data) == "false" {
					slog.Error("data out push error")
				} else if string(resp.Data) == "true" {
					// push success
					slog.Info("<===============data out push success", slog.Any("event", event))
					go b.UpdateUsageWhenOutput(proto.Size(event))
					break
				}
			}
			time.Sleep(5 * time.Second)
		}

	}
	return nil
}
