package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/forgoer/openssl"
	"slgserver/util"
	"strconv"
	"strings"
	"time"
)

//有效时间30天
const validTime = 30*24*time.Hour
const key = ("1234567890123456")

type Session struct {
	MTime 	time.Time
	Uid   	int
}

func NewSession(uid int, time time.Time) *Session {
	return &Session{Uid: uid, MTime: time}
}

func ParseSession(session string) (*Session, error) {
	decode, err := base64.StdEncoding.DecodeString(session)
	if err != nil{
		return nil, err
	}

	data, _ := util.AesCBCDecrypt(decode, []byte(key), []byte(key),openssl.ZEROS_PADDING)
	arr := strings.Split(string(data), "|")
	if len(arr) != 2 {
		return nil, errors.New("session format error")
	}

	int, err := strconv.Atoi(arr[0])
	if err != nil{
		return nil, err
	}

	time, err :=time.Parse("2006-01-02 15:04:05", arr[1])
	if err != nil{
		return nil, err
	}

	return &Session{Uid: int, MTime: time}, nil
}

func (self * Session) String() string {
	timeStr := self.MTime.Format("2006-01-02 15:04:05")
	str := fmt.Sprintf("%d|%s", self.Uid, timeStr)
	data, _ := util.AesCBCEncrypt([]byte(str), []byte(key), []byte(key),openssl.ZEROS_PADDING)
	encode := base64.StdEncoding.EncodeToString(data)
	return encode
}

func (self *Session) IsValid() bool {
	diff := time.Now().Sub(self.MTime)
	return diff - time.Duration(validTime) < 0
}
