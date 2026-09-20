package service

import (
	"AgentHub/internal/dao"
	"AgentHub/internal/dto/response"
	"AgentHub/internal/model"
	"AgentHub/pkg/constant"
	workflow_enum "AgentHub/pkg/enum/workflow"
	"AgentHub/pkg/util"
	"AgentHub/pkg/zlog"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/cloudwego/eino/compose"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ================================ workflow 元数据相关 ============================================
func Page(userID int64, page, size int) (constant.Code, constant.Msg, []response.WorkflowPageItem) {
	if page < 1 || size < 1 {
		return constant.BadRequest, "page and size must be greater than 0", nil
	}

	var workflows []model.Workflow
	res := dao.DB.Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(size).
		Offset((page - 1) * size).
		Find(&workflows)

	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return constant.Success, constant.Ok, nil
		}
		return constant.InternalServerError, constant.Error, nil
	}

	if len(workflows) == 0 {
		zlog.Debug("没有数据")
		return constant.Success, constant.Ok, nil
	}

	resp := make([]response.WorkflowPageItem, len(workflows))
	for i, workflow := range workflows {
		resp[i] = response.WorkflowPageItem{
			ID:          workflow.ID,
			Name:        workflow.Name,
			Description: workflow.Description,
			Status:      workflow.Status,
			CreatedAt:   workflow.CreatedAt,
		}
	}

	return constant.Success, constant.Ok, resp
}

func Create(userID int64, name, desc string) (constant.Code, constant.Msg) {
	workflow := model.Workflow{
		BaseModel: model.BaseModel{
			ID:        util.GenID(),
			CreatedAt: time.Now(),
		},
		UserID:      userID,
		Name:        name,
		Description: desc,
		Status:      workflow_enum.DRAFT, // 初始默认为草稿
	}

	res := dao.DB.Create(&workflow)
	if res.Error != nil {
		zlog.Error("创建工作流失败")
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}

func Detail(wfID string) (constant.Code, constant.Msg, *response.WorkflowDetailResp) {
	var workflowMeta model.Workflow
	resp := new(response.WorkflowDetailResp)

	// 获取工作流元数据
	res := dao.DB.First(&workflowMeta, "id = ?", wfID)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("工作流不存在")
			return constant.BadRequest, "工作流不存在", nil
		}
		zlog.Error("系统出错")
		return constant.InternalServerError, constant.Error, nil
	}

	resp.ID = workflowMeta.ID
	resp.Description = workflowMeta.Description
	resp.Name = workflowMeta.Name

	// 获取工作流的节点相关数据
	var nodes []model.WorkflowNode
	res = dao.DB.Model(&model.WorkflowNode{}).Where("workflow_id = ?", wfID).Find(&nodes)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("该工作流还未创建节点和边")
			return constant.Success, "该工作流还未创建节点和边", resp
		}
		zlog.Error("系统出错")
		return constant.InternalServerError, constant.Error, nil
	}

	nodesMeta := make([]response.NodeMetaData, len(nodes))
	for i, node := range nodes {
		nodesMeta[i] = response.NodeMetaData{
			ID:        node.ID,
			Name:      node.Name,
			Type:      node.Type,
			PositionX: node.PositionX,
			PositionY: node.PositionY,
			Config:    json.RawMessage(node.Config),
		}
	}
	resp.Nodes = nodesMeta

	// 获取工作流的边相关数据
	var edges []model.WorkflowEdge
	res = dao.DB.Model(&model.WorkflowEdge{}).Where("workflow_id = ?", wfID).Find(&edges)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("该工作流还未创建边")
			return constant.Success, "该工作流还未创建边", resp
		}
		zlog.Error("系统出错")
		return constant.InternalServerError, constant.Error, nil
	}

	edgesMeta := make([]response.EdgeMetaData, len(edges))
	for i, edge := range edges {
		edgesMeta[i] = response.EdgeMetaData{
			ID:           edge.ID,
			SourceNodeID: edge.SourceNodeID,
			TargetNodeID: edge.TargetNodeID,
		}
	}
	resp.Edges = edgesMeta

	return constant.Success, constant.Ok, resp
}

