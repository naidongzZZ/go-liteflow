package task_manager

import (
	"errors"
	"math"
)

// // DownStream send credit to Upstream when has enough buffer space
// type Notify interface {
// 	Notify(credit int) error
// }

// // Upstream receive credit from Downstream and send data to Downstream
// // Downstream receive data from Upstream and add data to queue
// type Response interface {
// 	ResponseDownstreamData(credit int,taskId string) error
// 	ReceiveUpstreamData(data [][]byte,taskId string) error
// }

// ever taskmanager has one monitor to manager buffer info
type TaskManagerBufferMonitor struct {
	bufferPool map[string]*Buffer 	// taskId to buffer
}

// monitoring the buffer info by taskId
func (t *TaskManagerBufferMonitor) JobBufferInfoMonitor(taskId string) *Buffer{
	return t.bufferPool[taskId]
}


// initial a buffer for a new task
func (t *TaskManagerBufferMonitor) InitialJobBuffer(taskId string) {
	size := 1024*1024*10; // pending......
	buffer := NewBuffer(size,taskId)
	t.bufferPool[taskId] = buffer;
	// default send 30% size credit to upstream
	credit := int(math.Round(float64(buffer.Size) * 0.3))
	t.Notify(credit,buffer.TaskId)
}


// release buffer (if has savepoint config, maybe need to save the buffer info?)
func (t *TaskManagerBufferMonitor) ReleaseJobBuffer(taskId string) error{
	delete(t.bufferPool,taskId)
	return nil
}


// push data to downstream
func (t *TaskManagerBufferMonitor) ResponseDownstream(credit int,taskId string) error{
	buffer := t.bufferPool[taskId]
	if buffer == nil {
		return errors.New("task info is not exist")
	}
	// data := buffer.RemoveData(OutputData,credit)
	// get downstream channel by taskId
	// channel := xxxx;
	// if push error , add data to buffer
	// for _,d := range data {
	// 	buffer.AddData(OutputData,d)
	// }
	return nil
}

func (t *TaskManagerBufferMonitor) ReceiveUpstreamData(data [][]byte,taskId string) error{
	buffer := t.bufferPool[taskId]
	if buffer == nil || len(data) == 0 {
		return errors.New("task info is not exist or has no input data")
	}

	result := buffer.AddData(InputData,data)
	
	if result{
		return errors.New("input data import errors")
	}

	return nil
}


func (t *TaskManagerBufferMonitor) Notify(credit int,taskId string) error{
	// taskId := t.TaskId
	// get upstream channel by taskId
	// channel := xxxx;
	return nil;
}
