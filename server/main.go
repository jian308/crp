/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-04 22:57:17
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-10 13:12:49
 */
package main

import (
	"io"
	"net"
	"server/pools"
	"sync"
	"time"
	"unpack"

	"github.com/jian308/go/conf"
	"github.com/jian308/go/log"
)

var nodes sync.Map
var poolmu sync.Mutex

var listen string

// 加载配置
func loadcfg() {
	conf.Auto()
	if conf.Get("common.listen") == nil {
		log.Fatal("未找到配置")
	}
	listen = conf.Get("common.listen").(string)
	unpack.MsgToken = conf.Get("common.token").(string)
}

func main() {
	loadcfg()
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		log.Debug("侦听错误：", err.Error())
		return
	}
	defer listener.Close()
	log.Debug("正在侦听" + listen)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Debug("接受错误：", err.Error())
			return
		}
		go handnode(conn)
	}
}

func handnode(conn net.Conn) {
	clientip := conn.RemoteAddr().String()
	log.Debug("接受来自：", clientip, "的连接")
	for {
		content, err := unpack.Decode(conn)
		if err != nil {
			log.Debugf("接收错误: %v", err)
			conn.Close()
			return
		}
		switch content[0] {
		case '0':
			p := string(content[1:])
			_, ok := nodes.Load("node" + p) //c
			if !ok {
				log.Debug("节点已经关闭")
				return
			}
			// //这个ip限制暂时不用限制 因为客户端ip为多ip
			// if fn.GetIp(c.(net.Conn)) != fn.GetIp(conn) {
			// 	log.Debugf("端口%s已经被其他人使用", p)
			// 	return
			// }
			poolmu.Lock()
			_, ok = ports.Load(p)
			poolmu.Unlock()
			if !ok {
				log.Debug("新开端口", p)
				go Pool(p)
				time.Sleep(time.Second)
			}
			portmap := pools.GetPort(p)
			portmap.Add(conn)
			return
		case '1':
			p := string(content[1:])
			log.Debug("1=>node" + p)
			defer func(p string) {
				if plisten, ok := ports.Load(p); ok {
					if listern, ok := plisten.(net.Listener); ok {
						log.Debugf("关闭%s监听", p)
						listern.Close()
					}
				}
				ports.Delete(p) //关闭后会自动删除
				nodes.Delete("node" + p)
				pools.DelPort(p)
				conn.Close()
			}(p)
			_, ok := nodes.Load("node" + p)
			if !ok {
				log.Debug("node"+p, "成功上线")
				nodes.Store("node"+p, conn)
			}
		default:
			log.Debug("其他指令", string(content))
		}
	}
}

var ports sync.Map

func Pool(p string) {
	listener, err := net.Listen("tcp", ":"+p)
	if err != nil {
		log.Debug("侦听错误：", err.Error())
		return
	}
	var clients sync.Map
	defer func() {
		clients.Range(func(key, value any) bool {
			if client, ok := key.(net.Conn); ok {
				log.Debug("关闭所有client:", client.RemoteAddr())
				client.Close()
			}
			return true
		})
	}()
	defer listener.Close()
	defer ports.Delete(p) //关闭之前先删除ports
	log.Debug("正在侦听 :", p)
	ports.Store(p, listener)
	for {
		client, err := listener.Accept()
		if err != nil {
			log.Debug("接受错误：", err.Error())
			return
		}
		clients.Store(client, struct{}{})
		log.Debug("终端客户请求来了")
		go handproxy(client, p)
	}
}

func handproxy(client net.Conn, p string) {
	defer client.Close()
	//从端口对应的pools里拿一个
	portconn := pools.GetConn(p)
	if portconn == nil {
		nodeconn, ok := nodes.Load("node" + p)
		if !ok {
			log.Debugf("没有找到%s的链接", p)
			return
		}
		log.Debug("请求一条新链接", client.RemoteAddr())
		err := unpack.Encode(nodeconn.(net.Conn), []byte("0"+p)) //首次握手发送
		if err != nil {
			log.Debugf("Unpack err: %v", err)
			return
		}
		log.Debug("等待新的链接,等待开始", client.RemoteAddr())
		for i := 0; i < 100; i++ {
			time.Sleep(30 * time.Millisecond)
			portconn = pools.GetConn(p)
			if portconn != nil {
				break
			}
		}
		log.Debug("等待新的链接,等待结束", client.RemoteAddr())
		if portconn == nil {
			log.Debugf("没有找到%s的链接", p)
			return
		}
	}
	defer portconn.Close()
	go func() {
		_, err := io.Copy(client, portconn)
		if err != nil {
			log.Error(err)
			return
		}
	}()
	//go func() {
	_, err := io.Copy(portconn, client)
	if err != nil {
		log.Error(err)
		return
	}
	//}()
}
