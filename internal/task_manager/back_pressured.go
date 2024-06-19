package task_manager

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	pb "go-liteflow/pb"
	"log/slog"
	"reflect"
	"time"
)

// ever taskmanager has one monitor to manager buffer info
type TaskManagerBufferMonitor struct {
	bufferPool      map[string]*Buffer                     // taskId to buffer
	taskPool        map[string]*pb.OperatorTask            //taskId to operator
	eventChanClient map[string]*pb.Core_EventChannelClient //opId to client
}

func NewTaskManagerBufferMonitor() *TaskManagerBufferMonitor {
	tmbm := &TaskManagerBufferMonitor{
		bufferPool:      make(map[string]*Buffer),
		taskPool:        make(map[string]*pb.OperatorTask),
		eventChanClient: make(map[string]*pb.Core_EventChannelClient),
	}
	return tmbm
}

// register a task to task monitor
func (t *TaskManagerBufferMonitor) RegisterOperatorTask(task *pb.OperatorTask) error {
	if t.taskPool[task.Id] != nil {
		// has registered
		return errors.New("this task has registered")
	}
	t.taskPool[task.Id] = task
	// init buffer
	t.initialTaskBuffer(task.Id, 1024)
	// assign a thread to push data to rel downstreams when has downstream
	if len(task.Downstream) > 0 {
		go func() {
			// waiting connect to rel downstream server
			for {
				if task.State == pb.TaskStatus_Deployed{
					break
				}else {
					time.Sleep(5*time.Second)
				}
			}
			outputQueues := make(map[string]chan *pb.Event)
			// init downstream output buffer
			for _, opt := range task.Downstream {
				outputQueues[opt.Id] = make(chan *pb.Event, t.bufferPool[task.Id].Size)
			}
			t.bufferPool[task.Id].OutQueue = outputQueues

			cases := make([]reflect.SelectCase, 0, len(outputQueues))
			keys := make([]string, 0, len(outputQueues))
			for key, ch := range outputQueues {
				cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)})
				keys = append(keys, key)
			}
			for len(cases) > 0 {
				idx, recv, ok := reflect.Select(cases)
				if !ok {
					// channal has closed, remove rel channel
					cases = append(cases[:idx], cases[idx+1:]...)
					keys = append(keys[:idx], keys[idx+1:]...)
					continue
				}
				event := recv.Interface().(*pb.Event)
				slog.Info(fmt.Sprintf("Received from output channel %s: %v\n", keys[idx], event))
				//push data
				client := *t.eventChanClient[keys[idx]]
				for {
					er := client.Send(NewSingleEventReq(event.Data, task.Id, keys[idx]))
					if er != nil {
						slog.Error("data out push error")
					} else {
						resp, err := client.Recv()
						if err != nil || resp.EventType != pb.EventType_ACK || string(resp.Data) == "false" {
							slog.Error("data out push error")
						} else if string(resp.Data) == "true" {
							// push success
							slog.Info("data out push success")
							break
						}
					}
					time.Sleep(5 * time.Second)
				}

			}
		}()
	}
	return nil
}

// close a opertator task
func (t *TaskManagerBufferMonitor) CloseOperatorTask(taskId string) error {
	return t.releaseTaskBuffer(taskId)
}

// monitoring the buffer info by taskId
func (t *TaskManagerBufferMonitor) TaskBufferInfoMonitor(taskId string) *Buffer {
	return t.bufferPool[taskId]
}

// initial a buffer for a new task
func (t *TaskManagerBufferMonitor) initialTaskBuffer(taskId string, size int) {
	buffer := NewBuffer(size, taskId)
	t.bufferPool[taskId] = buffer
}

// release buffer (if has savepoint config, maybe need to save the cache buffer data?)
func (t *TaskManagerBufferMonitor) releaseTaskBuffer(taskId string) error {
	// release resources
	// pending .....

	// close client
	task := t.taskPool[taskId]
	for _, ds := range task.Downstream {
		c, e := t.eventChanClient[ds.Id]
		if e {
			(*c).CloseSend()
			delete(t.eventChanClient, ds.Id)
		}
	}
	// remove pool
	delete(t.bufferPool, taskId)
	delete(t.taskPool, taskId)
	return nil
}

func encodeByteData(data [][]byte) []byte {
	var buf bytes.Buffer
	for _, b := range data {
		binary.Write(&buf, binary.LittleEndian, uint32(len(b)))
		buf.Write(b)
	}
	encoded := buf.Bytes()

	return encoded

}

func decodeByteData(data []byte) [][]byte {
	var buf bytes.Buffer
	var decoded [][]byte
	buf = *bytes.NewBuffer(data)
	for buf.Len() > 0 {
		var length uint32
		binary.Read(&buf, binary.LittleEndian, &length)
		b := make([]byte, length)
		buf.Read(b)
		decoded = append(decoded, b)
	}
	return decoded
}
