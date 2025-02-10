package structFunc

import (
	"fmt"
	"github.com/gookit/goutil"
	"github.com/gookit/goutil/strutil"
	"net"
	"strings"
)

type User struct {
	UserName    string
	UserAddress string
	UserChannel chan string

	//用户的连接地址
	UserConnect net.Conn

	server *Server
}

// 创建用户api
func NewUser(userConnect net.Conn, server *Server) *User {
	userAddress := userConnect.RemoteAddr().String()

	user := &User{
		UserName:    userAddress,
		UserAddress: userAddress,
		UserChannel: make(chan string),
		UserConnect: userConnect,
		server:      server,
	}

	//启动go程，监听服务端发送的消息
	go user.ListenMessage()

	return user
}

// 用户上线接口
func (this *User) Online() {
	//对公共资源进行同步加锁，解锁
	this.server.MapLock.Lock()
	this.server.OnlineUserMap[this.UserName] = this
	this.server.MapLock.Unlock()

	//发送上线推送
	this.server.BroadCastMsg(this, "已上线")
}

// 用户下线接口
func (this *User) Offline() {
	//对公共资源进行同步加锁，解锁
	this.server.MapLock.Lock()
	delete(this.server.OnlineUserMap, this.UserName)
	this.server.MapLock.Unlock()

	//发送上线推送
	this.server.BroadCastMsg(this, "已下线")
}

// 发送给自己的消息
func (this *User) connectWriteData(msg string) {
	_, err := this.UserConnect.Write([]byte(msg + "\n"))
	if err != nil {
		fmt.Println("UserConnect.Write error = ", err)
	}
}

// 用户消息推送
func (this *User) DoMessage(msg string) {

	//加入查询在线人数判断，通过who指令判断
	if msg == "who" {

		this.server.MapLock.Lock()
		//遍历在线map
		for _, user := range this.server.OnlineUserMap {
			onlineUserMsg := "[ip：" + user.UserAddress + "] " + "[name：" + user.UserName + "] 在线..."
			this.connectWriteData(onlineUserMsg)
		}

		this.server.MapLock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式：rename|xxx
		newName := strings.Split(msg, "|")[1]

		//判断新名称是否非法
		if goutil.IsEmpty(newName) || strutil.IsBlank(newName) {
			this.connectWriteData("你输入的名称为空")
			return
		}

		//判断名字是否已经使用
		_, isExist := this.server.OnlineUserMap[newName]
		if isExist {
			//名字已占用
			this.connectWriteData(newName + "已被使用")
			return
		}

		//修改用户名称
		this.server.MapLock.Lock()

		//移除旧的key，新增新的key
		delete(this.server.OnlineUserMap, this.UserName)
		this.server.OnlineUserMap[newName] = this

		this.server.MapLock.Unlock()

		this.UserName = newName
		this.connectWriteData("您已更改用户名：" + newName)

	} else if len(msg) > 3 && msg[:3] == "to|" {
		//私聊消息格式to|xxx|消息内容
		//消息格式：rename|xxx
		remoteName := strings.Split(msg, "|")[1]

		//判断新名称是否非法
		if strutil.IsBlank(remoteName) {
			this.connectWriteData("消息格式不正确，请使用\"to|张三|消息内容\"")
			return
		}

		//判断名字是否已经使用
		_, isExist := this.server.OnlineUserMap[remoteName]
		if !isExist {
			//名字已占用
			this.connectWriteData(remoteName + "该用户不存在")
			return
		}

		chatContent := strings.Split(msg, "|")[2]
		if strutil.IsBlank(chatContent) {
			this.connectWriteData("你发发送的消息无内容")
			return
		}

		remoteUser := this.server.OnlineUserMap[remoteName]
		remoteUser.connectWriteData(this.UserName + "对你说：" + chatContent)

	} else {
		this.server.BroadCastMsg(this, msg)
	}
}

// 监听服务端发送的数据
func (this *User) ListenMessage() {
	for msg := range this.UserChannel {
		//将服务度消息写回客户端
		this.connectWriteData(msg)
	}
}