func isValidWorkflowStatus(status workflow_enum.WorkflowStatus) bool {
	switch status {
	case workflow_enum.DRAFT, workflow_enum.ACTIVE:
		return true
	default:
		return false
	}
}

func isValidNodeType(nodeType workflow_enum.WorkflowNodeType) bool {
	switch nodeType {
	case workflow_enum.ChatModel:
		return true
	case workflow_enum.ChatTemplate:
		return true
	case workflow_enum.Branch:
		return true
	case workflow_enum.Tool:
		return true
	case workflow_enum.Start:
		return true
	case workflow_enum.End:
		return true
	default:
		return false
	}
}

func Update(wfID, name, desc string, status *workflow_enum.WorkflowStatus) (constant.Code, constant.Msg) {
	if strings.TrimSpace(wfID) == "" {
		zlog.Info("更新工作流参数不合法: wfID 为空")
		return constant.BadRequest, "wfID 不能为空"
	}

	updateData := map[string]any{}
	if strings.TrimSpace(name) != "" {
		updateData["name"] = name
	}
	if strings.TrimSpace(desc) != "" {
		updateData["description"] = desc
	}
	if status != nil {
		if !isValidWorkflowStatus(*status) {
			zlog.Info("更新工作流参数不合法: status 非法")
			return constant.BadRequest, "status 必须为合法值"
		}
		updateData["status"] = *status
	}

	if len(updateData) == 0 {
		zlog.Info("更新工作流参数不合法: 未传入可更新字段")
		return constant.BadRequest, "至少传入一个要更新的字段"
	}

	var existing model.Workflow
	res := dao.DB.Select("id").First(&existing, "id = ?", wfID)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("工作流不存在")
			return constant.BadRequest, "工作流不存在"
		}
		zlog.Error("查询工作流失败: " + err.Error())
		return constant.InternalServerError, constant.Error
	}

	res = dao.DB.Model(&model.Workflow{}).Where("id = ?", wfID).Updates(updateData)
	if err := res.Error; err != nil {
		zlog.Error("更新工作流失败: " + err.Error())
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}

func Delete(wfID string) (constant.Code, constant.Msg) {
	if strings.TrimSpace(wfID) == "" {
		zlog.Info("更新工作流参数不合法: wfID 为空")
		return constant.BadRequest, "wfID 不能为空"
	}

	deletedAt := gorm.DeletedAt{Time: time.Now(), Valid: true}

	err := dao.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Workflow{}).
			Where("id = ?", wfID).
			Where("deleted_at IS NULL").
			Update("deleted_at", deletedAt)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		if err := tx.Model(&model.WorkflowNode{}).
			Where("workflow_id = ?", wfID).
			Where("deleted_at IS NULL").
			Update("deleted_at", deletedAt).Error; err != nil {
			return err
		}

		return tx.Model(&model.WorkflowEdge{}).
			Where("workflow_id = ?", wfID).
			Where("deleted_at IS NULL").
			Update("deleted_at", deletedAt).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("工作流不存在")
			return constant.BadRequest, "工作流不存在"
		}
		zlog.Error("删除工作流出错")
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}

