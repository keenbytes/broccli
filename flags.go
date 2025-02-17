package broccli

const (
	// Command param type
	_         = iota
	ParamFlag = iota + 1
	ParamArg
	ParamEnvVar
)

const (
	// Value types
	TypeString = iota * 1
	TypeBool
	TypeInt
	TypeFloat
	TypeAlphanumeric
	TypePathFile
)

const (
	_ = 1 << iota
	// Validation
	IsRequired    // IsRequired means that the value is required
	IsExistent    // IsExistent is used with TypePathFile and requires file to exist
	IsNotExistent // IsNotExistent is used with TypePathFile and requires file not to exist
	IsDirectory   // IsDirectory is used with TypePathFile and requires file to be a directory
	IsRegularFile // IsRegularFile is used with TypePathFile and requires file to be a regular file
	IsValidJSON   // IsValidJSON is used with TypeString or TypePathFile with RegularFile to check if the contents are a valid JSON

	AllowDots       // AllowDots can be used only with TypeAlphanumeric and additionally allows flag to have dots.
	AllowUnderscore // AllowUnderscore can be used only with TypeAlphanumeric and additionally allows flag to have underscore chars.
	AllowHyphen     // AllowHyphen can be used only with TypeAlphanumeric and additionally allows flag to have hyphen chars.

	// AllowMultipleValues allows param to have more than one value separated by comma by default.
	// For example: AllowMany with TypeInt allows values like: 123 or 123,455,666 or 12,222
	// AllowMany works only with TypeInt, TypeFloat and TypeAlphanumeric.
	AllowMultipleValues
	SeparatorColon     // SeparatorColon works with AllowMultipleValues and sets colon to be the value separator, instead of colon.
	SeparatorSemiColon // SeparatorSemiColon works with AllowMultipleValues and sets semi-colon to be the value separator.
)
