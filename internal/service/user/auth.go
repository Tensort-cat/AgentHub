package service

import (
	"AgentHub/internal/dao"
	"AgentHub/internal/dto/response"
	"AgentHub/internal/model"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/util"
	"AgentHub/pkg/zlog"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Login(email, password string) (constant.Code, constant.Msg, *response.LoginResp) {
	/*
		1. 根据 email 找记录，没有直接返回
		2. 校验密码，错误直接返回
		3. 返回正确信息
	*/
	var user model.User
	if err := dao.DB.First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 没有邮箱对应用户
			zlog.Info("用户不存在")
			return constant.BadRequest, "用户不存在", nil
		}
		return constant.InternalServerError, constant.Error, nil
	}

	if !util.CheckPwd(password, user.Password) {
		zlog.Info("密码错误")
		return constant.BadRequest, "密码错误", nil
	}

	// 登陆成功
	// 生成 token
	token, err := util.GenerateJWT(fmt.Sprint(user.ID))
	if err != nil {
		zlog.Error("jwt 令牌生成出错", zap.Error(err))
	}
	resp := &response.LoginResp{
		Token:  "Bearer " + token,
		ID:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Avatar: user.Avatar,
	}

	return constant.Success, constant.Ok, resp
}

func SendCaptcha(email string) (constant.Code, constant.Msg) {
	key := dao.GenCaptchaPrefix(email)
	captcha, err := dao.Get(key)
	if err == nil { // 验证码已经存在
		return constant.BadRequest, "请求频率过快"
	}
	if !errors.Is(err, redis.Nil) { // 内部错误
		return constant.InternalServerError, constant.Error
	}

	captcha, err = util.SendCaptcha(email)
	if err != nil {
		return constant.InternalServerError, "验证码发送失败"
	}
	zlog.Info("邮箱验证码", zap.String("captcha", captcha))

	if err = dao.SetEx(key, captcha, constant.CaptchaExpiration); err != nil {
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}

func Register(name, email, password, captcha string) (constant.Code, constant.Msg) {
	/*
		1. 检查邮箱格式
		2. 验证码校验
		3. 将用户信息存入数据库
		4. 返回成功的状态码即可
	*/
	// 检查邮箱格式
	if vaild := util.IsValidEmail(email); !vaild {
		zlog.Info("邮箱格式错误")
		return constant.BadRequest, "邮箱格式错误"
	}

	// 验证码校验
	key := dao.GenCaptchaPrefix(email)
	trueCap, err := dao.Get(key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return constant.BadRequest, "验证码不存在或已过期，请重新获取"
		}
		return constant.BadRequest, constant.Error
	}

	if trueCap != captcha {
		return constant.BadRequest, "验证码错误"
	}

	// 加密密码
	encryptPwd, err := util.HashPassword(password)
	if err != nil {
		return constant.BadRequest, constant.Error
	}
	user := model.User{
		BaseModel: model.BaseModel{
			ID:        util.GenID(),
			CreatedAt: time.Now(),
		},
		Name:     name,
		Email:    email,
		Password: encryptPwd,
	}

	err = dao.DB.Create(&user).Error
	if err != nil {
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}
