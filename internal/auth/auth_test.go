package auth

import "testing"

func assertError(t testing.TB, got bool, want bool) {
	t.Helper()
	if got != want {
		t.Errorf("got error = %v, want error = %v", got, want)
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "mypassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "holyshitthisisasuperlongpasswordhopeitsgood1234567890",
			wantErr:  false,
		},
		{
			name:     "short password",
			password: "it",
			wantErr:  false,
		},
		{
			name:     "password with special characters",
			password: "cool!@@##$$%^&&",
			wantErr:  false,
		}, {
			name:     "password with spaces",
			password: "cool beans",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := HashPassword(tt.password)
			assertError(t, gotErr != nil, tt.wantErr)
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "correctpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to generate hash for test: %v", err)
	}
	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "matching password and hash",
			password: password,
			hash:     hash,
			wantErr:  false,
		},
		{
			name:     "wrong hash for password",
			password: password,
			hash:     "stuff",
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "wrong password",
			password: "wrongpassword",
			hash:     hash,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := CheckPasswordHash(tt.password, tt.hash)
			assertError(t, gotErr != nil, tt.wantErr)
		})
	}
}
