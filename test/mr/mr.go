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

	fmt.Printf("optask: %s, id: %s, tmid: %s \n", op, id, tmid)

	otmid := strings.Split(otmid, ",")
	otid := strings.Split(otid, ",")

	establishEvent := &pb.Event{
		Id:            uuid.New().String(),
		EventTime:     time.Now().Unix(),
		EventType:     pb.EventType_Establish,
		TaskManagerId: tmid,
		TaskId:        id,
	}
	bytes, _ := json.Marshal(establishEvent)
	conn.Write(bytes)
	conn.Write([]byte("\n"))

	if op == "Map" {
		Map(conn, otmid, otid, tmid, id)
	} else if op == "Reduce" {
		Reduce(conn, otmid, otid, tmid, id)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-ch
}

func Map(conn net.Conn, otmid []string, otid []string, tmid string, id string) {
	words := []string{"hello", "world"}
	for i := 0; i < 100; i++ {
		for i := range words {
			outputEvent := &pb.Event{
				Id:            uuid.New().String(),
				EventTime:     time.Now().Unix(),
				EventType:     pb.EventType_Output,
				TaskManagerId: tmid,
				TaskId:        id,
			}
			if len(otmid) > 0 {
				outputEvent.OutputTaskManagerId = otmid[0]
			}
			if len(otid) > 0 {
				outputEvent.OutputTaskId = otid[0]
			}
			outputEvent.Data = []byte(words[i])
			bytes, _ := json.Marshal(outputEvent)

			
			n, err := conn.Write(bytes)
			if err != nil {
				fmt.Printf("failed to write event: %v, n: %d\n", err, n)
				return
			}
			fmt.Printf("Map send %d bytes event: %+v\n", n, outputEvent)
			n, err = conn.Write([]byte("\n"))
			if err != nil {
				fmt.Printf("failed to write newline: %v, n: %d\n", err, n)
				return
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func Reduce(conn net.Conn, otmid []string, otid []string, tmid string, id string) {
	reader := bufio.NewReaderSize(conn, 1024*1024*1024)
	for {
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
		words := ReduceOp(string(ev.Data))
		fmt.Printf("Reduce result: %v\n", words)
	}

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
