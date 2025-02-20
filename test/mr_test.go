package test

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"go-liteflow/internal/pkg/md5"
	pb "go-liteflow/pb"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func Test_SubmitOpTask(t *testing.T) {

	dir, _ := os.Getwd()
	t.Logf("dir: %s", dir)

	cmd := exec.Command("go", "build", "-o", dir+"/mr", dir+"/mr/mr.go")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build mr: %v", err)
	}

	os.Chmod(dir+"/mr/mr", 0755)

	ef, err := os.ReadFile(dir + "/mr/mr")
	if err != nil {
		t.Fatalf("failed to read mr: %v", err)
	}
	efHash := md5.Calc(ef)

	conn, err := grpc.Dial("127.0.0.1:20021", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to connect to coordinator: %v", err)
	}
	defer conn.Close()

	resp, err := pb.NewCoreClient(conn).SubmitOpTask(context.Background(),
		&pb.SubmitOpTaskReq{
			ClientId: uuid.New().String(),
			Digraph: &pb.Digraph{
				GraphId: "1",
				Adj: []*pb.OperatorTask{
					{
						Id:         "1",
						OpType:     pb.OpType_Map,
						State:      pb.TaskStatus_Ready,
						Downstream: []*pb.OperatorTask{{Id: "2"}},
					},
					{
						Id:         "2",
						OpType:     pb.OpType_Reduce,
						State:      pb.TaskStatus_Ready,
						Downstream: []*pb.OperatorTask{},
					},
				},
				EfHash: efHash,
			},
			Ef: ef,
		})
	t.Logf("resp: %+v, err: %v", resp, err)
}
