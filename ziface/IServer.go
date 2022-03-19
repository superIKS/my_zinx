package ziface

// IServer 定义一个服务器接口
type IServer interface {
	// Start 启动
	Start()
	// Stop 停止
	Stop()
	// Serve 运行
	Serve()
}