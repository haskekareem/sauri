package validator

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"mime/multipart"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// CustomValidationFunc defines a function for custom validation.
type CustomValidationFunc func(value string, params ...string) bool

// ErrorContainer ValidatorErrors holds the validation errors.
type ErrorContainer map[string][]string

// Validation struct holds the data to be validated and the validation rules.
type Validation struct {
	Data             url.Values
	Errors           ErrorContainer
	Rules            map[string][]string
	CustomValidation map[string]CustomValidationFunc
	CustomMessages   map[string]string
	AttributeAliases map[string]string
	FileData         map[string]*multipart.FileHeader
	DIContainer      map[string]interface{}
	StopOnFirstFail  bool
	DBPool           struct {
		DBPoolSQL *sql.DB
		PoolPGX   *pgxpool.Pool
	}
}

// ============ main functionalities and features definitions ========

// Validate runs the validation rules on the data.
func (v *Validation) Validate() bool {

	// Iterate over each field and its associated rules
	for field, fieldRules := range v.Rules {
		// Get the value of the field
		value, exists := v.getFieldValue(field)
		if !exists {
			value = ""
		}
		// Apply each rule to the field's value
		for _, rule := range fieldRules {
			// Directly apply each rule — no conditional logic
			if !v.applyRule(field, value, rule) && v.StopOnFirstFail {
				break
			}
		}
	}
	return len(v.Errors) == 0
}

// getFieldValue retrieves the value of a field from the data.
func (v *Validation) getFieldValue(field string) (interface{}, bool) {
	// Check if the field is in the file data
	if fileValue, exists := v.FileData[field]; exists {
		return fileValue, true
	}
	// Check if the field is in the URL values
	if value, exists := v.Data[field]; exists && len(value) > 0 {
		return value[0], true
	}
	return nil, false
}

// addError adds an error message for a field.
func (v *Validation) addError(field, defaultMsg, rule string, params ...string) {
	key := fmt.Sprintf("%s.%s", field, rule)

	// Retrieve the custom message if it exists, otherwise use the default message
	message, ok := v.CustomMessages[key]
	if !ok {
		message = defaultMsg
	}
	// Use the attribute alias if it exists, otherwise use the field name
	alias := field
	if customAlias, exists := v.AttributeAliases[field]; exists {
		alias = customAlias
	}

	// Replace the first %s with the alias
	formattedMessage := strings.Replace(message, "%s", alias, 1)
	// Replace subsequent %s with params
	for _, param := range params {
		formattedMessage = strings.Replace(formattedMessage, "%s", param, 1)
	}

	// Store formatted message in the errors map
	v.Errors[field] = append(v.Errors[field], formattedMessage)
}

