package broccli

const (
	// Command param type
	ParamFlag   = 1
	ParamArg    = 2
	ParamEnvVar = 3

	// Value types
	TypeString       = 0
	TypeBool         = 1
	TypeInt          = 2
	TypeFloat        = 3
	TypeAlphanumeric = 4
	TypePathFile     = 5

	// Validation
	// IsRequired means that the value is required
	IsRequired = 1
	// IsExistent is used with TypePathFile and requires file to exist
	IsExistent = 2
	// IsNotExistent is used with TypePathFile and requires file not to exist
	IsNotExistent = 4
	// IsDirectory is used with TypePathFile and requires file to be a directory
	IsDirectory = 8
	// IsRegularFile is used with TypePathFile and requires file to be a regular file
	IsRegularFile = 16
	// IsValidJSON is used with TypeString or TypePathFile with RegularFile to check if the contents are a valid JSON
	IsValidJSON = 32

	// AllowDots can be used only with TypeAlphanumeric and additionally allows flag to have dots.
	AllowDots = 2048
	// AllowUnderscore can be used only with TypeAlphanumeric and additionally allows flag to have underscore chars.
	AllowUnderscore = 4096
	// AllowHyphen can be used only with TypeAlphanumeric and additionally allows flag to have hyphen chars.
	AllowHyphen = 8192

	// AllowMultipleValues allows param to have more than one value separated by comma by default.
	// For example: AllowMany with TypeInt allows values like: 123 or 123,455,666 or 12,222
	// AllowMany works only with TypeInt, TypeFloat and TypeAlphanumeric.
	AllowMultipleValues = 16384
	// SeparatorColon works with AllowMultipleValues and sets colon to be the value separator, instead of colon.
	SeparatorColon = 32768
	// SeparatorSemiColon works with AllowMultipleValues and sets semi-colon to be the value separator.
	SeparatorSemiColon = 65536
)
