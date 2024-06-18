package test

import (
	"fmt"
	"go-liteflow/pkg"
	"strings"
	"testing"
)

func TestExecute(t *testing.T) {
	data := []string{"1,Apple", "2,Banana", "1,Cherry"}
	ds, err := pkg.FromText(data)
	if err != nil {
		t.Fatal(err)
	}
	st := pkg.NewDataStream(ds)
	mo := &MyOperator{}
	st.Map(mo).ReduceWithoutInit(mo).KeyBy(mo)

	// exec
	input := st.Source.Data
	// for _, operation := range st.Operations {
	// 	resultList := [][]byte{}
	// 	// resultList = mergeResult(resultList, operation.Func())
	// 	input = make(chan *pb.Event, len(resultList))
	// 	//将output结果给到input并情况output
	// 	for _, r := range resultList {
	// 		input <- &pb.Event{
	// 			Data: r,
	// 		}
	// 	}
	// 	st.Source.Data = input
	// 	close(input)
	// }
	//输出结果
	for d := range input {
		fmt.Println(string(d.Data))
	}
}

type MyOperator struct{}

func (MyOperator) Map(val []byte) []byte {
	valStr := string(val) + "->map|"
	return []byte(valStr)
}

func (MyOperator) Reduce(val1, val2 []byte) []byte {
	val1Str := string(val1)
	val2Str := string(val2)
	return []byte(val1Str + "|" + val2Str)
}

func (MyOperator) GetKey(val []byte) []byte {
	str := string(val)
	parts := strings.Split(str, ",")
	return []byte(parts[0])
}

func mergeResult(resultList [][]byte, result [][]byte) [][]byte {
	resultList = append(resultList, result...)
	return resultList
}
