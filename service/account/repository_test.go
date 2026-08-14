package account

import "testing"

func TestDuplicateAccountID(t *testing.T) {
	tests := []struct {
		message string
		want    bool
	}{
		{"Duplicate entry 'id' for key 'acc'", true},
		{"Duplicate entry 'id' for key 'account.acc'", true},
		{"Duplicate entry 'id' for key `nova.account.acc`", true},
		{"Duplicate entry 'name' for key 'PRIMARY'", false},
	}
	for _, test := range tests {
		if got := duplicateAccountID(test.message); got != test.want {
			t.Errorf("duplicateAccountID(%q) = %v, want %v", test.message, got, test.want)
		}
	}
}
