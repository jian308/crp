/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-08 16:24:54
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-10 13:13:36
 */
package pools

import (
	"net"
	"sync"

	"github.com/jian308/go/log"
)

var ports = make(map[string]*Port, 1024)
var portsMu sync.RWMutex

type Port struct {
	Conns []net.Conn
	mu    sync.Mutex
}

func (p *Port) Add(c net.Conn) {
	if p == nil {
		log.Error("p未初始化")
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Conns = append(p.Conns, c)
	log.Debug("len(p.Conns)=>", len(p.Conns))
	log.Debug("增加一条", c.RemoteAddr())
}

func (p *Port) Del(c net.Conn) {
	if p == nil {
		log.Error("p未初始化")
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	j := 0
	//log.Debug("del=len(p.Conns)=>", len(p.Conns))
	for _, v := range p.Conns {
		if v != c {
			p.Conns[j] = v
			j++
		}
	}
	newcoon := make([]net.Conn, 0, 1024)
	copy(newcoon, p.Conns[:j])
	p.Conns = newcoon
	//p.Conns = p.Conns[:j]
	//log.Debug("del=len(p.Conns)=>", len(p.Conns))
	log.Debug("删除掉线链接", c.RemoteAddr())
}

func (p *Port) Get(t string) net.Conn {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.Conns) > 0 {
		getconn := p.Conns[0] //取出第一个
		// newcoon := make([]net.Conn, 0, 1024)
		// copy(newcoon, p.Conns[1:])
		//p.Conns = newcoon
		p.Conns = p.Conns[1:]
		log.Debug("取出一条链接", getconn.RemoteAddr())
		//log.Debug("len(p.Conns)=>", len(p.Conns))
		return getconn
	}
	return nil
}

func DelPort(p string) {
	portsMu.Lock()
	defer portsMu.Unlock()
	if portmap, ok := ports[p]; ok {
		//删除标识
		delete(ports, p)
		//关闭所有链接
		for _, v := range portmap.Conns {
			v.Close()
		}
	}
}

func InitPort(p string) *Port {
	portsMu.Lock()
	defer portsMu.Unlock()
	portmap, ok := ports[p]
	if !ok {
		log.Debug("初始化Port:", p)
		portmap = &Port{
			Conns: make([]net.Conn, 0, 1024),
		}
		ports[p] = portmap
	}
	return portmap
}

func GetPort(p string) *Port {
	portsMu.RLock()
	portmap, ok := ports[p]
	portsMu.RUnlock()
	if !ok {
		return InitPort(p)
	}
	return portmap
}

func GetConn(p string) net.Conn {
	portsMu.RLock()
	portmap, ok := ports[p]
	portsMu.RUnlock()
	if !ok {
		return nil
	}
	return portmap.Get(p)
}
