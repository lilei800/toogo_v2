// Package encrypt
// @Description AES加密测试
package encrypt

import (
	"testing"
)

func TestAESEncryptor(t *testing.T) {
	// 测试加密器创建
	key := "toogo_test_key_32_bytes_length!!"
	enc, err := NewAESEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	// 测试加密解密
	testCases := []struct {
		name      string
		plaintext string
	}{
		{"Empty string", ""},
		{"Simple string", "hello world"},
		{"API Key", "vmPUZE6mv9SD5VNHk4HlWFsOr6aKE2zvsw0MuIgwCIPy6utIco14y7Ju91duEh8A"},
		{"Secret Key", "NhqPtmdSJYdKjVHjA7PZj4Mge3R5YNiP1e3UZjInClVN65XAbvqqM6A7H5fATj0j"},
		{"Chinese text", "这是一个测试字符串"},
		{"Special chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"Long string", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 加密
			encrypted, err := enc.Encrypt(tc.plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// 空字符串不需要前缀
			if tc.plaintext != "" && !IsEncrypted(encrypted) {
				t.Errorf("Encrypted string should have prefix")
			}

			// 解密
			decrypted, err := enc.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			// 验证
			if decrypted != tc.plaintext {
				t.Errorf("Decrypted text mismatch: got %q, want %q", decrypted, tc.plaintext)
			}
		})
	}
}

func TestEncryptApiKey(t *testing.T) {
	// 测试全局加密函数
	apiKey := "test_api_key_12345"

	encrypted, err := EncryptApiKey(apiKey)
	if err != nil {
		t.Fatalf("EncryptApiKey failed: %v", err)
	}

	if !IsEncrypted(encrypted) {
		t.Error("Encrypted string should have prefix")
	}

	decrypted, err := DecryptApiKey(encrypted)
	if err != nil {
		t.Fatalf("DecryptApiKey failed: %v", err)
	}

	if decrypted != apiKey {
		t.Errorf("Decrypted key mismatch: got %q, want %q", decrypted, apiKey)
	}
}

func TestDecryptUnencrypted(t *testing.T) {
	// 测试解密未加密的字符串
	plaintext := "not_encrypted_string"

	decrypted, err := DecryptApiKey(plaintext)
	if err != nil {
		t.Fatalf("Decrypt unencrypted should not fail: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Should return original string: got %q, want %q", decrypted, plaintext)
	}
}

func TestMaskApiKey(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"abcd", "****"},
		{"abcdefgh", "****"},
		{"abcdefghi", "abcd****fghi"},
		{"vmPUZE6mv9SD5VNHk4HlWFsOr6aKE2zvsw0MuIgwCIPy6utIco14y7Ju91duEh8A", "vmPU****Eh8A"},
	}

	for _, tc := range testCases {
		result := MaskApiKey(tc.input)
		if result != tc.expected {
			t.Errorf("MaskApiKey(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestMaskSecretKey(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"abcd", "********"},
		{"abcdefgh", "********"},
		{"abcdefghij", "abcd********ghij"},
	}

	for _, tc := range testCases {
		result := MaskSecretKey(tc.input)
		if result != tc.expected {
			t.Errorf("MaskSecretKey(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func BenchmarkEncrypt(b *testing.B) {
	enc, _ := NewAESEncryptor("toogo_test_key_32_bytes_length!!")
	plaintext := "vmPUZE6mv9SD5VNHk4HlWFsOr6aKE2zvsw0MuIgwCIPy6utIco14y7Ju91duEh8A"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Encrypt(plaintext)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	enc, _ := NewAESEncryptor("toogo_test_key_32_bytes_length!!")
	plaintext := "vmPUZE6mv9SD5VNHk4HlWFsOr6aKE2zvsw0MuIgwCIPy6utIco14y7Ju91duEh8A"
	encrypted, _ := enc.Encrypt(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Decrypt(encrypted)
	}
}

