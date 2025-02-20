package task_manager

import (
	"bufio"
	"context"
	"encoding/json"
	"go-liteflow/internal/pkg/log"
	pb "go-liteflow/pb"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (tm *taskManager) InitEventChannel() {

	lis, err := net.Listen("tcp", ":20023")
	if err != nil {
		log.Errorf("listen tcp fail, err: %v", err)
		panic(err)
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Errorf("accept tcp fail, err: %v", err)
			continue
		}

		go func(conn net.Conn) {
			defer func() {
				conn.Close()
			}()
			log.Infof("accept conn: %v", conn.RemoteAddr())
			for {
				reader := bufio.NewReader(conn)
				line, err := reader.ReadString('\n')
				if err != nil {
					log.Errorf("read tcp fail, err: %v", err)
				}

				var event pb.Event
				err = json.Unmarshal([]byte(line), &event)
				if err != nil {
					log.Errorf("unmarshal event fail, err: %v", err)
					return
				}
				tm.SetTaskConn(event.TaskId, conn)

				conn, ok := tm.GetTaskConn(event.OutputTaskId)
				if ok {
					_, err := conn.Write([]byte(line))
					if err != nil {
						log.Errorf("write event fail, err: %v", err)
						continue
					}
					if _, err = conn.Write([]byte("\n")); err != nil {
						log.Errorf("write newline fail, err: %v", err)
					}
				} else {
					servnfo, ok := tm.serviceInfos[event.OutputTaskManagerId]
					if !ok {
						log.Errorf("service info not found, tmid: %s", event.OutputTaskManagerId)
						continue
					}
					_, err := servnfo.ClientConn.EmitEvent(context.Background(), &event)
					if err != nil {
						log.Errorf("emit event fail, err: %v", err)
					}
				}
			}

		}(conn)
	}
}

func (tm *taskManager) EmitEvent(ctx context.Context, event *pb.Event) (resp *pb.EmitEventResp, err error) {
	conn, ok := tm.GetTaskConn(event.OutputTaskId)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "task conn not found, task_id: %s", event.OutputTaskId)
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshal event fail, err: %v", err)
	}

	if _, err = conn.Write(bytes); err != nil {
		return nil, status.Errorf(codes.Internal, "write event fail, err: %v", err)
	}
	if _, err = conn.Write([]byte("\n")); err != nil {
		return nil, status.Errorf(codes.Internal, "write newline fail, err: %v", err)
	}

	return &pb.EmitEventResp{}, nil
}
