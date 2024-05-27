package main

import (
	"context"
	"fmt"
	"go-liteflow/internal/pkg/operator"
	pb "go-liteflow/pb"
	"strconv"
	"strings"
)

type Tokenizer struct{}

func (t *Tokenizer) Map(input []*pb.Event) (output []*pb.Event) {

	for _, ev := range input {
		for _, word := range strings.Split(string(ev.Data), " ") {
			output = append(output, &pb.Event{Key: word, Data: []byte("1")})
		}
	}
	return output
}

// should generate by program
func MapOpFn(opId string) operator.OpFn {
	t := new(Tokenizer)
	return func(ctx context.Context, input ...*pb.Event) (output []*pb.Event) {
		output = t.Map(input)
		return
	}
}

type Sum struct{}

func (s *Sum) Reduce(collector map[string]int, input []*pb.Event) (output []*pb.Event) {

	for _, ev := range input {
		if ev.EventType == pb.EventType_DataSent {
			for k, v := range collector {
				output = append(output, &pb.Event{Key: k, Data: []byte(strconv.Itoa(v))})
			}
			return
		}

		num, err := strconv.Atoi(string(ev.Data))
		if err != nil {
			fmt.Println(err)
			return
		}
		collector[ev.Key] += num
	}

	return output
}

// should generate by program
func ReduceOpFn(opId string) operator.OpFn {
	var (
		m = make(map[string]int)
		s = new(Sum)
	)

	return func(ctx context.Context, input ...*pb.Event) (output []*pb.Event) {
		return s.Reduce(m, input)
	}
}
