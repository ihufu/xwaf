package model

import (
	"sort"
	"strings"
	"time"

	"github.com/xwaf/rule_engine/internal/errors"
)

// SQLToken SQL Token
type SQLToken struct {
	Type  SQLTokenType
	Value string
	Pos   int
}

// SQLLexer SQL词法分析器
type SQLLexer struct {
	input string
	pos   int
	start int
}

// NewSQLLexer 创建SQL词法分析器
func NewSQLLexer(input string) *SQLLexer {
	return &SQLLexer{
		input: input,
		pos:   0,
		start: 0,
	}
}

// SQLInjectionDetector SQL注入检测器
type SQLInjectionDetector struct {
	lexer *SQLLexer
}

// NewSQLInjectionDetector 创建SQL注入检测器
func NewSQLInjectionDetector() *SQLInjectionDetector {
	return &SQLInjectionDetector{}
}

// DetectInjection 检测SQL注入
func (d *SQLInjectionDetector) DetectInjection(input string) (bool, error) {
	if input == "" {
		return false, errors.NewError(errors.ErrValidation, "输入不能为空")
	}

	d.lexer = NewSQLLexer(input)
	tokens := make([]*SQLToken, 0)

	// 使用词法分析器获取所有token
	for {
		token := d.lexer.NextToken()
		if token.Type == SQLTokenEOF {
			break
		}
		tokens = append(tokens, token)
	}

	score := 0
	reasons := make([]string, 0)

	// 检查语法异常
	if d.hasUnionSelect(tokens) {
		score += 3
		reasons = append(reasons, "存在UNION SELECT注入")
	}

	// 检查永真条件
	if d.hasAlwaysTrueCondition(tokens) {
		score += 2
		reasons = append(reasons, "存在永真条件")
	}

	// 检查注释截断
	if d.hasCommentTruncation(tokens) {
		score += 2
		reasons = append(reasons, "存在注释截断")
	}

	// 根据得分判断是否存在注入
	if score >= 3 {
		return true, errors.NewError(errors.ErrSQLInjection, strings.Join(reasons, "；"))
	}

	return false, nil
}

// hasUnionSelect 检查是否存在UNION SELECT注入
func (d *SQLInjectionDetector) hasUnionSelect(tokens []*SQLToken) bool {
	for i := 0; i < len(tokens)-1; i++ {
		if tokens[i].Type == SQLTokenKeyword && strings.EqualFold(tokens[i].Value, "UNION") {
			if i+1 < len(tokens) && tokens[i+1].Type == SQLTokenKeyword && strings.EqualFold(tokens[i+1].Value, "SELECT") {
				return true
			}
		}
	}
	return false
}

// hasAlwaysTrueCondition 检查是否存在永真条件
func (d *SQLInjectionDetector) hasAlwaysTrueCondition(tokens []*SQLToken) bool {
	for i := 0; i < len(tokens)-2; i++ {
		if tokens[i].Type == SQLTokenNumber && tokens[i+1].Type == SQLTokenOperator && tokens[i+1].Value == "=" {
			if tokens[i+2].Type == SQLTokenNumber && tokens[i].Value == tokens[i+2].Value {
				return true
			}
		}
	}
	return false
}

// hasCommentTruncation 检查是否存在注释截断
func (d *SQLInjectionDetector) hasCommentTruncation(tokens []*SQLToken) bool {
	for i := 0; i < len(tokens); i++ {
		if tokens[i].Type == SQLTokenComment {
			if i > 0 && tokens[i-1].Type == SQLTokenOperator {
				return true
			}
		}
	}
	return false
}

