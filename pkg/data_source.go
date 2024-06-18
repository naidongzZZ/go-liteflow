package pkg

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	pb "go-liteflow/pb"
	"log/slog"

	kafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

type DataStream struct {
	Env         uint
	Parallelism uint
	Source      *DataSource
	Operations  []Operation
	OperaPara   map[*Operation]uint
}

type DataSource struct {
	Data   chan *pb.Event
	Output []any
}

func NewDataStream(source *DataSource) *DataStream {
	return &DataStream{
		Source:      source,
		Operations:  make([]Operation, 0),
		Parallelism: 1,
	}
}

func FromKafka(brokers, groupID string, topics []string) (*DataSource, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return nil, err
	}
	source := &DataSource{Data: make(chan *pb.Event)}
	go func() {
		defer consumer.Close()
		defer close(source.Data)
		for {
			msg, err := consumer.ReadMessage(-1)
			if err != nil {
				slog.Error("Consumer error: %v (%v)\n", err, msg)
				continue
			}
			event := &pb.Event{}
			source.Data <- event
		}
	}()

	return source, nil
}

func FromText(data []string) (*DataSource, error) {
	source := &DataSource{Data: make(chan *pb.Event, len(data))}
	defer close(source.Data)
	for _, v := range data {
		event := &pb.Event{Data: []byte(v)}
		source.Data <- event
	}
	return source, nil
}

func (st *DataStream) Exec() error {

	return nil
}

func (st *DataStream) SetParallelism(parallelism uint) *DataStream {
	//get the last operation and set parallelism if exist
	// or set the global parallelism
	if len(st.Operations) == 0 {
		// global
		st.Parallelism = parallelism
	} else {
		// single oper
		st.OperaPara[&st.Operations[len(st.Operations)-1]] = parallelism
	}
	return st
}

func (st *DataStream) Map(mapOp MapOp) *DataStream {
	operation := Operation{
		Type: pb.OpType_Reduce,
		Func: func(outputFunc func(data []byte, param ...any)) {
			for d := range st.Source.Data {
				outputFunc(mapOp.Map(d.Data))
			}
		},
	}
	st.Operations = append(st.Operations, operation)
	return st
}

func (st *DataStream) Reduce(initialValue []byte, reduceOp ReduceOp) *DataStream {
	// init calculate val
	accumulator := initialValue
	operation := Operation{
		Type: pb.OpType_Reduce,
		Func: func(outputFunc func(data []byte, param ...any)) {
			for d := range st.Source.Data {
				if accumulator == nil {
					accumulator = d.Data
				} else {
					accumulator = reduceOp.Reduce(accumulator, d.Data)
				}
				outputFunc(accumulator)
			}
		},
	}
	st.Operations = append(st.Operations, operation)
	return st
}

func (st *DataStream) ReduceWithoutInit(reduceOp ReduceOp) *DataStream {
	return st.Reduce(nil, reduceOp)
}

type KeyedDataStream struct {
	*DataStream
	KeySelector KeySelector
}

func (st *DataStream) KeyBy(ks KeySelector) *DataStream {
	operation := Operation{
		Type: pb.OpType_KeyBy,
		Func: func(outputFunc func(data []byte, param ...any)) {
			for d := range st.Source.Data {
				k := ks.GetKey(d.Data)
				tu := &Tuple{
					Key:   d.Data,
					Value: k,
				}
				outputFunc(anyToBytes(tu), k)
			}
		},
	}
	st.Operations = append(st.Operations, operation)
	return st
}

// func (st *KeyedDataStream) Reduce(initialValue []byte, reduceOp ReduceOp) *DataStream {
// 	// init calculate val
// 	accumulator := initialValue
// 	operation := Operation{
// 		Type: pb.OpType_KeyByReduce,
// 		Func: func(outputFunc func(data []byte, param ...any)) {
// 			output := [][]byte{}
// 			for d := range st.Source.Data {
// 				if accumulator == nil {
// 					accumulator = d.Data
// 				} else {
// 					accumulator = reduceOp.Reduce(accumulator, d.Data)
// 				}
// 			}
// 			output = append(output, accumulator)
// 			return output
// 		},
// 	}
// 	st.Operations = append(st.Operations, operation)
// 	return st.DataStream
// }

func anyToBytes(value any) []byte {
	switch v := value.(type) {
	case string:
		return []byte(v)
	case int:
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(v))
		return b
	case int32:
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(v))
		return b
	case int64:
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(v))
		return b
	default:
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(v)
		if err != nil {
			panic(fmt.Sprintf("unsupported type or encoding error: %v", err))
		}
		return buf.Bytes()
	}
}
