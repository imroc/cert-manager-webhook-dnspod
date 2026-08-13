package dnspod

import "testing"

func TestNormalizeDomainName(t *testing.T) {
	s := NewSolver()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "纯 ASCII 域名原样返回",
			input:    "_acme-challenge.example.com",
			expected: "_acme-challenge.example.com",
		},
		{
			name:     "punycode 转为 unicode",
			input:    "_acme-challenge.xn--fiqs8s.example.com",
			expected: "_acme-challenge.中国.example.com",
		},
		{
			name:     "已是 unicode 的域名保持不变",
			input:    "_acme-challenge.中国.example.com",
			expected: "_acme-challenge.中国.example.com",
		},
		{
			name:     "大小写归一化为小写",
			input:    "_acme-challenge.Example.COM",
			expected: "_acme-challenge.example.com",
		},
		{
			name:     "多标签 punycode 转 unicode",
			input:    "_acme-challenge.xn--fiqs8s.xn--fiqs8s",
			expected: "_acme-challenge.中国.中国",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.normalizeDomainName(tt.input); got != tt.expected {
				t.Errorf("normalizeDomainName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
