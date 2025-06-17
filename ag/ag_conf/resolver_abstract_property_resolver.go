package ag_conf

import (
	"fmt"
)

// AbstractPropertyResolver 属性解析器抽象实现
type AbstractPropertyResolver struct {
	PlaceholderPrefix                    string
	PlaceholderSuffix                    string
	ValueSeparator                       string
	IgnoreUnresolvableNestedPlaceholders bool
	// 必须的属性key集合
	RequiredProperties []string

	GetProperty func(key string) string // 由具体实现提供
}

/*
 == 实现PropertyResolver ==
*/

// ContainsProperty impl IPropertyResolver.ContainsProperty
func (apr *AbstractPropertyResolver) ContainsProperty(key string) bool {
	v := apr.GetProperty(key)
	return v != ""
}

// GetPropertyDefault impl IPropertyResolver.GetPropertyDefault
func (apr *AbstractPropertyResolver) GetPropertyDefault(key string, defaultValue string) string {
	v := apr.GetProperty(key)
	if v == "" {
		return defaultValue
	}
	return v
}

// GetRequiredProperty impl IPropertyResolver.GetRequiredProperty
func (apr *AbstractPropertyResolver) GetRequiredProperty(key string) (string, error) {
	v := apr.GetProperty(key)
	if v == "" {
		return "", fmt.Errorf("Required key [%s] not found", key)
	}
	return v, nil
}

// ResolvePlaceholders impl IPropertyResolver.ResolvePlaceholders
func (apr *AbstractPropertyResolver) ResolvePlaceholders(text string) string {
	//TODO 需要实现PropertyPlaceholderHelper，以支持占位符解析
	return ""
}

// ResolveRequiredPlaceholders impl IPropertyResolver.ResolveRequiredPlaceholders
func (apr *AbstractPropertyResolver) ResolveRequiredPlaceholders(text string) (string, error) {
	//TODO 需要实现PropertyPlaceholderHelper，以支持占位符解析
	return "", nil
}

/*
 == 实现ConfigurablePeopertyResolver ==
*/

// SetPlaceholdPrefix impl IConfigurablePeopertyResolver.SetPlaceholdPrefix
func (apr *AbstractPropertyResolver) SetPlaceholdPrefix(placeholderPrefic string) {
	apr.PlaceholderPrefix = placeholderPrefic
}

// SetPlaceholdSuffix impl IConfigurablePeopertyResolver.SetPlaceholdSuffix
func (apr *AbstractPropertyResolver) SetPlaceholdSuffix(placeholderSuffix string) {
	apr.PlaceholderSuffix = placeholderSuffix
}

// SetValueSeparator impl IConfigurablePeopertyResolver.SetValueSeparator
func (apr *AbstractPropertyResolver) SetValueSeparator(valueSeparator string) {
	apr.ValueSeparator = valueSeparator
}

// SetIgnoreUnresolvableNestedPlaceholders impl IConfigurablePeopertyResolver.SetIgnoreUnresolvableNestedPlaceholders
func (apr *AbstractPropertyResolver) SetIgnoreUnresolvableNestedPlaceholders(ignoreUnresolvableNestedPlaceholders bool) {
	apr.IgnoreUnresolvableNestedPlaceholders = ignoreUnresolvableNestedPlaceholders
}

// SetRequiredProperties 设置必须的属性key集合
func (apr *AbstractPropertyResolver) SetRequiredProperties(requiredProperties ...string) {
	for _, key := range requiredProperties {
		apr.RequiredProperties = append(apr.RequiredProperties, key)
	}
}

// ValidateRequiredProperties 校验必须的属性是否存在，如果不存在则返回错误
func (apr *AbstractPropertyResolver) ValidateRequiredProperties() error {
	missingKeys := []string{}
	for _, key := range apr.RequiredProperties {
		if apr.GetProperty(key) == "" {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf("The following properties were declared as required but could not be resolved: %v", missingKeys)
	}

	return nil
}
