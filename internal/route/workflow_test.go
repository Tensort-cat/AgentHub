package route

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWorkflowRunEventsRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	InitWorkflow(engine.Group("/api/v1/workflows"))

	for _, route := range engine.Routes() {
		if route.Method == "GET" && route.Path == "/api/v1/workflows/run/:task_id/events" {
			return
		}
	}

	t.Fatal("workflow run events route is not registered")
}
