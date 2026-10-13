package service

import (
	"testing"

	"AgentHub/internal/dao"
	"AgentHub/internal/model"
	"AgentHub/pkg/constant"
	model_enum "AgentHub/pkg/enum/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestModelProviderValidationAndPersistence(t *testing.T) {
	for _, provider := range []*model_enum.ModelProvider{nil, new(model_enum.ModelProvider(0)), new(model_enum.ModelProvider(3))} {
		if code, _ := Create("model", "https://example.com/v1", "test", 1, model_enum.CHAT, provider); code != constant.BadRequest {
			t.Fatalf("Create accepted invalid provider %v", provider)
		}
		if provider != nil {
			if code, _ := Update("1", "", "", "", provider); code != constant.BadRequest {
				t.Fatalf("Update accepted invalid provider %d", *provider)
			}
		}
	}
	db, err := gorm.Open(mysql.New(mysql.Config{SkipInitializeWithVersion: true}), &gorm.Config{
		DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	previousDB := dao.DB
	dao.DB = db
	t.Cleanup(func() { dao.DB = previousDB })
	var saved model.Model
	if err := db.Callback().Create().Before("gorm:create").Register("test:capture", func(tx *gorm.DB) {
		saved = *tx.Statement.Dest.(*model.Model)
	}); err != nil {
		t.Fatal(err)
	}
	var updated map[string]any
	if err := db.Callback().Update().Before("gorm:update").Register("test:capture", func(tx *gorm.DB) {
		updated = tx.Statement.Dest.(map[string]any)
	}); err != nil {
		t.Fatal(err)
	}
	for _, provider := range []model_enum.ModelProvider{model_enum.ModelProviderOpenAI, model_enum.ModelProviderArk} {
		if code, msg := Create("model", "https://example.com/v1", "test", 1, model_enum.CHAT, &provider); code != constant.Success {
			t.Fatalf("Create failed: %s", msg)
		}
		if saved.Provider != provider {
			t.Fatalf("saved provider = %d, want %d", saved.Provider, provider)
		}
		if code, msg := Update("1", "", "", "", &provider); code != constant.Success {
			t.Fatalf("Update failed: %s", msg)
		}
		if updated["provider"] != provider {
			t.Fatalf("updated provider = %v", updated["provider"])
		}
	}
	if code, msg := Update("1", "renamed", "", "", nil); code != constant.Success {
		t.Fatalf("Update failed: %s", msg)
	}
	if _, exists := updated["provider"]; exists {
		t.Fatal("omitted provider must not overwrite existing value")
	}
}
