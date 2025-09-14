package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		secret      string
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "successful hash",
			password:    "testpassword",
			secret:      "secretkey",
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "empty password",
			password:    "",
			secret:      "secretkey",
			wantErr:     true,
			expectedErr: ErrEmptyPassword,
		},
		{
			name:        "empty secret",
			password:    "testpassword",
			secret:      "",
			wantErr:     true,
			expectedErr: ErrEmptySecret,
		},
		{
			name:        "special characters password",
			password:    "p@ssw0rd!123",
			secret:      "secretkey",
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "long password",
			password:    "verylongpasswordthatexceedsnormalpasswordlengthlimits",
			secret:      "secretkey",
			wantErr:     false,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := &PasswordManager{}
			hashed, err := pm.HashPassword(tt.password, tt.secret)

			if tt.wantErr {
				if err == nil {
					t.Errorf("HashPassword() expected error, got nil")
					return
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("HashPassword() error = %v, expectedErr %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("HashPassword() unexpected error = %v", err)
				return
			}

			if hashed == "" {
				t.Errorf("HashPassword() returned empty string")
			}

			err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte("dummy"))
			if err != nil && err != bcrypt.ErrMismatchedHashAndPassword {
				t.Errorf("HashPassword() returned invalid bcrypt hash: %v", err)
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	secret := "testsecret"
	password := "correctpassword"
	wrongPassword := "wrongpassword"
	pm := &PasswordManager{}

	hashedPassword, err := pm.HashPassword(password, secret)
	if err != nil {
		t.Fatalf("Failed to hash password for test setup: %v", err)
	}

	tests := []struct {
		name        string
		hashed      string
		input       string
		secret      string
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "correct password",
			hashed:      hashedPassword,
			input:       password,
			secret:      secret,
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "wrong password",
			hashed:      hashedPassword,
			input:       wrongPassword,
			secret:      secret,
			wantErr:     true,
			expectedErr: bcrypt.ErrMismatchedHashAndPassword,
		},
		{
			name:        "wrong secret",
			hashed:      hashedPassword,
			input:       password,
			secret:      "wrongsecret",
			wantErr:     true,
			expectedErr: bcrypt.ErrMismatchedHashAndPassword,
		},
		{
			name:        "empty input password",
			hashed:      hashedPassword,
			input:       "",
			secret:      secret,
			wantErr:     true,
			expectedErr: bcrypt.ErrMismatchedHashAndPassword,
		},
		{
			name:        "empty secret",
			hashed:      hashedPassword,
			input:       password,
			secret:      "",
			wantErr:     true,
			expectedErr: bcrypt.ErrMismatchedHashAndPassword,
		},
		{
			name:        "malformed hash",
			hashed:      "invalidhash",
			input:       password,
			secret:      secret,
			wantErr:     true,
			expectedErr: bcrypt.ErrHashTooShort,
		},
		{
			name:        "empty hash",
			hashed:      "",
			input:       password,
			secret:      secret,
			wantErr:     true,
			expectedErr: bcrypt.ErrHashTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := &PasswordManager{}
			err := pm.CheckPassword(tt.hashed, tt.input, tt.secret)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CheckPassword() expected error, got nil")
					return
				}
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("CheckPassword() error = %v, expectedErr %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("CheckPassword() unexpected error = %v", err)
			}
		})
	}
}

func TestHashAndCheckIntegration(t *testing.T) {
	tests := []struct {
		name     string
		password string
		secret   string
	}{
		{"normal case", "password123", "secret123"},
		{"special chars", "p@ssw0rd!@#", "s3cr3t!@#"},
		{"long values", "verylongpassword", "verylongsecretkeythatisalsoverylong"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := &PasswordManager{}
			hashed, err := pm.HashPassword(tt.password, tt.secret)
			if err != nil {
				t.Errorf("HashPassword() failed: %v", err)
				return
			}

			err = pm.CheckPassword(hashed, tt.password, tt.secret)
			if err != nil {
				t.Errorf("CheckPassword() with correct password failed: %v", err)
			}

			err = pm.CheckPassword(hashed, "wrongpassword", tt.secret)
			if err == nil {
				t.Errorf("CheckPassword() with wrong password should have failed")
			}

			err = pm.CheckPassword(hashed, tt.password, "wrongsecret")
			if err == nil {
				t.Errorf("CheckPassword() with wrong secret should have failed")
			}
		})
	}
}

func TestHashPassword_DeterministicPeppering(t *testing.T) {
	// Test that the same password and secret always produces the same pepper
	password := "testpassword"
	secret := "testsecret"

	pepper1 := generatePepper(password, secret)
	pepper2 := generatePepper(password, secret)

	if pepper1 != pepper2 {
		t.Errorf("Peppering is not deterministic: %s != %s", pepper1, pepper2)
	}

	differentSecret := "differentsecret"
	pepper3 := generatePepper(password, differentSecret)

	if pepper1 == pepper3 {
		t.Errorf("Different secrets should produce different peppers")
	}

	differentPassword := "differentpassword"
	pepper4 := generatePepper(differentPassword, secret)

	if pepper1 == pepper4 {
		t.Errorf("Different passwords should produce different peppers")
	}
}

// Helper function to generate pepper for testing
func generatePepper(password, secret string) string {
	pepperedPassword := hmac.New(sha256.New, []byte(secret))
	pepperedPassword.Write([]byte(password))
	return hex.EncodeToString(pepperedPassword.Sum(nil))
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkpassword"
	secret := "benchmarksecret"
	pm := &PasswordManager{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.HashPassword(password, secret)
		if err != nil {
			b.Fatalf("HashPassword failed: %v", err)
		}
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	password := "benchmarkpassword"
	secret := "benchmarksecret"
	pm := &PasswordManager{}

	hashed, err := pm.HashPassword(password, secret)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := pm.CheckPassword(hashed, password, secret)
		if err != nil {
			b.Fatalf("CheckPassword failed: %v", err)
		}
	}
}
