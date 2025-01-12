package service

import "errors"

var (
	// ErrInvalidRuleVariable 无效的规则变量类型
	ErrInvalidRuleVariable = errors.New("invalid rule variable")
	// ErrInvalidRuleType 无效的规则类型
	ErrInvalidRuleType = errors.New("invalid rule type")
	// ErrInvalidActionType 无效的动作类型
	ErrInvalidActionType = errors.New("invalid action type")
	// ErrInvalidSeverity 无效的风险级别
	ErrInvalidSeverity = errors.New("invalid severity")
	// ErrInvalidRulesOperation 无效的规则组合操作
	ErrInvalidRulesOperation = errors.New("invalid rules operation")
)
