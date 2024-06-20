package task_manager

import (
	"bytes"
	"encoding/binary"
	pb "go-liteflow/pb"
	"log/slog"
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

// close a opertator task
func (t *TaskManagerBufferMonitor) CloseOperatorTask(taskId string) error {
	slog.Info("release task resource........")

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

	buffer := t.bufferPool[taskId]
	for {
		// only has no data in buffer can close
		if buffer.Usage == 0 {
			// close client
			task := t.taskPool[taskId]
			for _, ds := range task.Downstream {
				c, e := t.eventChanClient[ds.Id]
				if e {
					(*c).CloseSend()
					delete(t.eventChanClient, ds.Id)
				}
			}
			// close downstream data queue channel
			bufferChannel := make([]chan *pb.Event, 0)
			for _, oq := range buffer.OutQueue {
				bufferChannel = append(bufferChannel, oq)
			}
			for _, c := range bufferChannel {
				close(c)
			}
			// remove pool
			delete(t.bufferPool, taskId)
			delete(t.taskPool, taskId)
			break
		}
		time.Sleep(5 * time.Second)
	}
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
