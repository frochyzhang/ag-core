package ag_conf

// CommandArgsPrefix 命令行参数前缀取值key
const CommandArgsPrefix = "GS_ARGS_PREFIX"

const (
	//ConstPlaceholderPrefix Prefix for system property placeholders: "${".
	ConstPlaceholderPrefix = "${"
	//ConstPlaceholderSuffix Suffix for system property placeholders: "}".
	ConstPlaceholderSuffix = "}"
	//ConstValueSeparator Value separator for system property placeholders: ":".
	ConstValueSeparator = ":"
	// ConstEncryptKeyWords value start with keywords means has been encrypt
	ConstEncryptKeyWords = "{cipher}"
	// ConstEncrptSystemPropertiesSource system 加密后的环境变量
	ConstEncrptSystemPropertiesSource = "EncrptSystemPropertiesSource"
	// ConstEncrptCustomerPropertiesSource 客户端配置的数据解密资源
	ConstEncrptCustomerPropertiesSource = "EncrptCustomerPropertiesSource"
)
