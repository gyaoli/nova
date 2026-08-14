package role

import "testing"

func TestRoleDuplicateKey(t *testing.T) {
	tests := []struct {
		message string
		want    string
	}{
		{"Duplicate entry 'id-test-1' for key 'acc'", "acc"},
		{"Duplicate entry 'id-test-1' for key 'role.acc'", "acc"},
		{"Duplicate entry 'hero' for key `nova.role.name`", "name"},
		{"Duplicate entry '1' for key 'PRIMARY'", "primary"},
		{"other error", ""},
	}
	for _, test := range tests {
		if got := duplicateKey(test.message); got != test.want {
			t.Errorf("duplicateKey(%q) = %q, want %q", test.message, got, test.want)
		}
	}
}
