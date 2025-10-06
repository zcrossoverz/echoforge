package logging

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SanitizingEncoder wraps zapcore.Encoder to sanitize sensitive data
type SanitizingEncoder struct {
	zapcore.Encoder
	filter SensitiveFieldFilter
}

// NewSanitizingEncoder creates a new sanitizing encoder
func NewSanitizingEncoder(base zapcore.Encoder, filter SensitiveFieldFilter) *SanitizingEncoder {
	return &SanitizingEncoder{
		Encoder: base,
		filter:  filter,
	}
}

// Clone creates a copy of the encoder
func (s *SanitizingEncoder) Clone() zapcore.Encoder {
	return &SanitizingEncoder{
		Encoder: s.Encoder.Clone(),
		filter:  s.filter,
	}
}

// AddObject sanitizes object fields before encoding
func (s *SanitizingEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	// If the key itself is sensitive, sanitize it
	if s.filter.IsSensitive(key) {
		s.Encoder.AddString(key, "[REDACTED]")
		return nil
	}

	// For complex objects, we'll sanitize recursively
	sanitizedMarshaler := &SanitizingObjectMarshaler{
		original: marshaler,
		filter:   s.filter,
	}

	return s.Encoder.AddObject(key, sanitizedMarshaler)
}

// AddString sanitizes string values if the key is sensitive
func (s *SanitizingEncoder) AddString(key, val string) {
	sanitizedVal := s.filter.Sanitize(key, val)
	s.Encoder.AddString(key, sanitizedVal.(string))
}

// AddByteString sanitizes byte string values if the key is sensitive
func (s *SanitizingEncoder) AddByteString(key string, val []byte) {
	sanitizedVal := s.filter.Sanitize(key, string(val))
	s.Encoder.AddByteString(key, []byte(sanitizedVal.(string)))
}

// SanitizingObjectMarshaler wraps zapcore.ObjectMarshaler to sanitize fields
type SanitizingObjectMarshaler struct {
	original zapcore.ObjectMarshaler
	filter   SensitiveFieldFilter
}

// MarshalLogObject sanitizes object fields during marshaling
func (s *SanitizingObjectMarshaler) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	sanitizingEnc := &SanitizingObjectEncoder{
		ObjectEncoder: enc,
		filter:        s.filter,
	}
	return s.original.MarshalLogObject(sanitizingEnc)
}

// SanitizingObjectEncoder wraps zapcore.ObjectEncoder to sanitize fields
type SanitizingObjectEncoder struct {
	zapcore.ObjectEncoder
	filter SensitiveFieldFilter
}

// AddString in object encoder with sanitization
func (s *SanitizingObjectEncoder) AddString(key, val string) {
	sanitizedVal := s.filter.Sanitize(key, val)
	s.ObjectEncoder.AddString(key, sanitizedVal.(string))
}

// AddByteString in object encoder with sanitization
func (s *SanitizingObjectEncoder) AddByteString(key string, val []byte) {
	sanitizedVal := s.filter.Sanitize(key, string(val))
	s.ObjectEncoder.AddByteString(key, []byte(sanitizedVal.(string)))
}

// AddObject in object encoder with recursive sanitization
func (s *SanitizingObjectEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	if s.filter.IsSensitive(key) {
		s.ObjectEncoder.AddString(key, "[REDACTED]")
		return nil
	}

	sanitizedMarshaler := &SanitizingObjectMarshaler{
		original: marshaler,
		filter:   s.filter,
	}

	return s.ObjectEncoder.AddObject(key, sanitizedMarshaler)
}

// EnhancedSensitiveFieldFilter provides advanced sensitive data detection
type EnhancedSensitiveFieldFilter struct {
	sensitivePatterns []string
	sensitiveRegexes  []*regexp.Regexp
	urlRegex          *regexp.Regexp
	jwtRegex          *regexp.Regexp
}

// NewEnhancedSensitiveFieldFilter creates a new enhanced filter
func NewEnhancedSensitiveFieldFilter() *EnhancedSensitiveFieldFilter {
	patterns := []string{
		"password", "secret", "token", "dsn", "key", "auth", "credential",
		"passwd", "pwd", "authorization", "bearer", "api_key", "apikey",
		"private_key", "session", "cookie", "x-api-key", "x-auth-token",
	}

	// Compile regex patterns for more sophisticated detection
	regexes := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		regexes[i] = regexp.MustCompile(`(?i)` + pattern)
	}

	// URL regex to detect connection strings
	urlRegex := regexp.MustCompile(`\w+://[^:]+:[^@]+@`)

	// JWT token regex
	jwtRegex := regexp.MustCompile(`^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]*$`)

	return &EnhancedSensitiveFieldFilter{
		sensitivePatterns: patterns,
		sensitiveRegexes:  regexes,
		urlRegex:          urlRegex,
		jwtRegex:          jwtRegex,
	}
}

// IsSensitive checks if a field name or value is sensitive
func (f *EnhancedSensitiveFieldFilter) IsSensitive(fieldName string) bool {
	lowerFieldName := strings.ToLower(fieldName)

	// Check against patterns
	for _, regex := range f.sensitiveRegexes {
		if regex.MatchString(lowerFieldName) {
			return true
		}
	}

	return false
}

