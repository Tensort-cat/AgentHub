package service

import (
	"AgentHub/internal/dao"
	"AgentHub/internal/dao/rabbitmq"
	"AgentHub/internal/dto/response"
	"AgentHub/pkg/constant"
	workflow_enum "AgentHub/pkg/enum/workflow"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const workflowRunStateTTL = 24 * time.Hour

type workflowRunState struct {
	TaskID     string                          `json:"task_id"`
	WorkflowID int64                           `json:"workflow_id"`
	UserID     int64                           `json:"user_id"`
	SessionID  int64                           `json:"session_id,omitempty"`
	Status     workflow_enum.WorkflowRunStatus `json:"status"`
	Result     string                          `json:"result,omitempty"`
	Error      string                          `json:"error,omitempty"`
	UpdatedAt  time.Time                       `json:"updated_at"`
}

func workflowRunStateKey(taskID string) string {
	return fmt.Sprintf("workflow:run:%s", taskID)
}

func saveWorkflowRunState(ctx context.Context, state workflowRunState) error {
	if state.UpdatedAt.IsZero() {
		state.UpdatedAt = time.Now().UTC()
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return dao.RedisCli.Set(ctx, workflowRunStateKey(state.TaskID), data, workflowRunStateTTL).Err()
}

func HandleRunResult(ctx context.Context, result rabbitmq.WorkflowRunResult) error {
	return saveWorkflowRunState(ctx, workflowRunState{
		TaskID:     result.TaskID,
		WorkflowID: result.WorkflowID,
		UserID:     result.UserID,
		SessionID:  result.SessionID,
		Status:     result.Status,
		Result:     result.Result,
		Error:      result.Error,
		UpdatedAt:  result.FinishedAt,
	})
}

func GetRunState(
	ctx context.Context,
	taskID string,
	userID int64,
) (constant.Code, constant.Msg, response.WorkflowRunState) {
	data, err := dao.RedisCli.Get(ctx, workflowRunStateKey(taskID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return constant.NotFound, "运行任务不存在或已过期", response.WorkflowRunState{}
		}
		return constant.InternalServerError, constant.Error, response.WorkflowRunState{}
	}

	var state workflowRunState
	if err := json.Unmarshal(data, &state); err != nil {
		return constant.InternalServerError, constant.Error, response.WorkflowRunState{}
	}
	if state.UserID != userID {
		return constant.Forbidden, "无权访问该运行任务", response.WorkflowRunState{}
	}

	return constant.Success, constant.Ok, response.WorkflowRunState{
		TaskID:    state.TaskID,
		Status:    state.Status,
		SessionID: state.SessionID,
		Result:    state.Result,
		Error:     state.Error,
		UpdatedAt: state.UpdatedAt,
	}
}
