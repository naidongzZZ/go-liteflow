package task_manager

import (
	"bytes"
	"encoding/binary"
	"errors"
	pb "go-liteflow/pb"
	"math/rand"
	"time"
)

// ever taskmanager has one monitor to manager buffer info
type TaskManagerBufferMonitor struct {
	bufferPool      map[string]*Buffer                    // taskId to buffer
	taskPool        map[string]*pb.OperatorTask           //taskId to operator
	notifyChan      chan []any                            // index-0:opId,index-1:currentTaskId,index-2:targetTaskId,index-3:credit
	eventChanClient map[string]pb.Core_EventChannelClient //opId to client
}

func NewTaskManagerBufferMonitor() *TaskManagerBufferMonitor {
	return &TaskManagerBufferMonitor{
		bufferPool:      make(map[string]*Buffer),
		taskPool:        make(map[string]*pb.OperatorTask),
		notifyChan:      make(chan []any),
		eventChanClient: make(map[string]pb.Core_EventChannelClient),
	}
}

// register a task to task monitor
func (t *TaskManagerBufferMonitor) RegisterOperatorTask(task *pb.OperatorTask) error {
	if t.taskPool[task.Id] != nil {
		// has registered
		return errors.New("this task has registered")
	}
	t.taskPool[task.Id] = task
	// init buffer
	t.initialTaskBuffer(task.Id)
	// assign a thread to manager this task
	go func(taskId string) {
		for {
			if t.taskPool[taskId] == nil || t.bufferPool[taskId] == nil {
				// task is over
				break
			}
			buffer := t.bufferPool[taskId]
			if buffer.Size-buffer.Usage > 1024 {
				upstream := task.Upstream[rand.Intn(len(task.Upstream))]
				t.notifyChan <- []any{upstream.OpId, taskId, upstream.Id, 1024}
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}(task.Id)
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
func (t *TaskManagerBufferMonitor) initialTaskBuffer(taskId string) {
	size := 1024 * 1024 * 10 // pending......
	buffer := NewBuffer(size, taskId)
	t.bufferPool[taskId] = buffer
	// default send 30% size credit to upstream
	// credit := int(math.Round(float64(buffer.Size) * 0.3))
	// t.Notify(credit, buffer.TaskId)
}

// release buffer (if has savepoint config, maybe need to save the cache buffer data?)
func (t *TaskManagerBufferMonitor) releaseTaskBuffer(taskId string) error {
	// release resources
	// pending .....

	// remove pool
	delete(t.bufferPool, taskId)
	delete(t.taskPool, taskId)
	delete(t.eventChanClient, taskId)
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
