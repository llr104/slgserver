package logic

import (
	"sync"
	"time"

	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/chatserver/proto"
)

type Group struct {
	userMutex sync.RWMutex
	msgMutex  sync.RWMutex
	users     map[int]*User
	msgs      ItemQueue
}

func NewGroup() *Group {
	return &Group{users: map[int]*User{}}
}

func (this*Group) Enter(u *User) {
	this.userMutex.Lock()
	defer this.userMutex.Unlock()
	this.users[u.rid] = u
}

func (this*Group) Exit(rid int) {
	this.userMutex.Lock()
	defer this.userMutex.Unlock()
	delete(this.users, rid)
}

func (this*Group) GetUser(rid int) *User {
	this.userMutex.Lock()
	defer this.userMutex.Unlock()
	return this.users[rid]
}

func (this*Group) PutMsg(text string, rid int, t int8) *proto.ChatMsg {

	this.userMutex.RLock()
	u, ok := this.users[rid]
	this.userMutex.RUnlock()
	if ok == false{
		return nil
	}

	msg := &Msg{Msg: text, RId: rid, Time: time.Now(), NickName: u.nickName}
	this.msgMutex.Lock()
	size := this.msgs.Size()
	if size > 100 {
		this.msgs.Dequeue()
	}
	this.msgs.Enqueue(msg)
	this.msgMutex.Unlock()

	//广播
	this.userMutex.RLock()
	c := &proto.ChatMsg{RId: msg.RId, NickName: msg.NickName, Time: msg.Time.Unix(), Msg: msg.Msg, Type: t}
	for _, user := range this.users {
		net.ConnMgr.PushByRoleId(user.rid, "chat.push", c)
	}
	this.userMutex.RUnlock()
	return c
}

func (this*Group) History() []proto.ChatMsg {
	r := make([]proto.ChatMsg, 0)
	this.msgMutex.RLock()
	items := this.msgs.items
	for _, item := range items {
		msg := item.(*Msg)
		c := proto.ChatMsg{RId: msg.RId, NickName: msg.NickName, Time: msg.Time.Unix(), Msg: msg.Msg}
		r = append(r, c)
	}
	this.msgMutex.RUnlock()

	return r
}