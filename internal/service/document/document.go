package service

import (
	"AgentHub/internal/dao"
	"AgentHub/internal/dto/response"
	"AgentHub/internal/model"
	"AgentHub/pkg/constant"
	document_enum "AgentHub/pkg/enum/document"
	"AgentHub/pkg/util"
	"AgentHub/pkg/zlog"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// ------------------------
// TODO: mq 相关, DocUpload应该要完全重构一下
// ------------------------
func DocUpload(kbID int64, file *multipart.FileHeader) (constant.Code, constant.Msg, response.DocUploadResp) {
	/*
		1. 先生成文档的元数据，把文件保存到指定目录
		2. 向量化存储到 Redis 中
		3. 文档元数据存储到 Mysql 中
		4. 返回数据给前端
	*/
	// 获取文件后缀
	fileExt := filepath.Ext(file.Filename)
	doc := model.Document{
		BaseModel: model.BaseModel{
			ID:        util.GenID(),
			CreatedAt: time.Now(),
		},
		KnowledgeBaseID: kbID,
		Name:            file.Filename,
		Type:            document_enum.DocType(fileExt),
		Status:          document_enum.Pending,
		Size:            file.Size,
	}
	// 将 pending 状态的 doc 先入库
	if err := dao.DB.Create(&doc).Error; err != nil {
		zlog.Error(err.Error())

		return constant.InternalServerError, constant.Error, response.DocUploadResp{}
	}

	// 更新文档状态为 parsing
	if err := UpdateDocStatus(doc.ID, document_enum.Parsing); err != nil {
		zlog.Error(err.Error())
		return constant.InternalServerError, constant.Error, response.DocUploadResp{}
	}
	// 将文件存储到指定目录
	kbIDStr := fmt.Sprint(kbID)
	docIDStr := fmt.Sprint(doc.ID)
	path := filepath.Join("static", "files", kbIDStr, docIDStr, file.Filename)
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		zlog.Error("打开文件失败", zap.Error(err))
		if err := UpdateDocStatus(doc.ID, document_enum.Error); err != nil {
			zlog.Error(err.Error())
		}
		return constant.InternalServerError, constant.Error, response.DocUploadResp{}
	}
	defer src.Close()
	dst, err := os.Create(path)
	if err != nil {
		zlog.Error("创建文件失败", zap.Error(err))
		if err := UpdateDocStatus(doc.ID, document_enum.Error); err != nil {
			zlog.Error(err.Error())
		}
		return constant.InternalServerError, constant.Error, response.DocUploadResp{}
	}

	if _, err := io.Copy(dst, src); err != nil {
		zlog.Error("拷贝文件失败", zap.Error(err))
		if err := UpdateDocStatus(doc.ID, document_enum.Error); err != nil {
			zlog.Error(err.Error())
		}
		return constant.InternalServerError, constant.Error, response.DocUploadResp{}
	}

	if err := dst.Close(); err != nil {
		zlog.Error("关闭文件失败", zap.Error(err))
		if err := UpdateDocStatus(doc.ID, document_enum.Error); err != nil {
			zlog.Error(err.Error())
		}
		return constant.InternalServerError, constant.Error, response.DocUploadResp{}
	}

	zlog.Info("文件存储成功")

	// 更新状态为 indexing
	if err := UpdateDocStatus(doc.ID, document_enum.Indexing); err != nil {
		zlog.Error(err.Error())
		return constant.InternalServerError, constant.Error, response.DocUploadResp{}
	}
	chunkNum, err := dao.IndexFile(path)
	if err != nil {
		zlog.Error("文件存储到 Redis 出错", zap.Error(err))

		// IndexFile 可能已经成功写入部分 Redis 数据，
		// 因此尝试清理该文档的 Redis 数据
		if cleanupErr := dao.DeleteDocumentIndex(
			kbIDStr,
			docIDStr,
		); cleanupErr != nil {
			zlog.Error(
				"清理 Redis 文档索引失败",
				zap.Error(cleanupErr),
			)
		}

		// 清理本地文件
		if removeErr := os.RemoveAll(filepath.Dir(path)); removeErr != nil {
			zlog.Error(
				"清理本地文件失败",
				zap.Error(removeErr),
			)
		}

		// 更新状态为 Error
		if err := UpdateDocStatus(doc.ID, document_enum.Error); err != nil {
			zlog.Error(err.Error())
		}

		return constant.InternalServerError, constant.Error, response.DocUploadResp{}
	}
	doc.ChunkNum = chunkNum
	doc.Status = document_enum.Success
	res := dao.DB.Save(&doc)
	if res.Error != nil {
		zlog.Error(
			"文件元数据存储到 MySQL 出错",
			zap.Error(res.Error),
		)

		// MySQL 写入失败，Redis 已经成功，因此也需要回滚 Redis
		if cleanupErr := dao.DeleteDocumentIndex(
			kbIDStr,
			docIDStr,
		); cleanupErr != nil {
			zlog.Error(
				"清理 Redis 文档索引失败",
				zap.Error(cleanupErr),
			)
		}

		// 清理本地文件
		if removeErr := os.RemoveAll(filepath.Dir(path)); removeErr != nil {
			zlog.Error(
				"清理本地文件失败",
				zap.Error(removeErr),
			)
		}

		if err := UpdateDocStatus(doc.ID, document_enum.Error); err != nil {
			zlog.Error(err.Error())
		}

		return constant.InternalServerError, constant.Error, response.DocUploadResp{}
	}

	return constant.Success, constant.Ok, response.DocUploadResp{
		ID:     doc.ID,
		Name:   doc.Name,
		Type:   doc.Type,
		Size:   doc.Size,
		Status: doc.Status,
	}
}

func UpdateDocStatus(id int64, status document_enum.DocStatus) error {
	return dao.DB.
		Model(&model.Document{}).
		Where("id = ?", id).
		Update("status", status).Error
}
