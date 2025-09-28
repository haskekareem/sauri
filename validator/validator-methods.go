package validator

import "fmt"

// ============================== User Methods ===========================

// ErrorReturner returns the validation errors.
func (v *Validation) ErrorReturner() ErrorContainer {
	return v.Errors
}

// DefaultRules defines a set of commonly used rules
func (v *Validation) DefaultRules() {
	v.Rules = map[string][]string{
		"username": {"required", "min:3", "max:20"},   // Username must be unique, min 3 characters, max 20
		"email":    {"required", "email"},             // Email must be valid and unique
		"password": {"required", "min:8", "password"}, // Password must be min 8 characters and confirmed
		"age":      {"required"},                      // Age must be numeric and at least 18
	}
}

// AddCustomValidation adds a custom validation function.
func (v *Validation) AddCustomValidation(name string, fn CustomValidationFunc) {
	v.CustomValidation[name] = fn
}

// SetCustomMessageForRule sets a custom error message for a field.
func (v *Validation) SetCustomMessageForRule(field, rule, msg string) {
	key := fmt.Sprintf("%s.%s", field, rule)
	v.CustomMessages[key] = msg
}

// SetAttributeAlias sets an attribute alias for a field.
func (v *Validation) SetAttributeAlias(fieldName, alias string) {
	v.AttributeAliases[fieldName] = alias
}

// AddCompositeRule adds a composite rule that combines multiple simple rules
func (v *Validation) AddCompositeRule(name string, rules []string) {
	v.Rules[name] = rules
}

// AddRule dynamically adds a rule to a field.
func (v *Validation) AddRule(field, rule string) {
	v.Rules[field] = append(v.Rules[field], rule)
}

// SetDependency sets a dependency in the DI container.
func (v *Validation) SetDependency(key string, value interface{}) {
	v.DIContainer[key] = value
}

// GetDependency retrieves a dependency from the DI container.
func (v *Validation) GetDependency(key string) (interface{}, bool) {
	value, exists := v.DIContainer[key]
	return value, exists
}
