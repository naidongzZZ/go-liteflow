package coordinator

import (
	"errors"
	"go-liteflow/internal/core"
	"go-liteflow/internal/pkg"
	pb "go-liteflow/pb"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// 注册服务信息
func (co *coordinator) RegistServiceInfo(si *pb.ServiceInfo) (err error) {
	if err = uuid.Validate(si.Id); err != nil {
		return err
	}
	if !pkg.ValidateIPv4WithPort(si.ServiceAddr) {
		return errors.New("addr is illegal")
	}
	co.mux.Lock()
	defer co.mux.Unlock()

	// TODO  service_status change? clean unused conn
	info, ok := co.serviceInfos[si.Id]
	if !ok {
		conn, err := grpc.Dial(si.ServiceAddr, grpc.WithInsecure())
		if err != nil {
			return err
		}	

		// 注册服务信息, 并初始化client
		info = &core.Service{
			ServiceInfo: pb.ServiceInfo{
				Id:            si.Id,
				ServiceAddr:   si.ServiceAddr,
				ServiceType:   si.ServiceType,
				ServiceStatus: si.ServiceStatus,
				Timestamp:     si.Timestamp,
			},
			ClientConn: pb.NewCoreClient(conn),
		}
		co.serviceInfos[si.Id] = info
	} 

	// 更新服务状态和时间戳
	info.ServiceStatus = pb.ServiceStatus_SsRunning
	info.Timestamp = time.Now().Unix()

	// TODO 超时自动更新为SsOffline, 或者太久没更新时间戳也认为下线状态
	return nil
}

// 根据服务id批量获取服务信息
func (co *coordinator) GetServiceInfo(ids ...string) map[string]*pb.ServiceInfo {
	tmp := make(map[string]*pb.ServiceInfo)

	co.mux.Lock()
	defer co.mux.Unlock()

	if len(ids) == 0 {
		for id, info := range co.serviceInfos {
			tmp[id] = &info.ServiceInfo
		}	
		return tmp
	}

	for _, id := range ids {
		tmp[id] = &co.serviceInfos[id].ServiceInfo
	}
	return tmp
}

// 获取空闲的服务
func (co *coordinator) GetIdle(st pb.ServiceType, ss pb.ServiceStatus) (srv []*core.Service) {
	co.mux.Lock()
	defer co.mux.Unlock()

	for _, si := range co.serviceInfos {
		if si.ServiceType == st && si.ServiceStatus == ss {
			srv = append(srv, si)
		}
	}
	return srv
}