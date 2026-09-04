package service

import (
	"context"
	"fmt"

	"AgentHub/internal/dao"
	my_model "AgentHub/internal/model"

	"github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

type KBRetriever struct {
	Retriever retriever.Retriever
	Filter    string
}

func NewKBRetriever(
	ctx context.Context,
	kbID int64,
	topK int,
) (*KBRetriever, error) {

	if kbID <= 0 {
		return nil, fmt.Errorf("invalid kb_id: %d", kbID)
	}

	if topK < 0 {
		return nil, fmt.Errorf("invalid top_k: %d", topK)
	}

	// 查询知识库
	var kb my_model.KnowledgeBase
	if err := dao.DB.First(&kb, "id = ?", kbID).Error; err != nil {
		return nil, fmt.Errorf(
			"查询知识库失败 kb_id=%d: %w",
			kbID,
			err,
		)
	}

	// 查询 Embedder Model
	var mdl my_model.Model
	if err := dao.DB.First(&mdl, "id = ?", kb.EmbedderID).Error; err != nil {
		return nil, fmt.Errorf(
			"查询知识库 Embedder Model 失败 kb_id=%d embedder_id=%d: %w",
			kbID,
			kb.EmbedderID,
			err,
		)
	}

	// 创建 Embedder
	emb, err := dao.NewEmbedder(ctx, mdl)
	if err != nil {
		return nil, fmt.Errorf(
			"创建 Embedder 失败: %w",
			err,
		)
	}

	// 创建 Redis Retriever
	ret, err := redis.NewRetriever(ctx, &redis.RetrieverConfig{
		Client:      dao.RedisCli,
		Index:       "kb_vector_idx",
		TopK:        topK,
		Embedding:   emb,
		VectorField: "vector",
		ReturnFields: []string{
			"content",
			"kb_id",
			"doc_id",
		},
	})
	if err != nil {
		return nil, fmt.Errorf(
			"创建 Retriever 失败: %w",
			err,
		)
	}

	return &KBRetriever{
		Retriever: ret,
		Filter:    fmt.Sprintf("@kb_id:{%d}", kbID),
	}, nil
}

func (r *KBRetriever) Retrieve(
	ctx context.Context,
	query string,
	opts ...retriever.Option,
) ([]*schema.Document, error) {

	opts = append(
		opts,
		redis.WithFilterQuery(r.Filter),
	)

	return r.Retriever.Retrieve(ctx, query, opts...)
}
