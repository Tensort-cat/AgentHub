package test

import (
	"AgentHub/internal/config"
	"AgentHub/internal/dao"
	"AgentHub/internal/dao/rabbitmq"
	"AgentHub/internal/model"
	workflow_enum "AgentHub/pkg/enum/workflow"
	"AgentHub/pkg/util"
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
)

func initEnv() {
	config.InitConfig()
	ctx := context.Background()
	if err := dao.InitRedisCli(ctx); err != nil {
		panic(err)
	}

	if err := dao.InitMySQL(); err != nil {
		panic(err)
	}
}

func TestSelect(t *testing.T) {
	initEnv()

	var models []model.Model
	res := dao.DB.Where("user_id = ?", "2082298060007800832").Order("created_at asc").Find(&models)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 没有数据
			t.Log("无数据")
		}
		t.Fatal(err)
	}
	t.Log(models)
	t.Log(len(models))
}

func TestInsertNode(t *testing.T) {
	initEnv()
	defer dao.CloseMySQL()
	defer dao.CloseRedis()

	node1 := model.WorkflowNode{
		BaseModel: model.BaseModel{
			ID:        util.GenID(),
			CreatedAt: time.Now(),
		},
		WorkflowID: 2083113594655866880,
		Name:       "EMT",
		Type:       workflow_enum.ChatModel,
		PositionX:  10,
		PositionY:  10,
		Config:     []byte(`{"sys_prompt": "你叫艾米莉亚，是一个异世界的银发半精灵美少女"}`),
	}
	t.Log(node1)

	node2 := model.WorkflowNode{
		BaseModel: model.BaseModel{
			ID:        util.GenID(),
			CreatedAt: time.Now(),
		},
		WorkflowID: 2083113594655866880,
		Name:       "486",
		Type:       workflow_enum.ChatModel,
		PositionX:  15,
		PositionY:  15,
		Config:     []byte(`{"sys_prompt": "你是一个穿越到异世界的男高中生，拥有死亡回档的能力，你爱上了艾米莉亚"}`),
	}
	t.Log(node2)

	if err := dao.DB.Create(&node1).Error; err != nil {
		t.Error(err)
	}
	if err := dao.DB.Create(&node2).Error; err != nil {
		t.Error(err)
	}
}

func TestInsertEdge(t *testing.T) {
	initEnv()
	defer dao.CloseMySQL()
	defer dao.CloseRedis()

	edge := model.WorkflowEdge{
		BaseModel: model.BaseModel{
			ID:        util.GenID(),
			CreatedAt: time.Now(),
		},
		WorkflowID:   2083113594655866880,
		SourceNodeID: 2084272267944001537,
		TargetNodeID: 2084272267944001536,
	}
	if err := dao.DB.Create(&edge).Error; err != nil {
		t.Error(err)
	}
}

func TestHMSet(t *testing.T) {
	initEnv()

	ctx := context.Background()
	ret, err := dao.RedisCli.HMSet(ctx, "111:222", "name", "xxx.md", "chunk_num", 123).Result()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)
}

func TestRabbitMQ(t *testing.T) {
	initEnv()

	cfg := config.Cfg.RabbitmqConfig
	t.Log(cfg)

	if err := rabbitmq.InitRabbitmq(); err != nil {
		t.Fatal(err)
	}
}
