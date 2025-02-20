package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	pb "go-liteflow/pb"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
)

var (
	id    string
	op    string
	otmid string
	otid  string
	tmid  string
)

func main() {
	flag.StringVar(&id, "id", "", "task id")
	flag.StringVar(&op, "op", "", "operation")
	flag.StringVar(&otmid, "otmid", "", "")
	flag.StringVar(&otid, "otid", "", "")
	flag.StringVar(&tmid, "tmid", "", "")
	flag.Parse()

	conn, err := net.Dial("tcp", "127.0.0.1:20023")
	if err != nil {
		fmt.Printf("failed to connect to coordinator: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Printf("optask: %s, id:%s, tmid: %s \n", op, id, tmid)

	for {
		if op == pb.OpType_Map.String() {
			for i := 0; i < 100; i++ {
				outputEvent := &pb.Event{
					Id:                  uuid.New().String(),
					EventTime:           time.Now().Unix(),
					EventType:           pb.EventType_Output,
					OutputTaskManagerId: strings.Split(otmid, ",")[0],
					OutputTaskId:        strings.Split(otid, ",")[0],
					TaskManagerId:       tmid,
					TaskId:              id,
				}
				for _, word := range []string{"hello", "world"} {
					outputEvent.Data = []byte(word)
					bytes, _ := json.Marshal(outputEvent)
					conn.Write(bytes)
					conn.Write([]byte("\n"))
				}
				time.Sleep(1 * time.Second)
			}
			continue
		}

		reader := bufio.NewReader(conn)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("failed to receive event: %v\n", err)
			return
		}
		fmt.Printf("recv event: %+v\n", line)

		var ev pb.Event
		err = json.Unmarshal([]byte(line), &ev)
		if err != nil {
			fmt.Printf("failed to unmarshal event: %v\n", err)
			continue
		}
		if op == pb.OpType_Reduce.String() {
			words := ReduceOp(string(ev.Data))
			fmt.Printf("reduce result: %v\n", words)
		}
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-ch
}

func MapOp(s string) []string {
	return strings.Split(s, " ")
}

var m = make(map[string]int)
func ReduceOp(s string) map[string]int {
	for _, v := range strings.Split(s, " ") {
		m[v]++
	}
	return m
}
