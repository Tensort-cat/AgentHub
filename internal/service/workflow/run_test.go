package service

import (
	"context"
	"testing"

	"AgentHub/internal/dao"
	"AgentHub/internal/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestRunStatePersistsMessagesAfterInvoke(t *testing.T) {
	ctx := context.Background()
	state := newRuntimeState(9002, "hello")
	graph, err := buildGraph(ctx, nil, nil, state.SessionID, func(context.Context) *RuntimeState { return state })
	if err != nil {
		t.Fatal(err)
	}
	defer Sweep(state.SessionID)
	err = graph.AddLambdaNode("model", compose.InvokableLambda(func(_ context.Context, input string) (string, error) {
		return "reply: " + input, nil
	}), compose.WithStatePostHandler(func(_ context.Context, output string, s *RuntimeState) (string, error) {
		s.appendMessages(schema.AssistantMessage(output, nil))
		return output, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.AddEdge(compose.START, "model"); err != nil {
		t.Fatal(err)
	}
	if err := graph.AddEdge("model", compose.END); err != nil {
		t.Fatal(err)
	}
	runner, err := graph.Compile(ctx)
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Invoke(ctx, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if result != "reply: hello" {
		t.Fatalf("result = %q", result)
	}

	// Eino's local state is scoped to graph execution, not the caller context.
	if err := compose.ProcessState(ctx, func(context.Context, *RuntimeState) error { return nil }); err == nil {
		t.Fatal("expected caller context to have no graph state")
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
	var saved []*model.Message
	if err := db.Callback().Create().Before("gorm:create").Register("test:capture", func(tx *gorm.DB) {
		saved = tx.Statement.Dest.([]*model.Message)
	}); err != nil {
		t.Fatal(err)
	}
	if err := state.SaveMsg2DB(); err != nil {
		t.Fatal(err)
	}
	if len(saved) != 2 {
		t.Fatalf("saved %d messages, want 2", len(saved))
	}
	if saved[0].Content != "hello" || saved[1].Content != result {
		t.Fatalf("saved message contents = %q, %q", saved[0].Content, saved[1].Content)
	}
	for _, msg := range saved {
		if msg.SessionID != state.SessionID {
			t.Fatalf("session id = %d", msg.SessionID)
		}
	}
}
