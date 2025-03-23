package security

// SecurityTestHelper provides test helper methods for security functions
type SecurityTestHelper struct{}

// HashUsername returns a deterministic hash of a username for testing
func (h *SecurityTestHelper) HashUsername(username string) string {
	return HashUsername(username)
}
