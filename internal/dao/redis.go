package dao

import (
	"AgentHub/internal/config"
	"AgentHub/internal/model"
	model_enum "AgentHub/pkg/enum/model"
	"AgentHub/pkg/zlog"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"

	redisIndexer "github.com/cloudwego/eino-ext/components/indexer/redis"
	redisRetriever "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var RedisCli *redis.Client
var ctx = context.Background()

func InitRedisCli(ctx context.Context) error {
	cfg := config.Cfg.RedisConfig
	cli := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		Protocol: 2, // // FT.SEARCH 必需
	})
	cli.Options().UnstableResp3 = true // 向量搜索必需

	pong := cli.Ping(ctx)
	if _, err := pong.Result(); err != nil {
		return err
	}
	RedisCli = cli

	// 初始化向量索引
	if err := initRedisIndex(
		"kb_vector_idx",
		"rag:kb:",
	); err != nil {
		return err
	}

	return nil
}

func CloseRedis() error {
	if RedisCli == nil {
		return nil
	}

	return RedisCli.Close()
}

// 初始化 Redis 中的向量索引
func initRedisIndex(indexName, prefix string) error {
	// 检查索引是否存在
	_, err := RedisCli.Do(ctx, "FT.INFO", indexName).Result()
	if err == nil {
		zlog.Info("Redis 索引已存在，跳过创建",
			zap.String("index", indexName),
		)
		return nil
	}

	// FT.INFO 返回 Unknown index name，说明索引不存在，可以继续创建
	if !strings.Contains(err.Error(), "Unknown index name") {
		return fmt.Errorf("检查 Redis 索引失败: %w", err)
	}

	dimension := config.Cfg.RedisConfig.Dimension
	if dimension <= 0 {
		return fmt.Errorf("invalid vector dimension: %d", dimension)
	}

	zlog.Info("正在创建 Redis 向量索引...",
		zap.String("index", indexName),
		zap.String("prefix", prefix),
		zap.Int("dimension", dimension),
	)

	createArgs := []any{
		"FT.CREATE", indexName,
		"ON", "HASH",
		"PREFIX", "1", prefix,
		"SCHEMA",

		// 文档内容
		"content", "TEXT",

		// 知识库 ID，用于过滤
		"kb_id", "TAG",

		// 文档 ID，后续可以用于文档级过滤
		"doc_id", "TAG",

		// 向量
		"vector", "VECTOR", "FLAT",
		"6",
		"TYPE", "FLOAT32",
		"DIM", dimension,
		"DISTANCE_METRIC", "COSINE",
	}

	if err := RedisCli.Do(ctx, createArgs...).Err(); err != nil {
		zlog.Error("创建 Redis 向量索引失败",
			zap.Error(err),
		)
		return err
	}

	zlog.Info("Redis 向量索引创建成功",
		zap.String("index", indexName),
	)

	return nil
}

func Get(key string) (string, error) {
	return RedisCli.Get(ctx, key).Result()
}

func SetEx(key string, value string, expiration time.Duration) error {
	_, err := RedisCli.Set(ctx, key, value, expiration).Result()
	if err != nil {
		zlog.Error("failed to set redis key", zap.Error(err))
		return err
	}
	return nil
}

