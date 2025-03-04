package local

import (
	"encoding/json"
	"github.com/FollowLille/goph-vault/internal/models"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadSecrets(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(tempDir string)
		want    map[string]models.Secret
		wantErr bool
	}{
		{
			name: "Existing file is loaded successfully",
			setup: func(tempDir string) {
				secretsFilePath = filepath.Join(tempDir, "secrets.json")
				data := `{
                    "secret1": {
                        "Name": "secret1",
                        "Data": "ZGF0YTE=",
                        "Type": "password"
                    },
                    "secret2": {
                        "Name": "secret2",
                        "Data": "ZGF0YTI=", 
                        "Type": "text"
                    }
                }`
				os.WriteFile(secretsFilePath, []byte(data), 0600)
			},
			want: map[string]models.Secret{
				"secret1": {
					Name: "secret1",
					Data: []byte("data1"),
					Type: "password",
				},
				"secret2": {
					Name: "secret2",
					Data: []byte("data2"),
					Type: "text",
				},
			},
			wantErr: false,
		},
		{
			name: "Default empty map is returned when file does not exist",
			setup: func(tempDir string) {
				secretsFilePath = filepath.Join(tempDir, "nonexistent.json")
			},
			want:    map[string]models.Secret{},
			wantErr: false,
		},
		{
			name: "Error when file contains invalid JSON",
			setup: func(tempDir string) {
				secretsFilePath = filepath.Join(tempDir, "secrets.json")
				os.WriteFile(secretsFilePath, []byte("invalid-json"), 0600)
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			if tt.setup != nil {
				tt.setup(tempDir)
			}

			got, err := LoadSecrets()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadSecrets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadSecrets() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveSecrets(t *testing.T) {
	type args struct {
		secrets map[string]models.Secret
	}
	tests := []struct {
		name    string
		args    args
		setup   func(tempDir string)
		wantErr bool
	}{
		{
			name: "Secrets are saved successfully",
			args: args{
				secrets: map[string]models.Secret{
					"secret1": {
						Name: "secret1",
						Data: []byte("data1"),
						Type: "password",
					},
					"secret2": {
						Name: "secret2",
						Data: []byte("data2"),
						Type: "text",
					},
				},
			},
			setup: func(tempDir string) {
				secretsFilePath = filepath.Join(tempDir, "secrets.json")
			},
			wantErr: false,
		},
		{
			name: "Error when unable to create directory",
			args: args{
				secrets: map[string]models.Secret{},
			},
			setup: func(tempDir string) {
				secretsFilePath = filepath.Join(tempDir, "unwritable/secrets.json")
				// Создаем файл с тем же именем, что и директория
				os.WriteFile(filepath.Dir(secretsFilePath), []byte{}, 0600)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			if tt.setup != nil {
				tt.setup(tempDir)
			}

			err := SaveSecrets(tt.args.secrets)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveSecrets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Проверяем содержимое файла при успешном сохранении
			if !tt.wantErr {
				fileData, err := os.ReadFile(secretsFilePath)
				if err != nil {
					t.Errorf("Failed to read saved secrets file: %v", err)
					return
				}

				var savedSecrets map[string]models.Secret
				if err := json.Unmarshal(fileData, &savedSecrets); err != nil {
					t.Errorf("Failed to unmarshal saved secrets: %v", err)
					return
				}

				if !reflect.DeepEqual(savedSecrets, tt.args.secrets) {
					t.Errorf("Saved secrets do not match input. Got = %v, Want = %v", savedSecrets, tt.args.secrets)
				}
			}
		})
	}
}
