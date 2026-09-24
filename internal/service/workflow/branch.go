package service

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// branchConfigVersion 用于以后扩展规则协议时拒绝不兼容的配置。
	branchConfigVersion = 1
	// branchDefaultRuleID 不属于普通规则，只在所有规则均未命中时使用。
	branchDefaultRuleID = "$default"

	branchOperatorEquals     = "equals"
	branchOperatorContains   = "contains"
	branchOperatorStartsWith = "starts_with"
	branchOperatorEndsWith   = "ends_with"
	branchOperatorRegex      = "regex"
)

var branchRuleIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// compiledBranchRule 缓存预编译的正则，避免工作流每次执行时重复编译。
type compiledBranchRule struct {
	rule  BranchRule
	regex *regexp.Regexp
}

type branchEvaluator struct {
	trimSpace bool
	rules     []compiledBranchRule
}

// newBranchEvaluator 同时完成规则配置校验和运行时求值器构建。
// 因此非法正则、重复 ID 或未知操作符会在工作流启动前失败。
func newBranchEvaluator(cfg BranchConfig) (*branchEvaluator, error) {
	if cfg.Version != branchConfigVersion {
		return nil, fmt.Errorf("branch config version must be %d", branchConfigVersion)
	}
	if len(cfg.Rules) == 0 {
		return nil, fmt.Errorf("branch rules cannot be empty")
	}

	seen := make(map[string]struct{}, len(cfg.Rules))
	compiled := make([]compiledBranchRule, 0, len(cfg.Rules))
	for index, rule := range cfg.Rules {
		if !branchRuleIDPattern.MatchString(rule.ID) {
			return nil, fmt.Errorf("branch rule %d has invalid id %q", index, rule.ID)
		}
		if rule.ID == branchDefaultRuleID {
			return nil, fmt.Errorf("branch rule id %q is reserved", branchDefaultRuleID)
		}
		if _, exists := seen[rule.ID]; exists {
			return nil, fmt.Errorf("duplicate branch rule id %q", rule.ID)
		}
		seen[rule.ID] = struct{}{}

		compiledRule := compiledBranchRule{rule: rule}
		switch rule.Operator {
		case branchOperatorEquals,
			branchOperatorContains,
			branchOperatorStartsWith,
			branchOperatorEndsWith:
		case branchOperatorRegex:
			pattern := rule.Value
			if !rule.CaseSensitive {
				pattern = "(?i:" + pattern + ")"
			}
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("branch rule %q has invalid regex: %w", rule.ID, err)
			}
			compiledRule.regex = re
		default:
			return nil, fmt.Errorf("branch rule %q has unsupported operator %q", rule.ID, rule.Operator)
		}

		compiled = append(compiled, compiledRule)
	}

	return &branchEvaluator{
		trimSpace: cfg.TrimSpace,
		rules:     compiled,
	}, nil
}

// Match 按配置顺序返回第一条命中的规则 ID；全部未命中时返回默认出口 ID。
// 此处只为匹配创建局部的规范化字符串，不会修改或替换传给下游节点的原始输入。
func (e *branchEvaluator) Match(input string) string {
	if e.trimSpace {
		input = strings.TrimSpace(input)
	}

	// 顺序遍历是 Branch “首条命中”语义的核心，不能改为 map。
	for _, compiled := range e.rules {
		rule := compiled.rule
		value := rule.Value
		candidate := input
		if e.trimSpace {
			value = strings.TrimSpace(value)
		}

		matched := false
		if rule.Operator == branchOperatorRegex {
			matched = compiled.regex.MatchString(candidate)
		} else {
			// 非正则规则统一转换大小写，使四种字符串操作保持一致语义。
			if !rule.CaseSensitive {
				candidate = strings.ToLower(candidate)
				value = strings.ToLower(value)
			}

			switch rule.Operator {
			case branchOperatorEquals:
				matched = candidate == value
			case branchOperatorContains:
				matched = strings.Contains(candidate, value)
			case branchOperatorStartsWith:
				matched = strings.HasPrefix(candidate, value)
			case branchOperatorEndsWith:
				matched = strings.HasSuffix(candidate, value)
			}
		}

		if matched {
			return rule.ID
		}
	}

	return branchDefaultRuleID
}