// 将文件索引到 Redis 并返回切片数量和错误
func IndexFile(path string) (int, error) {
	/*
		1. 解析文件路径，获取 kbID、docID 和文件名
		2. 读取文件内容
		3. 将文档分 chunk
		4. 查询知识库及其使用的 Embedder Model
		5. 创建 Embedder
		6. 创建 Redis Indexer
		7. 将 chunks 向量化并存入 Redis
		8. 将文档元数据存入 Redis
	*/

	// 解析文件路径
	// path: static/files/kbID/docID/xxx.md
	filePath := filepath.Clean(path)

	fileName := filepath.Base(filePath)
	docDir := filepath.Dir(filePath)
	docID := filepath.Base(docDir)
	kbDir := filepath.Dir(docDir)
	kbID := filepath.Base(kbDir)

	// 读取文件内容
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		zlog.Debug(err.Error())
		return -1, err
	}

	// 将文件内容封装成 schema.Document
	doc := &schema.Document{
		ID:      docID,
		Content: string(bytes),
		MetaData: map[string]any{
			"name": fileName,
		},
	}

	// 初始化 transformer
	trans, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   1500,
		OverlapSize: 300,
	})
	if err != nil {
		zlog.Debug(err.Error())
		return -1, err
	}

	// 切分文档
	transformedDocs, err := trans.Transform(ctx, []*schema.Document{doc})
	if err != nil {
		zlog.Debug(err.Error())
		return -1, err
	}

	chunkNum := len(transformedDocs)

	// 设置每个 chunk 的 ID
	for i := range transformedDocs {
		transformedDocs[i].ID = strconv.Itoa(i)
	}

	// 查询知识库
	var kb model.KnowledgeBase
	if err := DB.First(&kb, "id = ?", kbID).Error; err != nil {
		zlog.Debug(err.Error())
		return -1, err
	}

	// 查询知识库使用的 Embedder Model
	var embedderModel model.Model
	if err := DB.First(&embedderModel, "id = ?", kb.EmbedderID).Error; err != nil {
		zlog.Debug(err.Error())
		return -1, err
	}

	// 创建 Embedder
	emb, err := NewEmbedder(ctx, embedderModel)
	if err != nil {
		zlog.Debug(err.Error())
		return -1, err
	}

	// 创建 Redis Indexer
	prefix := fmt.Sprintf(
		"rag:kb:%s:doc:%s:chunk:",
		kbID,
		docID,
	)

	indexer, err := redisIndexer.NewIndexer(ctx, &redisIndexer.IndexerConfig{
		Client:    RedisCli,
		KeyPrefix: prefix,
		BatchSize: 10,
		Embedding: emb,

		DocumentToHashes: func(
			ctx context.Context,
			doc *schema.Document,
		) (*redisIndexer.Hashes, error) {

			return &redisIndexer.Hashes{
				Key: doc.ID,

				Field2Value: map[string]redisIndexer.FieldValue{
					"content": {
						Value:    doc.Content,
						EmbedKey: "vector",
					},
					"kb_id": {
						Value: kbID,
					},
					"doc_id": {
						Value: docID,
					},
				},
			}, nil
		},
	})
	if err != nil {
		zlog.Debug(err.Error())
		return -1, err
	}

	// 将 chunks 向量化并存入 Redis
	_, err = indexer.Store(ctx, transformedDocs)
	if err != nil {
		zlog.Debug(err.Error())
		return -1, err
	}

	// 所有 chunks 索引成功后，再将文档元数据存入 Redis
	key4Meta := fmt.Sprintf(
		"kb:%s:doc:%s",
		kbID,
		docID,
	)

	_, err = RedisCli.HMSet(
		ctx,
		key4Meta,
		"name", fileName,
		"chunk_num", chunkNum,
	).Result()
	if err != nil {
		zlog.Debug(err.Error())
		return -1, err
	}

	return chunkNum, nil
}

func RetrieveDocChunks(kbID string, query string) ([]*schema.Document, error) {

	// 查询知识库
	var kb model.KnowledgeBase
	if err := DB.First(&kb, "id = ?", kbID).Error; err != nil {
		return nil, err
	}

	// 查询知识库所使用的 Embedder Model
	var mdl model.Model
	if err := DB.First(&mdl, "id = ?", kb.EmbedderID).Error; err != nil {
		return nil, err
	}

	// 创建 Embedder
	emb, err := NewEmbedder(ctx, mdl)
	if err != nil {
		return nil, err
	}

	// 创建 Retriever
	retriever, err := redisRetriever.NewRetriever(ctx, &redisRetriever.RetrieverConfig{
		Client:      RedisCli,
		Index:       "kb_vector_idx",
		TopK:        5,
		Embedding:   emb,
		VectorField: "vector",
		ReturnFields: []string{
			"content",
			"kb_id",
			"doc_id",
		},
	})
	if err != nil {
		return nil, err
	}

	// 只检索指定知识库
	filter := fmt.Sprintf("@kb_id:{%s}", kbID)
	chunks, err := retriever.Retrieve(
		ctx,
		query,
		redisRetriever.WithFilterQuery(filter),
	)
	if err != nil {
		return nil, err
	}

	return chunks, nil
}

func NewEmbedder(ctx context.Context, model model.Model) (embedding.Embedder, error) {
	switch model.Provider {

	case model_enum.ModelProviderOpenAI:
		return openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
			APIKey: model.APIKey,
			Model:  model.Name,
		})

	case model_enum.ModelProviderArk:
		apiType := ark.APITypeMultiModal
		return ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
			APIKey:  model.APIKey,
			Model:   model.Name,
			APIType: &apiType,
		})

	default:
		return nil, fmt.Errorf("不支持的 Embedder Provider: %v", model.Provider)
	}
}

// 删除文档在 Redis 中的所有索引数据
func DeleteDocumentIndex(kbID, docID string) error {
	// 删除文档元数据
	metaKey := fmt.Sprintf(
		"kb:%s:doc:%s",
		kbID,
		docID,
	)

	if err := RedisCli.Del(ctx, metaKey).Err(); err != nil {
		zlog.Error(
			"删除 Redis 文档元数据失败",
			zap.String("key", metaKey),
			zap.Error(err),
		)
		return err
	}

	// 删除文档对应的所有 chunk
	prefix := fmt.Sprintf(
		"rag:kb:%s:doc:%s:chunk:",
		kbID,
		docID,
	)

	var cursor uint64

	for {
		keys, nextCursor, err := RedisCli.Scan(
			ctx,
			cursor,
			prefix+"*",
			100,
		).Result()
		if err != nil {
			zlog.Error(
				"扫描 Redis 文档 chunk 失败",
				zap.String("prefix", prefix),
				zap.Error(err),
			)
			return err
		}

		if len(keys) > 0 {
			if err := RedisCli.Del(ctx, keys...).Err(); err != nil {
				zlog.Error(
					"删除 Redis 文档 chunk 失败",
					zap.String("prefix", prefix),
					zap.Error(err),
				)
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