// Sanitize replaces sensitive values with [REDACTED] and handles special cases
func (f *EnhancedSensitiveFieldFilter) Sanitize(fieldName string, value interface{}) interface{} {
	// Check if field name is sensitive
	if f.IsSensitive(fieldName) {
		return "[REDACTED]"
	}

	// Check if value itself is sensitive (regardless of field name)
	strValue := fmt.Sprintf("%v", value)

	// Check for URL with credentials
	if f.urlRegex.MatchString(strValue) {
		return "[REDACTED]"
	}

	// Check for JWT tokens
	if f.jwtRegex.MatchString(strValue) && len(strValue) > 50 {
		return "[REDACTED]"
	}

	// Check for Bearer tokens
	if strings.HasPrefix(strings.ToLower(strValue), "bearer ") {
		return "[REDACTED]"
	}

	// For long strings that might be secrets (>30 chars, alphanumeric)
	if len(strValue) > 30 && isLikelySecret(strValue) {
		// Only sanitize if field name suggests it's sensitive
		lowerFieldName := strings.ToLower(fieldName)
		if strings.Contains(lowerFieldName, "secret") ||
			strings.Contains(lowerFieldName, "token") ||
			strings.Contains(lowerFieldName, "key") {
			return "[REDACTED]"
		}
	}

	return value
}

// isLikelySecret heuristic to detect if a string looks like a secret
func isLikelySecret(s string) bool {
	// Check if string is mostly alphanumeric (common in secrets/tokens)
	alphanumCount := 0
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			alphanumCount++
		}
	}

	// If >80% alphanumeric and long enough, might be a secret
	return float64(alphanumCount)/float64(len(s)) > 0.8 && len(s) > 20
}

// SanitizeMapRecursively sanitizes a map recursively
func SanitizeMapRecursively(data map[string]interface{}, filter SensitiveFieldFilter) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		result[key] = SanitizeValueRecursively(key, value, filter)
	}

	return result
}

// SanitizeValueRecursively sanitizes a value recursively based on its type
func SanitizeValueRecursively(key string, value interface{}, filter SensitiveFieldFilter) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case map[string]interface{}:
		// Recursively sanitize nested maps
		return SanitizeMapRecursively(v, filter)
	case []interface{}:
		// Sanitize array elements
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = SanitizeValueRecursively(fmt.Sprintf("%s[%d]", key, i), item, filter)
		}
		return result
	case string, []byte, int, int64, float64, bool:
		// Sanitize primitive values
		return filter.Sanitize(key, value)
	default:
		// For other types, try to sanitize using reflection
		return sanitizeWithReflection(key, value, filter)
	}
}

// sanitizeWithReflection uses reflection to sanitize struct fields
func sanitizeWithReflection(key string, value interface{}, filter SensitiveFieldFilter) interface{} {
	v := reflect.ValueOf(value)

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		// For structs, we'll convert to map and sanitize
		result := make(map[string]interface{})
		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			fieldValue := v.Field(i)

			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			fieldName := field.Name
			result[fieldName] = SanitizeValueRecursively(fieldName, fieldValue.Interface(), filter)
		}

		return result
	case reflect.Map:
		// Handle maps with string keys
		if v.Type().Key().Kind() == reflect.String {
			result := make(map[string]interface{})
			for _, mapKey := range v.MapKeys() {
				keyStr := mapKey.String()
				mapValue := v.MapIndex(mapKey)
				result[keyStr] = SanitizeValueRecursively(keyStr, mapValue.Interface(), filter)
			}
			return result
		}
	}

	// For other types, just apply the filter
	return filter.Sanitize(key, value)
}

// CreateSanitizingLogger creates a logger with sanitization enabled
func CreateSanitizingLogger(baseLogger *zap.Logger) *zap.Logger {
	// Get the base logger's core
	core := baseLogger.Core()

	// Create enhanced filter
	filter := NewEnhancedSensitiveFieldFilter()

	// Wrap the core with sanitization
	sanitizingCore := &SanitizingCore{
		Core:   core,
		filter: filter,
	}

	// Create new logger with sanitizing core
	return zap.New(sanitizingCore)
}

// SanitizingCore wraps zapcore.Core to sanitize all log entries
type SanitizingCore struct {
	zapcore.Core
	filter SensitiveFieldFilter
}

// With adds structured context to the Core
func (s *SanitizingCore) With(fields []zapcore.Field) zapcore.Core {
	// Sanitize fields before adding them
	sanitizedFields := make([]zapcore.Field, len(fields))
	for i, field := range fields {
		sanitizedFields[i] = s.sanitizeField(field)
	}

	return &SanitizingCore{
		Core:   s.Core.With(sanitizedFields),
		filter: s.filter,
	}
}

// Check determines whether the supplied Entry should be logged
func (s *SanitizingCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return s.Core.Check(ent, ce)
}

// Write serializes the Entry and any Fields supplied at the log site and writes them to their destination
func (s *SanitizingCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	// Sanitize fields before writing
	sanitizedFields := make([]zapcore.Field, len(fields))
	for i, field := range fields {
		sanitizedFields[i] = s.sanitizeField(field)
	}

	return s.Core.Write(ent, sanitizedFields)
}

// Sync flushes buffered logs
func (s *SanitizingCore) Sync() error {
	return s.Core.Sync()
}

// sanitizeField sanitizes a zapcore.Field
func (s *SanitizingCore) sanitizeField(field zapcore.Field) zapcore.Field {
	// Check if field key is sensitive
	if s.filter.IsSensitive(field.Key) {
		return zap.String(field.Key, "[REDACTED]")
	}

	// For string fields, check the value
	if field.Type == zapcore.StringType {
		sanitizedValue := s.filter.Sanitize(field.Key, field.String)
		return zap.String(field.Key, sanitizedValue.(string))
	}

	// For other types, we'll handle them as-is for now
	// Could be extended to handle more complex types
	return field
}
