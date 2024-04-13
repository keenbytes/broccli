package broccli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
)

// param represends a value and it is used for flags, args and environment variables.
// It has a name, alias, description, value that is shown when printing help, specific type (eg. TypeBool or TypeInt),
// If more than one value shoud be allowed, eg. '1,2,3' means "multiple integers" and the separator here is ','.
// Additional characters are used with type of TypeAlphanumeric to allow dots, underscore etc.  Hence, the value of that
// arg could be '._-'.
type param struct {
	name      string
	alias     string
	helpValue string
	desc      string
	valueType int64
	flags     int64
	options   paramOptions
}

// helpLine returns param usage info that is used when printing help.
func (p *param) helpLine() string {
	s := " "
	if p.alias == "" {
		s += " \t"
	} else {
		s += fmt.Sprintf(" -%s,\t", p.alias)
	}
	s += fmt.Sprintf(" --%s %s \t%s\n", p.name, p.helpValue, p.desc)
	return s
}

// ValidateValue takes value coming from --NAME and -ALIAS and validates it.
func (p *param) validateValue(v string) error {
	// empty, for every time except bool
	if p.valueType != TypeBool && (p.flags&IsRequired > 0) && v == "" {
		return errors.New("Missing value")
	}

	// string does not need any additional checks apart from the above one
	if p.valueType == TypeString {
		return nil
	}

	// if param is not required or not empty
	if !(p.flags&IsRequired > 0 || v != "") {
		return nil
	}

	// if flag is a file (regular file, directory, ...)
	if p.valueType == TypePathFile {
		fileInfo, err := os.Stat(v)
		if err != nil {
			if os.IsNotExist(err) {
				if p.flags&IsExistent > 0 {
					return errors.New(fmt.Sprintf("File %s does not exist", v))
				} else {
					return nil
				}
			} else {
				return errors.New(fmt.Sprintf("File %s cannot be opened for info", v))
			}
		}

		if p.flags&IsNotExistent > 0 {
			return errors.New(fmt.Sprintf("File %s already exists", v))
		}

		if !fileInfo.Mode().IsRegular() && (p.flags&IsRegularFile > 0) {
			return errors.New(fmt.Sprintf("Path %s is not a regular file", v))
		}

		if !fileInfo.Mode().IsDir() && (p.flags&IsDirectory > 0) {
			return errors.New(fmt.Sprintf("Path %s is not a directory", v))
		}

		if (p.flags&IsRegularFile > 0) && (p.flags&IsValidJSON > 0) {
			dat, err := os.ReadFile(v)
			if err != nil {
				return errors.New(fmt.Sprintf("File %s cannot be opened for JSON validation:", v))
			}
			if !json.Valid(dat) {
				return errors.New(fmt.Sprintf("File %s is not a valid JSON", v))
			}
		}

		return nil
	}

	// int, float, alphanumeric - single or many, separated by various chars
	var reType string
	var reValue string
	// set regexp part just for the type (eg. int, float, anum)
	switch p.valueType {
	case TypeInt:
		reType = "[0-9]+"
	case TypeFloat:
		reType = "[0-9]{1,16}\\.[0-9]{1,16}"
	case TypeAlphanumeric:
		reExtraChars := ""
		if p.flags&AllowUnderscore > 0 {
			reExtraChars += "_"
		}
		if p.flags&AllowDots > 0 {
			reExtraChars += "\\."
		}
		if p.flags&AllowHyphen > 0 {
			reExtraChars += "\\-"
		}
		reType = fmt.Sprintf("[0-9a-zA-Z%s]+", reExtraChars)
	default:
		return errors.New("Invalid type")
	}

	// create the final regexp depending on if single or many values are allowed
	if p.flags&AllowMultipleValues > 0 {
		var d string
		if p.flags&SeparatorColon > 0 {
			d = ":"
		} else if p.flags&SeparatorSemiColon > 0 {
			d = ";"
		} else {
			d = ","
		}
		reValue = "^" + reType + "(" + d + reType + ")*$"
	} else {
		reValue = "^" + reType + "$"
	}
	m, err := regexp.MatchString(reValue, v)
	if err != nil || !m {
		return errors.New("Invalid value")
	}

	return nil
}
