package auth

import (
	"reflect"
	"testing"
)

func TestAuthService_ValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "StrongPass1!",
			wantErr:  false,
		},
		{
			name:     "Too short password",
			password: "short",
			wantErr:  true,
		},
		{
			name:     "Missing uppercase letter",
			password: "nouppercase1!",
			wantErr:  true,
		},
		{
			name:     "Missing lowercase letter",
			password: "NOLOWERCASE1!",
			wantErr:  true,
		},
		{
			name:     "Missing number",
			password: "NoNumber!",
			wantErr:  true,
		},
		{
			name:     "Missing special character",
			password: "NoSpecial1",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthService{}
			if err := a.ValidatePassword(tt.password); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewAuthService(t *testing.T) {
	tests := []struct {
		name          string
		encryptionKey string
		want          *AuthService
	}{
		{
			name:          "Valid encryption key",
			encryptionKey: "test-key",
			want:          &AuthService{encryptionKey: "test-key"},
		},
		{
			name:          "Empty encryption key",
			encryptionKey: "",
			want:          &AuthService{encryptionKey: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAuthService(tt.encryptionKey); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAuthService() = %v, want %v", got, tt.want)
			}
		})
	}
}
