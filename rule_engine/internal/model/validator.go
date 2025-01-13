package model

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/xwaf/rule_engine/internal/errors"
)

// InputValidator 输入验证器
type InputValidator struct {
	maxLength      int
	allowedChars   string
	disallowedStrs []string
	sanitize       bool
}

// NewInputValidator 创建输入验证器
func NewInputValidator() *InputValidator {
	return &InputValidator{
		maxLength:    4096,                                                                                   // 默认最大长度
		allowedChars: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~:/?#[]@!$&'()*+,;=", // RFC 3986
		disallowedStrs: []string{
			"../", "..\\", // 目录遍历
			"<!--", "-->", // HTML注释
			"<script", "</script>", // XSS
			"javascript:", "vbscript:", "data:", // 危险协议
			"union", "select", "insert", "update", "delete", "drop", // SQL关键字
			"eval(", "exec(", "system(", // 命令执行
			"base64_decode(", "base64_encode(", // 编码函数
			"document.cookie", "document.domain", // DOM操作
		},
		sanitize: true,
	}
}

// SetMaxLength 设置最大长度
func (v *InputValidator) SetMaxLength(length int) *InputValidator {
	v.maxLength = length
	return v
}

// SetAllowedChars 设置允许的字符
func (v *InputValidator) SetAllowedChars(chars string) *InputValidator {
	v.allowedChars = chars
	return v
}

// SetDisallowedStrings 设置禁止的字符串
func (v *InputValidator) SetDisallowedStrings(strs []string) *InputValidator {
	v.disallowedStrs = strs
	return v
}

// SetSanitize 设置是否进行清理
func (v *InputValidator) SetSanitize(sanitize bool) *InputValidator {
	v.sanitize = sanitize
	return v
}

// ValidateString 验证字符串
func (v *InputValidator) ValidateString(input string) (string, error) {
	// 检查长度
	if len(input) > v.maxLength {
		return "", errors.NewError(errors.ErrValidation, fmt.Sprintf("输入长度超过限制: %d > %d", len(input), v.maxLength))
	}

	// 检查字符
	for _, c := range input {
		if !strings.ContainsRune(v.allowedChars, c) && !unicode.IsSpace(c) {
			if v.sanitize {
				input = strings.Map(func(r rune) rune {
					if strings.ContainsRune(v.allowedChars, r) || unicode.IsSpace(r) {
						return r
					}
					return -1
				}, input)
			} else {
				return "", errors.NewError(errors.ErrValidation, fmt.Sprintf("包含非法字符: %c", c))
			}
		}
	}

	// 检查禁止的字符串
	for _, str := range v.disallowedStrs {
		if strings.Contains(strings.ToLower(input), strings.ToLower(str)) {
			if v.sanitize {
				input = strings.ReplaceAll(strings.ToLower(input), strings.ToLower(str), "")
			} else {
				return "", errors.NewError(errors.ErrValidation, fmt.Sprintf("包含禁止的字符串: %s", str))
			}
		}
	}

	return input, nil
}

// ValidateURL 验证URL
func (v *InputValidator) ValidateURL(input string) (string, error) {
	// 解析URL
	u, err := url.Parse(input)
	if err != nil {
		return "", errors.NewError(errors.ErrValidation, fmt.Sprintf("无效的url格式: %v", err))
	}

	// 检查协议
	if u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.NewError(errors.ErrValidation, fmt.Sprintf("不支持的url协议: %s", u.Scheme))
	}

	// 验证路径
	if strings.Contains(u.Path, "..") {
		return "", errors.NewError(errors.ErrValidation, "url路径包含目录遍历")
	}

	// 验证查询参数
	for key, values := range u.Query() {
		for _, value := range values {
			if _, err := v.ValidateString(key); err != nil {
				return "", errors.NewError(errors.ErrValidation, fmt.Sprintf("url参数名称无效: %v", err))
			}
			if _, err := v.ValidateString(value); err != nil {
				return "", errors.NewError(errors.ErrValidation, fmt.Sprintf("url参数值无效: %v", err))
			}
		}
	}

	return input, nil
}

// ValidateHeaders 验证HTTP头
func (v *InputValidator) ValidateHeaders(headers map[string]string) (map[string]string, error) {
	result := make(map[string]string)

	for key, value := range headers {
		// 验证头名称
		if !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(key) {
			return nil, errors.NewError(errors.ErrValidation, fmt.Sprintf("无效的header名称: %s", key))
		}

		// 验证头值
		cleanValue, err := v.ValidateString(value)
		if err != nil {
			return nil, errors.NewError(errors.ErrValidation, fmt.Sprintf("header值无效 [%s]: %v", key, err))
		}

		result[key] = cleanValue
	}

	return result, nil
}

// ValidateJSON 验证JSON数据
func (v *InputValidator) ValidateJSON(input map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range input {
		// 验证键
		if _, err := v.ValidateString(key); err != nil {
			return nil, errors.NewError(errors.ErrValidation, fmt.Sprintf("json键无效: %v", err))
		}

		// 验证值
		switch val := value.(type) {
		case string:
			cleanVal, err := v.ValidateString(val)
			if err != nil {
				return nil, errors.NewError(errors.ErrValidation, fmt.Sprintf("json值无效 [%s]: %v", key, err))
			}
			result[key] = cleanVal
		case map[string]interface{}:
			cleanVal, err := v.ValidateJSON(val)
			if err != nil {
				return nil, errors.NewError(errors.ErrValidation, fmt.Sprintf("json对象无效 [%s]: %v", key, err))
			}
			result[key] = cleanVal
		case []interface{}:
			cleanVal, err := v.ValidateJSONArray(val)
			if err != nil {
				return nil, errors.NewError(errors.ErrValidation, fmt.Sprintf("json数组无效 [%s]: %v", key, err))
			}
			result[key] = cleanVal
		default:
			result[key] = value
		}
	}

	return result, nil
}

// ValidateJSONArray 验证JSON数组
func (v *InputValidator) ValidateJSONArray(input []interface{}) ([]interface{}, error) {
	result := make([]interface{}, 0, len(input))

	for i, value := range input {
		switch val := value.(type) {
		case string:
			cleanVal, err := v.ValidateString(val)
			if err != nil {
				return nil, errors.NewError(errors.ErrValidation, fmt.Sprintf("数组元素[%d]无效: %v", i, err))
			}
			result = append(result, cleanVal)
		case map[string]interface{}:
			cleanVal, err := v.ValidateJSON(val)
			if err != nil {
				return nil, errors.NewError(errors.ErrValidation, fmt.Sprintf("数组元素[%d]无效: %v", i, err))
			}
			result = append(result, cleanVal)
		case []interface{}:
			cleanVal, err := v.ValidateJSONArray(val)
			if err != nil {
				return nil, errors.NewError(errors.ErrValidation, fmt.Sprintf("数组元素[%d]无效: %v", i, err))
			}
			result = append(result, cleanVal)
		default:
			result = append(result, value)
		}
	}

	return result, nil
}
