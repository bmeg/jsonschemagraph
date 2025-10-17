package compile

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

func ValidateFhirDateTime(v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("value is not a string type")
	}
	if !strings.Contains(s, "T") {
		return validatePartialFhirDate(s, "FHIR date-time")
	}
	if _, err := time.Parse(time.RFC3339Nano, s); err != nil {
		return fmt.Errorf("value '%s' is not a valid FHIR date-time: %w", s, err)
	}
	return nil
}

func validatePartialFhirDate(s, typeName string) error {

	switch len(s) {
	case 4: // YYYY
		if _, err := time.Parse("2006", s); err != nil {
			return fmt.Errorf("value '%s' is not a valid partial %s (YYYY): %w", s, typeName, err)
		}
	case 7: // YYYY-MM
		if _, err := time.Parse("2006-01", s); err != nil {
			return fmt.Errorf("value '%s' is not a valid partial %s (YYYY-MM): %w", s, typeName, err)
		}
	case 10: // YYYY-MM-DD
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return fmt.Errorf("value '%s' is not a valid partial %s (YYYY-MM-DD): %w", s, typeName, err)
		}
	default:
		return fmt.Errorf("value '%s' has invalid length for a partial %s", s, typeName)
	}
	return nil
}

// ValidateFhirDate checks a string against FHIR date rules (YYYY[-MM[-DD]]).
// Eliminates the regex completely.
func ValidateFhirDate(v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("value is not a string type")
	}
	// Delegate to the shared partial date logic.
	return validatePartialFhirDate(s, "FHIR date")
}

func ValidateFhirTime(v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("value is not a string type")
	}

	// Fast check for minimum length (HH:MM:SS is 8 chars)
	if len(s) < 8 {
		return fmt.Errorf("value '%s' is too short for a FHIR time (minimum 8 characters)", s)
	}

	timePart := s
	if s[8] == '.' {
		// Time includes fractional seconds, separate HH:MM:SS part
		timePart = s[:8]

		// Ensure there's at least one digit after the decimal point
		if len(s) == 9 {
			return fmt.Errorf("value '%s' is not a valid FHIR time: missing fractional second digits after '.'", s)
		}

		// Ensure all characters after the dot are digits
		for i := 9; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				return fmt.Errorf("value '%s' is not a valid FHIR time: non-digit characters in fractional seconds", s)
			}
		}
	} else if len(s) != 8 {
		// If no '.', the length must be exactly 8 (HH:MM:SS)
		return fmt.Errorf("value '%s' has invalid length for FHIR time (expected 8 or 9+ with fractional seconds)", s)
	}

	// Fast structural check for colons
	if timePart[2] != ':' || timePart[5] != ':' {
		return fmt.Errorf("value '%s' is not a valid FHIR time: missing or misplaced colons", s)
	}

	// 2. Validate core HH:MM:SS components using highly performant time.Parse.
	// This single function ensures HH is < 24, MM < 60, SS < 60.
	if _, err := time.Parse("15:04:05", timePart); err != nil {
		return fmt.Errorf("value '%s' is not a valid FHIR time: invalid time components (%w)", s, err)
	}

	return nil
}

// ValidateFhirURI checks a string against FHIR URI rules (must be an absolute URI).
// Relies on a single, efficient standard library check.
func ValidateFhirURI(v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("value is not a string type")
	}

	// url.ParseRequestURI is efficient and checks for validity.
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return fmt.Errorf("value '%s' is not a valid URI: %w", s, err)
	}

	// IsAbs() is a quick check on the parsed structure.
	if !u.IsAbs() {
		return fmt.Errorf("value '%s' is not an absolute URI (must include a scheme like 'http://' or 'urn:')", s)
	}

	return nil
}

// ValidateFhirUUID checks a string against FHIR UUID rules (RFC 4122).
// Relies on a single, highly-optimized dedicated library check.
func ValidateFhirUUID(v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("value is not a string type")
	}

	// The uuid.Parse function is a single, highly performant check.
	if _, err := uuid.Parse(s); err != nil {
		return fmt.Errorf("value '%s' is not a valid UUID (RFC 4122): %w", s, err)
	}

	return nil
}

// ValidateFhirBinary checks that a string is a valid base64-encoded value.
// Relies on a single, efficient standard library check.
func ValidateFhirBinary(v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("value is not a string type")
	}

	// base64.StdEncoding.DecodeString is a single, robust, and performant check.
	if _, err := base64.StdEncoding.DecodeString(s); err != nil {
		return fmt.Errorf("value is not a valid Base64 encoded string: %w", err)
	}

	return nil
}
