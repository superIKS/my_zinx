package znet

import (
	"fmt"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

// Server 定义一个Server服务器模块
type Server struct {
	//服务器名称
	Name string
	//监听ip版本
	IPVersion string
	//监听ip
	IP string
	//监听端口
	Port int
	//当前server的消息管理模块，用来绑定msgID和对应的处理业务API
	MsgHandler ziface.IMsgHandler
	//该Server的连接管理器
	ConnMgr ziface.IConnManager
	//该Server创建连接之后自动调用Hook函数--OnConnStart
	OnConnStart func(conn ziface.IConnection)
	//该Server销毁连接之前自动调用Hook函数--OnConnStop
	OnConnStop func(conn ziface.IConnection)
}

// Start 启动
func (s *Server) Start() {
	fmt.Printf("[Zinx] Server %s, listenner at %s:%d is starting...\n",
		utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TcpPort)
	fmt.Printf("[Zinx] Version %s, MaxConn %d\n",
		utils.GlobalObject.Version, utils.GlobalObject.MaxConn)

	go func() {
		//0 开启消息队列及工作池
		s.MsgHandler.StartWorkerPool()
		//1 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr error:", err)
			return
		}
		//2 监听服务器的地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen ", s.IP, "error ", err)
			return
		}
		fmt.Println("start Zinx server ", s.Name, "is listening...")

		var cid uint32
		cid = 0
		//3 阻塞等待客户端进行连接，处理客户端连接业务(读写)
		for {
			//如果有客户端连接，阻塞会返回
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept error ", err)
				continue
			}
			//如果超过最大连接的数量，则关闭当前连接
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				//TODO 给客户端响应一个超出最大连接的错误包
				fmt.Println("[Conn Failed] Too Many Connections...MaxConn = ", utils.GlobalObject.MaxConn)
				conn.Close()
				continue
			}

			//将处理新连接的业务方法和conn进行绑定，得到我们的连接模块
			dealConn := NewConnection(s, conn, cid, s.MsgHandler)
			cid++

			//启动当前的连接业务处理
			go dealConn.Start()
		}
	}()
}

// Stop 停止
func (s *Server) Stop() {
	//将服务器资源、状态、已经开辟的连接信息进行停止或回收
	fmt.Println("[Server Stop] Zinx server ", s.Name, " stopped.")
	s.ConnMgr.ClearConn()
}

// Serve 运行
func (s *Server) Serve() {
	//启动server的服务功能
	s.Start()

	//TODO 做一些启动服务器之后的额外业务

	//阻塞状态
	select {}
}

// AddRouter 路由功能：给当前服务注册一个路由方法，供客户端的连接处理使用
func (s *Server) AddRouter(msgID uint32, router ziface.IRouter) {
	s.MsgHandler.AddRouter(msgID, router)
	fmt.Println("Add Router Success!")
}

// GetConnMgr 获取当前Server的连接管理器
func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

// NewServer 初始化Server模块
func NewServer(name string) ziface.IServer {
	s := &Server{
		Name:       utils.GlobalObject.Name,
		IPVersion:  "tcp4",
		IP:         utils.GlobalObject.Host,
		Port:       utils.GlobalObject.TcpPort,
		MsgHandler: NewMsgHandler(),
		ConnMgr:    NewConnManager(),
	}
	return s
}

// SetOnConnStart 注册OnConnStart Hook函数
func (s *Server) SetOnConnStart(hookFunc func(conn ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

// SetOnConnStop 注册OnConnStop Hook函数
func (s *Server) SetOnConnStop(hookFunc func(conn ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

// CallOnConnStart 调用OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---> Call OnConnStart()...")
		s.OnConnStart(conn)
	}
}

// CallOnConnStop 调用OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("---> Call OnConnStop()...")
		s.OnConnStop(conn)
	}
}
