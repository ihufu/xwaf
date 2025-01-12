package model

import "time"

// IPListType IP 名单类型
type IPListType string

const (
	IPListTypeWhite IPListType = "white" // 白名单
	IPListTypeBlack IPListType = "black" // 黑名单
)

// BlockType 封禁类型
type BlockType string

const (
	BlockTypePermanent BlockType = "permanent" // 永久封禁
	BlockTypeTemporary BlockType = "temporary" // 临时封禁
)

// IPRule IP 规则
type IPRule struct {
	ID          int64      `json:"id" db:"id"`                   // 规则ID
	IP          string     `json:"ip" db:"ip"`                   // IP地址
	IPType      IPListType `json:"ip_type" db:"ip_type"`         // IP类型（黑/白名单）
	BlockType   BlockType  `json:"block_type" db:"block_type"`   // 封禁类型
	ExpireTime  time.Time  `json:"expire_time" db:"expire_time"` // 过期时间（临时封禁用）
	Description string     `json:"description" db:"description"` // 规则描述
	CreatedBy   int64      `json:"created_by" db:"created_by"`   // 创建者
	UpdatedBy   int64      `json:"updated_by" db:"updated_by"`   // 更新者
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`   // 创建时间
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`   // 更新时间
}

// IPRuleQuery IP 规则查询参数
type IPRuleQuery struct {
	Page      int        `form:"page"`       // 页码
	Size      int        `form:"size"`       // 每页大小
	Keyword   string     `form:"keyword"`    // 关键词
	IPType    IPListType `form:"ip_type"`    // IP类型
	BlockType BlockType  `form:"block_type"` // 封禁类型
}

// Validate 验证 IP 规则
func (r *IPRule) Validate() error {
	// TODO: 实现 IP 规则验证
	return nil
}