// applyRule applies a single validation rule to a field value.
func (v *Validation) applyRule(field string, value interface{}, rule string) bool {
	// Split the rule into its name and parameter
	parts := strings.Split(rule, ":")
	//The first part of the split rule, which represents the name of the validation rule (e.g., "min").
	ruleName := parts[0]

	//The second part of the split rule, if it exists, which represents the parameter for the rule
	// (e.g., "3" for "min:3").
	var ruleParams string
	if len(parts) > 1 {
		ruleParams = parts[1]
	}

	// Apply the appropriate validation logic based on the rule name
	//The switch statement checks the ruleName and applies the corresponding validation logic.
	switch ruleName {
	case "required":
		if strValue, ok := value.(string); ok && strValue == "" {
			v.addError(field, "This %s is required", ruleName)
			return false
		} else if fileValue, ok := value.(*multipart.FileHeader); ok && fileValue == nil {
			v.addError(field, "This %s is required", ruleName)
			return false
		}

	case "name_format":
		if strValue, ok := value.(string); ok {
			if !v.isValidNameFormat(strValue) {
				v.addError(field, "Must start with a letter and contain only letters and numbers", ruleName)
				return false
			}
		}

	case "email":
		if strValue, ok := value.(string); ok && !v.isValidEmail(strValue) {
			v.addError(field, "The %s field must be a valid email address", ruleName)
			return false
		}

	case "min":
		if strValue, ok := value.(string); ok && !v.isMin(strValue, ruleParams) {
			v.addError(field, "The %s field must be at least %s characters.", ruleName, ruleParams)
			return false
		}

	case "max":
		if strValue, ok := value.(string); ok && !v.isMax(strValue, ruleParams) {
			v.addError(field, "The %s field must not exceed %s characters.", ruleName, ruleParams)
			return false
		}

	case "regexp":
		if strValue, ok := value.(string); ok && !v.matchesRegex(strValue, ruleParams) {
			v.addError(field, "The %s field format is invalid", ruleName)
			return false
		}

	case "numeric":
		if strValue, ok := value.(string); ok && !v.isNumeric(strValue) {
			v.addError(field, "The %s field must be a number", ruleName)
			return false
		}

	case "date":
		if strValue, ok := value.(string); ok {
			if !v.isValidDateFormat(strValue) {
				v.addError(field, "The %s field must be a valid date in YYYY-MM-DD format", ruleName)
				return false
			}
		}

	case "confirmed":
		if strValue, ok := value.(string); ok && !v.isConfirmed(field, strValue) {
			v.addError(field, "The %s field confirmation does not match", ruleName)
			return false
		}

	case "unique":
		if strValue, ok := value.(string); ok && !v.isUnique(field, strValue, ruleParams) {
			v.addError(field, "The %s field must be unique", ruleName)
			return false
		}

	case "exists":
		if strValue, ok := value.(string); ok && !v.exists(field, strValue, ruleParams) {
			v.addError(field, "The %s field does not exist", ruleName)
			return false
		}

	case "file":
		if fileValue, ok := value.(*multipart.FileHeader); !ok && fileValue == nil {
			v.addError(field, "The %s field must be a valid file", ruleName)
			return false
		}

	case "mimes":
		if fileValue, ok := value.(*multipart.FileHeader); ok && !v.isValidMimeType(fileValue, ruleParams) {
			v.addError(field, "The %s field must be a file of type: %s", ruleName, ruleParams)
			return false
		}

	case "max_size":
		if fileValue, ok := value.(*multipart.FileHeader); ok && !v.isValidFileSize(fileValue, ruleParams) {
			v.addError(field, "The %s field must not exceed %s kilobytes", ruleName, ruleParams)
			return false
		}

	case "image-dimensions":
		dims := strings.Split(ruleParams, ",")
		minWidth, _ := strconv.Atoi(dims[0])
		minHeight, _ := strconv.Atoi(dims[0])
		if fileValue, ok := value.(*multipart.FileHeader); ok && !v.isValidImageDimensions(fileValue, minWidth, minHeight) {
			v.addError(field, "The %s must be at least %s pixels wide and %s pixels tall.", ruleName, strconv.Itoa(minWidth), strconv.Itoa(minHeight))
			return false
		}

	case "password":
		if strValue, ok := value.(string); ok {
			if !v.isMixedCase(strValue) {
				v.addError(field, "The %s field must contain both uppercase and lowercase letters", ruleName)
				return false
			}
			if !v.hasSymbol(strValue) {
				v.addError(field, "The %s field must contain at least one symbol", ruleName)
				return false
			}
			if !v.hasNumber(strValue) {
				v.addError(field, "The %s field must contain at least one number", ruleName)
				return false
			}
			if !v.hasLetter(strValue) {
				v.addError(field, "The %s field must contain at least one letter", ruleName)
				return false
			}
		}

	default:
		if customFunc, ok := v.CustomValidation[ruleName]; ok {
			if strValue, ok := value.(string); ok && !customFunc(strValue, ruleParams) {
				v.addError(field, "The %s field failed custom validation for rule %s", ruleName, ruleName, ruleParams)
				return false
			}
		}
	}

	return true
}

// ValidateDateOrder  checks if the end date is after the start date.
func (v *Validation) ValidateDateOrder(startField, endField string) {
	startDate, startExist := v.getFieldValue(startField)
	endDate, endExist := v.getFieldValue(endField)
	if !startExist || !endExist {
		return
	}

	start, err1 := time.Parse("2006-01-02", startDate.(string))
	end, err2 := time.Parse("2006-01-02", endDate.(string))
	if err1 != nil || err2 != nil {
		return
	}

	if end.Before(start) {
		v.addErrorForCrossFieldValidation(startField, endField, "date_order", "The %s must be before %s.")
	}

}
