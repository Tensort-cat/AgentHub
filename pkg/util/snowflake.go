package util

import (
	"AgentHub/pkg/constant"
	"log"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func init() {
	gen, err := snowflake.NewNode(constant.NUM)
	if err != nil {
		log.Fatal(err)
	}

	node = gen
}

func GenID() int64 {
	return node.Generate().Int64()
}
