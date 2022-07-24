package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"grpc-conoha/api/conoha"
	"grpc-conoha/config"
	conohapb "grpc-conoha/pkg/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type conohaServer struct {
	conohapb.UnimplementedConohaServiceServer
}

var statusName = map[string]string{
	"SHUTOFF":       "シャットダウンしてるよ",
	"ACTIVE":        "起動してるよ",
	"RESIZE":        "リサイズ中",
	"REBOOT":        "再起動中",
	"VERIFY_RESIZE": "リサイズ承認待ち",
}

// サービスメソッドのサンプル
func (s *conohaServer) Minecraft(req *conohapb.MinecraftRequest, stream conohapb.ConohaService_MinecraftServer) error {
	token := conoha.GetToken(config.Config.Username, config.Config.Password, config.Config.TenantId)

	if req.GetCommand() == "!conoha server" {
		status, _ := conoha.GetServerStatus(token)

		if err := stream.Send(&conohapb.MinecraftResponse{
			Message:  statusName[string(status)],
			IsNormal: true,
		}); err != nil {
			return err
		}
		return nil
	}
	if req.GetCommand() == "!conoha start" {
		status, statusCode := conoha.StartServer(token)
		is_normal := true
		if statusCode != 202 {
			is_normal = false
		}
		return &conohapb.MinecraftResponse{
			Message:  string(status),
			IsNormal: is_normal,
		}, nil
	}
	if req.GetCommand() == "!conoha stop" {
		status, statusCode := conoha.StopServer(token)
		is_normal := true
		if statusCode != 202 {
			is_normal = false
		}
		return &conohapb.MinecraftResponse{
			Message:  string(status),
			IsNormal: is_normal,
		}, nil
	}
	if req.GetCommand() == "!conoha reboot" {
		status, statusCode := conoha.RebootServer(token)
		is_normal := true
		if statusCode != 202 {
			is_normal = false
		}
		return &conohapb.MinecraftResponse{
			Message:  string(status),
			IsNormal: is_normal,
		}, nil
	}
	grpcerr := status.Error(codes.Unimplemented, "登録されていないコマンドです")
	return nil, grpcerr
}

// 自作サービス構造体のコンストラクタ
func NewConohaServer() *conohaServer {
	return &conohaServer{}
}

func main() {
	port := 8080
	listner, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	// gRPCサーバーを作成
	s := grpc.NewServer()

	// gRPCサーバーにserviceを登録
	conohapb.RegisterConohaServiceServer(s, NewConohaServer())

	// grpcURL用にサーバーリフレクションを設定する
	reflection.Register(s)

	go func() {
		log.Printf("start gRPC server port->%v", port)
		s.Serve(listner)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("stopping gRPC server")
	s.GracefulStop()
}
