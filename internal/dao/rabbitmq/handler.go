package rabbitmq

import (
	"AgentHub/internal/dao"
	"AgentHub/internal/model"
	document_enum "AgentHub/pkg/enum/document"
	"context"
	"encoding/json"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handleIndexMessage(
	ctx context.Context,
	delivery amqp.Delivery,
) error {

	var msg DocumentMessage

	if err := json.Unmarshal(
		delivery.Body,
		&msg,
	); err != nil {
		return err
	}

	// Pending -> Indexing
	if err := dao.DB.
		Model(&model.Document{}).
		Where("id = ?", msg.DocID).
		Update("status", document_enum.Indexing).Error; err != nil {
		return err
	}

	// 真正执行 RAG Index
	chunkNum, err := dao.IndexFile(msg.FilePath)
	dao.DB.
		Model(&model.Document{}).
		Where("id = ?", msg.DocID).
		Update("status", document_enum.Indexing)
	if err != nil {

		// Redis 可能已经写入了一部分数据
		_ = dao.DeleteDocumentIndex(
			strconv.FormatInt(msg.KbID, 10),
			strconv.FormatInt(msg.DocID, 10),
		)

		return err
	}

	// 更新 MySQL
	err = dao.DB.
		Model(&model.Document{}).
		Where("id = ?", msg.DocID).
		Updates(map[string]interface{}{
			"chunk_num": chunkNum,
			"status":    document_enum.Success,
		}).Error

	if err != nil {
		return err
	}

	return nil
}
