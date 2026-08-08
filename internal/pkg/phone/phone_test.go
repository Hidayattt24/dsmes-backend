package phone

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "Standard 08 format",
			input:   "081234567890",
			want:    "6281234567890",
			wantErr: false,
		},
		{
			name:    "Direct 8 format without zero (from +62 UI field)",
			input:   "82211562536",
			want:    "6282211562536",
			wantErr: false,
		},
		{
			name:    "With +62 format",
			input:   "+6281234567890",
			want:    "6281234567890",
			wantErr: false,
		},
		{
			name:    "With spaces and dashes",
			input:   "+62 812-3456-7890",
			want:    "6281234567890",
			wantErr: false,
		},
		{
			name:    "Direct 62 format",
			input:   "6281234567890",
			want:    "6281234567890",
			wantErr: false,
		},
		{
			name:    "Empty input",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Too short",
			input:   "0812",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Non Indonesian prefix",
			input:   "1555123456",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Normalize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Normalize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Normalize() = %v, want %v", got, tt.want)
			}
		})
	}
}