// 启动工作流 (核心)
func Run(
	ctx context.Context,
	wfID int64,
	input string,
	userID int64,
) (constant.Code, constant.Msg, response.WorkflowResult) {

	// 1. 加载工作流
	workflow, nodes, edges, err := loadWorkflow(wfID)
	if err != nil {
		return constant.InternalServerError, constant.Error, response.WorkflowResult{}
	}

	// 2. 校验工作流
	if err := ValidateWorkflow(workflow, nodes, edges, userID); err != nil {
		return constant.BadRequest, constant.Msg(err.Error()), response.WorkflowResult{}
	}

	// 3. 创建 Session
	session := createSession(wfID, input)

	if err := dao.DB.Create(&session).Error; err != nil {
		return constant.InternalServerError, constant.Error, response.WorkflowResult{}
	}

	// 结束后清理现场
	defer Sweep(session.ID)

	// 4. 创建 Graph
	graph, err := BuildGraph(
		ctx,
		input,
		nodes,
		edges,
		session.ID,
	)
	if err != nil {
		return constant.InternalServerError, constant.Error, response.WorkflowResult{}
	}

	// 5. 编译
	runner, err := graph.Compile(ctx)
	if err != nil {
		zlog.Error(err.Error())
		return constant.InternalServerError, constant.Error, response.WorkflowResult{}
	}

	// 6. 执行
	result, err := runner.Invoke(ctx, input)
	if err != nil {
		zlog.Error(err.Error())
		return constant.InternalServerError, constant.Error, response.WorkflowResult{}
	}

	// 7. 保存本次运行产生的所有消息
	err = compose.ProcessState(ctx, func(
		_ context.Context,
		state *RuntimeState,
	) error {
		return state.SaveMsg2DB()
	})
	if err != nil {
		zlog.Error(err.Error())
		return constant.InternalServerError, constant.Error, response.WorkflowResult{}
	}

	// 8. 返回
	return constant.Success, constant.Ok, response.WorkflowResult{
		SessionID: session.ID,
		Content:   result,
	}
}

func ValidateWorkflow(
	wf model.Workflow,
	nodes []model.WorkflowNode,
	edges []model.WorkflowEdge,
	userID int64,
) error {

	if err := validateNodes(wf, nodes); err != nil {
		return err
	}

	if err := validateEdges(wf, nodes, edges); err != nil {
		return err
	}

	if err := validateGraph(nodes, edges); err != nil {
		return err
	}

	if err := validateNodeConfigs(nodes, userID); err != nil {
		return err
	}

	return nil
}

func loadWorkflow(wfID int64) (model.Workflow, []model.WorkflowNode, []model.WorkflowEdge, error) {
	// 取工作流
	var wf model.Workflow
	if err := dao.DB.First(&wf, "id = ?", wfID).Error; err != nil {
		return model.Workflow{}, nil, nil, err
	}

	// 取节点
	var nodes []model.WorkflowNode
	if err := dao.DB.
		Model(&model.WorkflowNode{}).
		Where("workflow_id = ?", wfID).
		Find(&nodes).Error; err != nil {
		return model.Workflow{}, nil, nil, err
	}

	// 取边
	var edges []model.WorkflowEdge
	if err := dao.DB.
		Model(&model.WorkflowEdge{}).
		Where("workflow_id = ?", wfID).
		Find(&edges).Error; err != nil {
		return model.Workflow{}, nil, nil, err
	}

	return wf, nodes, edges, nil
}

func createSession(wfID int64, input string) model.Session {
	session := model.Session{
		BaseModel: model.BaseModel{
			ID:        util.GenID(),
			CreatedAt: time.Now(),
		},
		WorkflowID: wfID,
	}

	if len(input) <= 15 {
		session.Title = input
	} else {
		session.Title = input[:15]
	}

	return session
}

func saveMessages(msgs []*model.Message) error {
	return dao.DB.Create(msgs).Error
}