// RuleVersion 规则版本
type RuleVersion struct {
	ID         int64     `json:"id" db:"id"`
	RuleID     int64     `json:"rule_id" db:"rule_id"`
	Version    int64     `json:"version" db:"version"`
	Hash       string    `json:"hash" db:"hash"`
	Content    string    `json:"content" db:"content"`
	ChangeType string    `json:"change_type" db:"change_type"`
	Status     string    `json:"status" db:"status"`
	CreatedBy  int64     `json:"created_by" db:"created_by"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// RuleVersionQuery 规则版本查询条件
type RuleVersionQuery struct {
	RuleID     int64  `json:"rule_id"`
	Version    int64  `json:"version"`
	ChangeType string `json:"change_type"`
	Status     string `json:"status"`
}

// RuleSyncLog 规则同步日志
type RuleSyncLog struct {
	ID        int64     `json:"id" db:"id"`
	RuleID    int64     `json:"rule_id" db:"rule_id"`
	Version   int64     `json:"version" db:"version"`
	Status    string    `json:"status" db:"status"`
	Message   string    `json:"message" db:"message"`
	SyncType  string    `json:"sync_type" db:"sync_type"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RuleAuditLog 规则审计日志
type RuleAuditLog struct {
	ID        int64     `json:"id" db:"id"`
	RuleID    int64     `json:"rule_id" db:"rule_id"`
	Action    string    `json:"action" db:"action"`
	Operator  string    `json:"operator" db:"operator"`
	OldValue  string    `json:"old_value" db:"old_value"`
	NewValue  string    `json:"new_value" db:"new_value"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RuleMatchStat 规则匹配统计
type RuleMatchStat struct {
	RuleID    int64             `json:"rule_id"`    // 规则ID
	StartTime time.Time         `json:"start_time"` // 开始时间
	EndTime   time.Time         `json:"end_time"`   // 结束时间
	Total     int64             `json:"total"`      // 总匹配次数
	Timeline  []*RuleMatchPoint `json:"timeline"`   // 时间线
}

// RuleMatchPoint 规则匹配点
type RuleMatchPoint struct {
	Timestamp int64 `json:"timestamp"` // 时间戳
	Count     int64 `json:"count"`     // 匹配次数
}

// CheckRequest 检查请求
type CheckRequest struct {
	ClientIP  string            `json:"client_ip"`
	URI       string            `json:"uri"`
	Headers   map[string]string `json:"headers"`
	Args      map[string]string `json:"args"`
	Body      string            `json:"body"`
	Method    string            `json:"method"`
	RuleTypes []RuleType        `json:"rule_types"`
}

// Validate 验证请求参数
func (r *CheckRequest) Validate() error {
	if r.URI == "" {
		return errors.NewError(errors.ErrValidation, "uri不能为空")
	}

	if r.ClientIP == "" {
		return errors.NewError(errors.ErrValidation, "client_ip不能为空")
	}

	if r.Method == "" {
		return errors.NewError(errors.ErrValidation, "method不能为空")
	}

	if len(r.RuleTypes) == 0 {
		return errors.NewError(errors.ErrValidation, "rule_types不能为空")
	}

	return nil
}

// RuleUpdateType 规则更新类型
type RuleUpdateType string

const (
	RuleUpdateTypeCreate   RuleUpdateType = "create"   // 创建规则
	RuleUpdateTypeUpdate   RuleUpdateType = "update"   // 更新规则
	RuleUpdateTypeDelete   RuleUpdateType = "delete"   // 删除规则
	RuleUpdateTypeRollback RuleUpdateType = "rollback" // 回滚规则
)

// RuleDiff 规则变更记录
type RuleDiff struct {
	RuleID     int64          `json:"rule_id"`     // 规则ID
	Name       string         `json:"name"`        // 规则名称
	Pattern    string         `json:"pattern"`     // 规则模式
	Action     ActionType     `json:"action"`      // 规则动作
	Status     StatusType     `json:"status"`      // 规则状态
	Version    int64          `json:"version"`     // 规则版本
	UpdateType RuleUpdateType `json:"update_type"` // 更新类型
	UpdateTime time.Time      `json:"update_time"` // 更新时间
}

// RuleUpdateEvent 规则更新事件
type RuleUpdateEvent struct {
	ID        int64          `json:"id"`
	Version   int64          `json:"version"`
	Action    RuleUpdateType `json:"action"`
	RuleDiffs []*RuleDiff    `json:"rule_diffs"`
	CreatedAt time.Time      `json:"created_at"`
}

// RuleMatch 规则匹配结果
type RuleMatch struct {
	Rule       *Rule   `json:"rule"`
	MatchedStr string  `json:"matched_str"`
	Position   int     `json:"position"`
	Score      float64 `json:"score"`
}

// RuleGroup 规则组
type RuleGroup struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Status      int       `json:"status" db:"status"`
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	UpdatedBy   int64     `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// RuleTestCase 规则测试用例
type RuleTestCase struct {
	ID        int64         `json:"id" db:"id"`
	RuleID    int64         `json:"rule_id" db:"rule_id"`
	Request   *CheckRequest `json:"request" db:"request"`
	Input     string        `json:"input" db:"input"`
	Expected  bool          `json:"expected" db:"expected"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
}

// RuleTestResult 规则测试结果
type RuleTestResult struct {
	TestCase    *RuleTestCase `json:"test_case"`
	IsMatch     bool          `json:"is_match"`
	MatchResult *RuleMatch    `json:"match_result"`
	Duration    time.Duration `json:"duration"`
	Error       string        `json:"error"`
}

// CheckResult 检查结果
type CheckResult struct {
	Matched     bool       `json:"matched"`      // 是否匹配
	Action      ActionType `json:"action"`       // 动作
	MatchedRule *Rule      `json:"matched_rule"` // 匹配的规则
	Message     string     `json:"message"`      // 消息
}

// NextToken 获取下一个Token
func (l *SQLLexer) NextToken() *SQLToken {
	l.skipWhitespace()
	l.start = l.pos

	if l.pos >= len(l.input) {
		return &SQLToken{Type: SQLTokenEOF}
	}

	c := l.input[l.pos]

	switch {
	case isLetter(c):
		return l.scanIdentifier()
	case isDigit(c):
		return l.scanNumber()
	case c == '\'' || c == '"':
		return l.scanString()
	case c == '-' && l.peek() == '-':
		return l.scanComment()
	case c == '/' && l.peek() == '*':
		return l.scanMultiLineComment()
	case c == '+' || c == '-' || c == '*' || c == '/' || c == '=' || c == '<' || c == '>' || c == '!':
		return l.scanOperator()
	default:
		l.pos++
		return &SQLToken{Type: SQLTokenError, Value: string(c), Pos: l.start}
	}
}

// 辅助函数
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (l *SQLLexer) peek() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *SQLLexer) skipWhitespace() {
	for l.pos < len(l.input) && (l.input[l.pos] == ' ' || l.input[l.pos] == '\t' || l.input[l.pos] == '\n' || l.input[l.pos] == '\r') {
		l.pos++
	}
}

func (l *SQLLexer) scanIdentifier() *SQLToken {
	for l.pos < len(l.input) && (isLetter(l.input[l.pos]) || isDigit(l.input[l.pos])) {
		l.pos++
	}
	return &SQLToken{
		Type:  SQLTokenIdentifier,
		Value: l.input[l.start:l.pos],
		Pos:   l.start,
	}
}

func (l *SQLLexer) scanNumber() *SQLToken {
	for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
		l.pos++
	}
	return &SQLToken{
		Type:  SQLTokenNumber,
		Value: l.input[l.start:l.pos],
		Pos:   l.start,
	}
}

func (l *SQLLexer) scanString() *SQLToken {
	quote := l.input[l.pos]
	l.pos++ // 跳过开始引号
	for l.pos < len(l.input) {
		if l.input[l.pos] == quote && l.input[l.pos-1] != '\\' {
			l.pos++ // 跳过结束引号
			break
		}
		l.pos++
	}
	return &SQLToken{
		Type:  SQLTokenString,
		Value: l.input[l.start:l.pos],
		Pos:   l.start,
	}
}

func (l *SQLLexer) scanComment() *SQLToken {
	l.pos += 2 // 跳过 --
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.pos++
	}
	return &SQLToken{
		Type:  SQLTokenComment,
		Value: l.input[l.start:l.pos],
		Pos:   l.start,
	}
}

func (l *SQLLexer) scanMultiLineComment() *SQLToken {
	l.pos += 2 // 跳过 /*
	for l.pos < len(l.input)-1 {
		if l.input[l.pos] == '*' && l.input[l.pos+1] == '/' {
			l.pos += 2 // 跳过 */
			break
		}
		l.pos++
	}
	return &SQLToken{
		Type:  SQLTokenComment,
		Value: l.input[l.start:l.pos],
		Pos:   l.start,
	}
}

func (l *SQLLexer) scanOperator() *SQLToken {
	l.pos++
	return &SQLToken{
		Type:  SQLTokenOperator,
		Value: l.input[l.start:l.pos],
		Pos:   l.start,
	}
}

// SortRuleMatchesByPriority 按规则优先级排序匹配结果
func SortRuleMatchesByPriority(matches []*RuleMatch) {
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Rule.Priority > matches[j].Rule.Priority
	})
}
