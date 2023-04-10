/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-04 22:57:23
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-10 13:16:36
 */
package main

import (
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"unpack"

	"github.com/jian308/go/conf"
	"github.com/jian308/go/log"
)

// 代理组 8001=>"192.168.31.168:3008"
var proxys map[string]string
var server = ":8000"
var td = time.NewTicker(5 * time.Second)

func chonglian() {
	td.Reset(5 * time.Second)
	for range td.C {
		log.Debug("重连服务端")
		start()
	}
}

func start() {
	go node()
	time.Sleep(time.Second)
	for k, v := range proxys {
		go StartProx(k, v) //先发起一条链接
	}
}

//需要开一条通道用来通信处理需要请求发起链接

func node() {
	crpsconn, err := net.Dial("tcp", server)
	if err != nil {
		log.Debugf("Connect tcp err: %v", err)
		return
	}
	td.Stop()
	for k := range proxys {
		err = unpack.Encode(crpsconn, []byte("1"+k)) //首次握手发送
		if err != nil {
			log.Debugf("Unpack err: %v", err)
			return
		}
	}
	for {
		content, err := unpack.Decode(crpsconn)
		if err != nil {
			log.Debugf("接收错误: %v", err)
			crpsconn.Close()
			chonglian() //重连
			return
		}
		if content[0] == '0' {
			p := string(content[1:])
			log.Debug("对方申请需求链接")
			if v, ok := proxys[p]; ok {
				StartProx(p, v)
			}
		}
	}
}

// 加载配置
func loadcfg() {
	conf.Auto()
	if conf.Get("common") == nil {
		log.Fatal("未找到配置")
	}
	server = conf.Get("common.server").(string)
	unpack.MsgToken = conf.Get("common.token").(string)
	proxys = make(map[string]string, 1024)
	//笨方法先用着
	for i := 0; i < 100; i++ {
		proxy := "proxy_" + strconv.Itoa(i)
		if conf.Get(proxy) != nil {
			log.Debug(proxy)
			proxys[conf.Get(proxy+".remote_port").(string)] = conf.Get(proxy + ".local_addr").(string)
		}
	}
	log.Debug(proxys)
}

// 指定端口维持5个链接
func main() {
	loadcfg()
	td.Stop()
	start()
	log.Info("服务启动成功!")
	//优雅关闭开启的服务
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	siger := <-c
	if siger == syscall.SIGINT {
		//方便调试的时候直接关闭
		log.Infof("ctrl+c直接关闭:%d", syscall.Getpid())
		os.Exit(0)
	}
	if siger == syscall.SIGTERM {
		log.Infof("开启无忧关闭...等待处理完请求将关闭进程id:%d", syscall.Getpid())
		log.Info("成功关闭!")
	}
}

func StartProx(p, t string) {
	log.Debug("启动一个链接")
	//先链接本地端口
	localconn, err := net.Dial("tcp", t)
	if err != nil {
		log.Debugf("Connect tcp err: %v", err)
		return
	}
	crpsconn, err := net.Dial("tcp", server)
	if err != nil {
		log.Debugf("Connect tcp err: %v", err)
		return
	}
	err = unpack.Encode(crpsconn, []byte("0"+p)) //首次握手发送
	if err != nil {
		log.Debugf("Unpack err: %v", err)
		return
	}
	//defer StartProx(p, t)
	go func() {
		_, _ = io.Copy(localconn, crpsconn)
		if err != nil {
			log.Error(err)
		}
		localconn.Close()
	}()
	go func() {
		_, _ = io.Copy(crpsconn, localconn)
		if err != nil {
			log.Error(err)
		}
		crpsconn.Close()
	}()
}
