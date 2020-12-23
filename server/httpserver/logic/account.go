package logic

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"math/rand"
	"slgserver/constant"
	"slgserver/db"
	myhttp "slgserver/server/httpserver"
	"slgserver/server/loginserver/model"
	"slgserver/util"
	"time"
)

type UserLogic struct{}
var DefaultUser = UserLogic{}

func (self UserLogic) CreateUser(ctx echo.Context) error {
	account := ctx.QueryParam("username")
	pwd := ctx.QueryParam("password")
	hardware := ctx.QueryParam("hardware")

	if len(account) > 0 && len(pwd) > 0 {
		if self.UserExists("username", account) {
			return myhttp.New("账号已经存在", constant.UserExist)
		}

		passcode := fmt.Sprintf("%x", rand.Int31())
		user := &model.User{
			Username: account,
			Passcode: passcode,
			Passwd: util.Password(pwd, passcode),
			Hardware: hardware,
			Ctime: time.Now(),
			Mtime: time.Now()}

		if _, err := db.MasterDB.Insert(user); err != nil {
			return myhttp.New("数据库出错", constant.DBError)
		} else{
			return nil
		}
	}else{
		return myhttp.New("用户名或密码是空", constant.InvalidParam)
	}
}

func (self UserLogic) ChangePassword(ctx echo.Context) error {
	account := ctx.QueryParam("username")
	pwd := ctx.QueryParam("password")
	newpwd := ctx.QueryParam("newpassword")

	user := &model.User{}
	if len(account) > 0 && len(pwd) > 0 && len(newpwd) > 0{
		if _, err := db.MasterDB.Where("username=?", account).Get(user); err != nil {
			return myhttp.New("数据库出错", constant.DBError)
		}else{
			if util.Password(pwd, user.Passcode) == user.Passwd {
				passcode := fmt.Sprintf("%x", rand.Int31())
				changeData := map[string]interface{}{
					"passwd": util.Password(newpwd, passcode),
					"passcode": passcode,
					"Mtime": time.Now(),
				}

				if _, err := db.MasterDB.Table(user).Where("username=?", account).Update(changeData); err !=nil {
					return myhttp.New("数据库出错", constant.DBError)
				}else{
					return nil
				}

			}else{
				return myhttp.New("原密码错误", constant.PwdIncorrect)
			}
		}

	}else{
		return myhttp.New("用户名或密码是空", constant.InvalidParam)
	}
}

func (UserLogic) UserExists(field, val string) bool {
	userLogin := &model.User{}
	_, err := db.MasterDB.Where(field+"=?", val).Get(userLogin)
	if err != nil || userLogin.UId == 0 {
		return false
	}
	return true
}
