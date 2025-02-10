package main

import (
	"flag"
	"fmt"
	"github.com/gookit/goutil"
	"github.com/gookit/goutil/strutil"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp      string
	ServerPort    int
	Name          string
	ClientConnect net.Conn
	flag          int //操作模式
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	//连接服务端，Dial函数是与服务器建立连接
	clientConnect, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))

	if !goutil.IsNil(err) {
		fmt.Println("net.Dial error: ", err)
		return nil
	}

	client.ClientConnect = clientConnect

	return client
}

// 功能菜单
func (this *Client) menu() bool {
	var scanFlag int

	fmt.Println("1。公聊模式")
	fmt.Println("2。私聊模式")
	fmt.Println("3。更新用户名")
	fmt.Println("0。退出")

	//从控制台输入读取
	fmt.Scanln(&scanFlag)

	if scanFlag >= 0 && scanFlag <= 3 {
		this.flag = scanFlag
		return true
	} else {
		fmt.Println("请输入正确的模式数字......")
		return false
	}
}

// 处理服务端数据返回
func (this *Client) DealResponse() {
	//一旦conn有数据返回，就会输出到控制台，且永久阻塞监听
	io.Copy(os.Stdout, this.ClientConnect)

	//等价于
	/*for {
		buffer := make([]byte, 4096)
		this.ClientConnect.Read(buffer)
		if len(buffer) != 0 {
			fmt.Println(string(buffer))
		}
	}*/
}

// 公聊模式
func (this *Client) PublicChat() {
	fmt.Println(">>>>>>>>>请输入聊天内容，\"exit\"退出聊天>>>>>>")
	var chatMsg string
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if !strutil.IsBlank(chatMsg) {
			this.sendConnectMsg(chatMsg)
		}
		fmt.Scanln(&chatMsg)
	}
}

// 查询在线用户
func (this *Client) queryOnlineUser() {
	this.sendConnectMsg("who")
}

// 私聊模式
func (this *Client) PrivateChat() {
	this.queryOnlineUser()
	fmt.Println(">>>>>>>>>>请输入你要私聊的用户名，\"exit\"退出选择>>>>>>>>>>>")
	var remoteName string
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {

		fmt.Println(">>>>>>>>>请输入聊天内容，\"exit\"退出聊天>>>>>>")
		var chatMsg string
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if !strutil.IsBlank(chatMsg) {
				this.sendConnectMsg("to|" + remoteName + "|" + chatMsg)
			}
			fmt.Scanln(&chatMsg)
		}

		this.queryOnlineUser()
		fmt.Println(">>>>>>>>>>请输入你要私聊的用户名，\"exit\"退出选择>>>>>>>>>>>")
		fmt.Scanln(&remoteName)
	}

}

// 更新用户名
func (this *Client) UpdateName() bool {
	fmt.Println(">>>>>>>>>>请输入新的用户名>>>>>>>>>>>")
	fmt.Scanln(&this.Name)

	//组装命令行
	renameMsg := "rename|" + this.Name
	return this.sendConnectMsg(renameMsg)
}

func (this *Client) sendConnectMsg(msg string) bool {
	_, err := this.ClientConnect.Write([]byte(msg + "\n"))
	if !goutil.IsNil(err) {
		fmt.Println("conn.Write error：", err)
		return false
	}
	return true
}

func (this *Client) Run() {
	for this.flag != 0 {
		for !this.menu() {
		}

		switch this.flag {
		case 1:
			this.PublicChat()
			break
		case 2:
			this.PrivateChat()
			break
		case 3:
			this.UpdateName()
			break
		}
	}
}

// 初始化，提示特定命令行，输入ip与端口
var serverIp string
var serverPort int

// ./client/client -ip 127.0.0.1 -port 8088
func init() {
	//flag函数，处理命令行
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址(默认ip地址-127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8088, "设置服务器端口(默认端口-8088)")
}

func main() {
	//解析命令行
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if goutil.IsNil(client) {
		fmt.Println("连接服务器失败.......")
		return
	}

	fmt.Println("连接服务器成功.......")

	go client.DealResponse()

	client.Run()
}
