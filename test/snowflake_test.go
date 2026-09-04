package test

import (
	"AgentHub/pkg/util"
	"testing"
)

func TestGenID(t *testing.T) {
	id := util.GenID()
	t.Log(id)
}
