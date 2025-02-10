package structFunc

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// 服务端
type Server struct {
	Ip   string
	Port int

	//线上用户集合
	OnlineUserMap map[string]*User

	//服务端读写锁
	MapLock sync.RWMutex

	//上线广播channel
	ServerMessage chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:            ip,
		Port:          port,
		OnlineUserMap: make(map[string]*User),
		ServerMessage: make(chan string),
	}

	return server
}

// 启动服务端
func (this *Server) Start() {
	//socket listen，Listen函数是创建服务器
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen error:", err)
		return
	}

	//close listen socket
	defer listen.Close()

	//启动go程，通知有用户上线
	go this.SendMsgToUser()

	for {
		//accept
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listen.Accept error:", err)
			continue
		}

		//do biz handle
		go this.BizHandle(conn)
	}
}

// 广播给所有在线用户
func (this *Server) SendMsgToUser() {
	for {
		serverMsg := <-this.ServerMessage

		onlineUserMap := this.OnlineUserMap

		this.MapLock.Lock()
		for _, onlineUser := range onlineUserMap {
			onlineUser.UserChannel <- serverMsg
		}

		this.MapLock.Unlock()

	}
}

// 业务处理
func (this *Server) BizHandle(conn net.Conn) {
	//将上线的用户添加到map中
	newUser := NewUser(conn, this)

	//用户上线
	newUser.Online()

	aliveChannel := make(chan bool)

	//启动go程，接收客户端发送的消息
	go this.receiveMsg(newUser, aliveChannel)

	//判断用户无操作超时，强制登出，一直循环监听
	for {
		select {
		case <-aliveChannel:
			//重置定时器
		case <-time.After(300 * time.Second): //执行过这个定时器channel之后，会自动重置10秒定时
			//给用户发送你已下线消息
			newUser.connectWriteData("您处于长期不活跃，已被强制下线....")

			this.MapLock.Lock()
			delete(this.OnlineUserMap, newUser.UserName)
			this.MapLock.Unlock()

			//关闭资源
			close(newUser.UserChannel)

			err := newUser.UserConnect.Close()
			if err != nil {
				fmt.Println("user.UserConnect.Close error ", err)
			}
			return
		}
	}
}

func (this *Server) BroadCastMsg(user *User, msg string) {
	sendMsg := "[ip：" + user.UserAddress + "] " + "[name：" + user.UserName + "] " + msg
	this.ServerMessage <- sendMsg
}

// 接收客户端发来的消息
func (this *Server) receiveMsg(user *User, aliveChannel chan bool) {
	receiveBuffer := make([]byte, 4080)

	for {
		length, err := user.UserConnect.Read(receiveBuffer)
		if length == 0 {
			//该用户下线
			user.Offline()
			return
		}

		if err != nil && err != io.EOF {
			fmt.Println("user.UserConnect.Read error ", err)
			return
		}

		//截取末尾的\n，广播给其他客户端，切片读取索引，以及结束索引
		msg := string(receiveBuffer[:length-1])
		user.DoMessage(msg)

		aliveChannel <- true
	}
}
