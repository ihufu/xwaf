package model

// RuleStats 规则统计信息
type RuleStats struct {
	TotalRules      int64 `json:"total_rules"`       // 规则总数
	EnabledRules    int64 `json:"enabled_rules"`     // 启用的规则数
	DisabledRules   int64 `json:"disabled_rules"`    // 禁用的规则数
	HighRiskRules   int64 `json:"high_risk_rules"`   // 高风险规则数
	MediumRiskRules int64 `json:"medium_risk_rules"` // 中风险规则数
	LowRiskRules    int64 `json:"low_risk_rules"`    // 低风险规则数

	// 按类型统计
	SQLiRules   int64 `json:"sqli_rules"`   // SQL注入规则数
	XSSRules    int64 `json:"xss_rules"`    // XSS规则数
	CCRules     int64 `json:"cc_rules"`     // CC规则数
	CustomRules int64 `json:"custom_rules"` // 自定义规则数

	// 最近统计
	LastDayMatches   int64 `json:"last_day_matches"`   // 最近一天匹配次数
	LastWeekMatches  int64 `json:"last_week_matches"`  // 最近一周匹配次数
	LastMonthMatches int64 `json:"last_month_matches"` // 最近一月匹配次数
	TotalMatches     int64 `json:"total_matches"`      // 总匹配次数
	LastUpdatedAt    int64 `json:"last_updated_at"`    // 最后更新时间
}