// ======================== Node 相关 =========================================
func NodeCreate(wfID int64, name string,
	nodeType *workflow_enum.WorkflowNodeType,
	x, y *int,
	config json.RawMessage) (constant.Code, constant.Msg) {
	node := model.WorkflowNode{
		BaseModel: model.BaseModel{
			ID:        util.GenID(),
			CreatedAt: time.Now(),
		},
		WorkflowID: wfID,
		Name:       name,
		Type:       *nodeType,
		PositionX:  *x,
		PositionY:  *y,
		Config:     datatypes.JSON(config),
	}

	res := dao.DB.Create(&node)
	if err := res.Error; err != nil {
		zlog.Error(err.Error())
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}

func NodeUpdate(nodeID, name string, nodeType *workflow_enum.WorkflowNodeType, x, y *int, config json.RawMessage) (constant.Code, constant.Msg) {
	if strings.TrimSpace(nodeID) == "" {
		zlog.Info("参数不合法: nodeID 为空")
		return constant.BadRequest, "nodeID 不能为空"
	}

	updateData := map[string]any{}
	if strings.TrimSpace(name) != "" {
		updateData["name"] = name
	}

	if nodeType != nil {
		if !isValidNodeType(*nodeType) {
			zlog.Info("参数不合法: nodeType 非法")
			return constant.BadRequest, "nodeType 必须为合法值"
		}
		updateData["type"] = *nodeType
	}

	if x != nil {
		updateData["position_x"] = *x
	}
	if y != nil {
		updateData["position_y"] = *y
	}

	if config != nil {
		updateData["config"] = config
	}

	if len(updateData) == 0 {
		zlog.Info("更新工作流参数不合法: 未传入可更新字段")
		return constant.BadRequest, "至少传入一个要更新的字段"
	}

	var existing model.WorkflowNode
	res := dao.DB.First(&existing, "id = ?", nodeID)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("节点不存在")
			return constant.BadRequest, "节点不存在"
		}
		zlog.Error("查询工作流失败: " + err.Error())
		return constant.InternalServerError, constant.Error
	}

	res = dao.DB.Model(&model.WorkflowNode{}).Where("id = ?", nodeID).Updates(updateData)
	if err := res.Error; err != nil {
		zlog.Error("更新工作流失败: " + err.Error())
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}

func NodeDelete(nodeID string) (constant.Code, constant.Msg) {
	if strings.TrimSpace(nodeID) == "" {
		zlog.Info("参数不合法: nodeID 为空")
		return constant.BadRequest, "nodeID 不能为空"
	}

	deletedAt := gorm.DeletedAt{Time: time.Now(), Valid: true}

	err := dao.DB.Transaction(func(tx *gorm.DB) error {
		// 删除节点
		res := tx.Model(&model.WorkflowNode{}).
			Where("id = ?", nodeID).
			Where("deleted_at IS NULL").
			Update("deleted_at", deletedAt)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		// 删除相关边
		return tx.Model(&model.WorkflowEdge{}).
			Where("deleted_at IS NULL").
			Where("source_node_id = ? OR target_node_id = ?", nodeID, nodeID).
			Update("deleted_at", deletedAt).Error
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("节点不存在")
			return constant.BadRequest, "节点不存在"
		}
		zlog.Error(err.Error())
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}

// =============================== Edge 相关 ===================================================
func EdgeCreate(wfID int64, sourceNodeID, targetNodeID int64) (constant.Code, constant.Msg) {
	edge := model.WorkflowEdge{
		BaseModel: model.BaseModel{
			ID:        util.GenID(),
			CreatedAt: time.Now(),
		},
		WorkflowID:   wfID,
		SourceNodeID: sourceNodeID,
		TargetNodeID: targetNodeID,
	}

	res := dao.DB.Create(&edge)
	if err := res.Error; err != nil {
		zlog.Error(err.Error())
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}

func EdgeDelete(edgeID string) (constant.Code, constant.Msg) {
	if strings.TrimSpace(edgeID) == "" {
		zlog.Info("参数不合法: edgeID 为空")
		return constant.BadRequest, "edgeID 不能为空"
	}

	deletedAt := gorm.DeletedAt{Time: time.Now(), Valid: true}

	err := dao.DB.Transaction(func(tx *gorm.DB) error {
		// 删除边
		res := tx.Model(&model.WorkflowEdge{}).
			Where("id = ?", edgeID).
			Where("deleted_at IS NULL").
			Update("deleted_at", deletedAt)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("边不存在")
			return constant.BadRequest, "边不存在"
		}
		zlog.Error(err.Error())
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}
