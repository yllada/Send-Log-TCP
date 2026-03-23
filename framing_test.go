package main

import (
	"bytes"
	"strings"
	"testing"
)

// =============================================================================
// FRAMING TESTS - RFC 6587 Compliance
// Tests for TCP framing methods: Octet Counting and Non-Transparent Framing
// =============================================================================

func TestDefaultFramingConfig(t *testing.T) {
	config := DefaultFramingConfig()

	if config.Method != OctetCounting {
		t.Errorf("expected Method = OctetCounting, got %v", config.Method)
	}
	if !config.ValidateUTF8 {
		t.Error("expected ValidateUTF8 = true")
	}
	if config.MaxMessageLength != 0 {
		t.Errorf("expected MaxMessageLength = 0, got %d", config.MaxMessageLength)
	}
}

func TestNewFramer(t *testing.T) {
	config := FramingConfig{
		Method:           NonTransparent,
		ValidateUTF8:     false,
		MaxMessageLength: 1024,
	}

	framer := NewFramer(config)

	if framer == nil {
		t.Fatal("NewFramer returned nil")
	}
	if framer.config.Method != NonTransparent {
		t.Errorf("expected Method = NonTransparent, got %v", framer.config.Method)
	}
}

// =============================================================================
// OCTET COUNTING TESTS (RFC 6587 Section 3.4.1)
// =============================================================================

func TestFrameOctetCounting(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "simple message",
			message:  "hello",
			expected: "5 hello",
		},
		{
			name:     "syslog format message",
			message:  "<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su",
			expected: "55 <34>1 2003-10-11T22:14:15.003Z mymachine.example.com su",
		},
		{
			name:     "message with newline (allowed in octet-counting)",
			message:  "line1\nline2",
			expected: "11 line1\nline2",
		},
		{
			name:     "unicode message",
			message:  "日本語メッセージ",
			expected: "24 日本語メッセージ", // 24 bytes in UTF-8 (8 chars × 3 bytes each)
		},
		{
			name:     "single character",
			message:  "x",
			expected: "1 x",
		},
		{
			name:     "message with special chars",
			message:  "alert: CPU > 90%",
			expected: "16 alert: CPU > 90%",
		},
	}

	framer := NewFramer(FramingConfig{
		Method:       OctetCounting,
		ValidateUTF8: true,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := framer.Frame(tt.message)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(result) != tt.expected {
				t.Errorf("got %q, want %q", string(result), tt.expected)
			}
		})
	}
}

// =============================================================================
// NON-TRANSPARENT FRAMING TESTS (RFC 6587 Section 3.4.2)
// =============================================================================

func TestFrameNonTransparent(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "simple message",
			message:  "hello",
			expected: "hello\n",
		},
		{
			name:     "syslog format message",
			message:  "<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su",
			expected: "<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su\n",
		},
		{
			name:     "unicode message",
			message:  "日本語メッセージ",
			expected: "日本語メッセージ\n",
		},
	}

	framer := NewFramer(FramingConfig{
		Method:       NonTransparent,
		ValidateUTF8: true,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := framer.Frame(tt.message)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(result) != tt.expected {
				t.Errorf("got %q, want %q", string(result), tt.expected)
			}
		})
	}
}

func TestFrameNonTransparentRejectsNewline(t *testing.T) {
	framer := NewFramer(FramingConfig{
		Method:       NonTransparent,
		ValidateUTF8: true,
	})

	_, err := framer.Frame("line1\nline2")
	if err == nil {
		t.Error("expected error for message containing newline in non-transparent framing")
	}

	if !strings.Contains(err.Error(), "line feed") {
		t.Errorf("error should mention 'line feed', got: %v", err)
	}
}

// =============================================================================
// VALIDATION TESTS
// =============================================================================

func TestFrameValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      FramingConfig
		message     string
		expectError bool
		errorMsg    string
	}{
		{
			name: "empty message",
			config: FramingConfig{
				Method:       OctetCounting,
				ValidateUTF8: true,
			},
			message:     "",
			expectError: true,
			errorMsg:    "empty",
		},
		{
			name: "message exceeds max length",
			config: FramingConfig{
				Method:           OctetCounting,
				ValidateUTF8:     true,
				MaxMessageLength: 10,
			},
			message:     "this message is way too long",
			expectError: true,
			errorMsg:    "exceeds maximum",
		},
		{
			name: "message at exact max length",
			config: FramingConfig{
				Method:           OctetCounting,
				ValidateUTF8:     true,
				MaxMessageLength: 10,
			},
			message:     "1234567890", // exactly 10 bytes
			expectError: false,
		},
		{
			name: "invalid UTF-8 with validation enabled",
			config: FramingConfig{
				Method:       OctetCounting,
				ValidateUTF8: true,
			},
			message:     string([]byte{0xff, 0xfe, 0x00}), // invalid UTF-8
			expectError: true,
			errorMsg:    "invalid UTF-8",
		},
		{
			name: "invalid UTF-8 with validation disabled",
			config: FramingConfig{
				Method:       OctetCounting,
				ValidateUTF8: false,
			},
			message:     string([]byte{0xff, 0xfe, 0x00}), // invalid UTF-8 but validation off
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			framer := NewFramer(tt.config)
			_, err := framer.Frame(tt.message)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("error should contain %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// =============================================================================
// BATCH FRAMING TESTS
// =============================================================================

func TestFrameBatch(t *testing.T) {
	tests := []struct {
		name     string
		method   FramingMethod
		messages []string
		expected string
	}{
		{
			name:     "multiple messages octet counting",
			method:   OctetCounting,
			messages: []string{"msg1", "msg2", "msg3"},
			expected: "4 msg14 msg24 msg3",
		},
		{
			name:     "multiple messages non-transparent",
			method:   NonTransparent,
			messages: []string{"msg1", "msg2", "msg3"},
			expected: "msg1\nmsg2\nmsg3\n",
		},
		{
			name:     "single message batch",
			method:   OctetCounting,
			messages: []string{"single"},
			expected: "6 single",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			framer := NewFramer(FramingConfig{
				Method:       tt.method,
				ValidateUTF8: true,
			})

			result, err := framer.FrameBatch(tt.messages)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(result) != tt.expected {
				t.Errorf("got %q, want %q", string(result), tt.expected)
			}
		})
	}
}

func TestFrameBatchEmpty(t *testing.T) {
	framer := NewFramer(DefaultFramingConfig())

	_, err := framer.FrameBatch([]string{})
	if err == nil {
		t.Error("expected error for empty batch")
	}
}

func TestFrameBatchWithInvalidMessage(t *testing.T) {
	framer := NewFramer(FramingConfig{
		Method:       NonTransparent,
		ValidateUTF8: true,
	})

	// Second message contains newline, should fail
	_, err := framer.FrameBatch([]string{"valid", "invalid\nmessage", "valid2"})
	if err == nil {
		t.Error("expected error for batch containing invalid message")
	}

	if !strings.Contains(err.Error(), "message 1") {
		t.Errorf("error should indicate which message failed, got: %v", err)
	}
}

// =============================================================================
// UTILITY FUNCTION TESTS
// =============================================================================

func TestIsValidFramingMethod(t *testing.T) {
	tests := []struct {
		method FramingMethod
		valid  bool
	}{
		{OctetCounting, true},
		{NonTransparent, true},
		{FramingMethod("invalid"), false},
		{FramingMethod(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			if got := IsValidFramingMethod(tt.method); got != tt.valid {
				t.Errorf("IsValidFramingMethod(%q) = %v, want %v", tt.method, got, tt.valid)
			}
		})
	}
}

func TestRecommendedFramingMethod(t *testing.T) {
	method := RecommendedFramingMethod()
	if method != OctetCounting {
		t.Errorf("expected OctetCounting, got %v", method)
	}
}

// =============================================================================
// EDGE CASES AND RFC COMPLIANCE
// =============================================================================

func TestOctetCountingLargeMessage(t *testing.T) {
	// Test with a message that has multi-digit length prefix
	framer := NewFramer(FramingConfig{
		Method:       OctetCounting,
		ValidateUTF8: false, // disable for raw bytes
	})

	// Create a 1000 byte message
	largeMsg := strings.Repeat("x", 1000)
	result, err := framer.Frame(largeMsg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be "1000 " + 1000 x's = 1005 bytes total
	if len(result) != 1005 {
		t.Errorf("expected 1005 bytes, got %d", len(result))
	}

	// Verify prefix
	if !bytes.HasPrefix(result, []byte("1000 ")) {
		t.Errorf("expected prefix '1000 ', got %q", result[:10])
	}
}

func TestFrameUnknownMethodFallsBackToOctetCounting(t *testing.T) {
	framer := NewFramer(FramingConfig{
		Method:       FramingMethod("unknown"),
		ValidateUTF8: true,
	})

	result, err := framer.Frame("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use octet-counting as fallback
	expected := "4 test"
	if string(result) != expected {
		t.Errorf("got %q, want %q (octet-counting fallback)", string(result), expected)
	}
}

// =============================================================================
// BENCHMARKS
// =============================================================================

func BenchmarkFrameOctetCounting(b *testing.B) {
	framer := NewFramer(FramingConfig{
		Method:       OctetCounting,
		ValidateUTF8: true,
	})
	msg := "<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - Test message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = framer.Frame(msg)
	}
}

func BenchmarkFrameNonTransparent(b *testing.B) {
	framer := NewFramer(FramingConfig{
		Method:       NonTransparent,
		ValidateUTF8: true,
	})
	msg := "<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - Test message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = framer.Frame(msg)
	}
}

func BenchmarkFrameBatch100Messages(b *testing.B) {
	framer := NewFramer(FramingConfig{
		Method:       OctetCounting,
		ValidateUTF8: true,
	})

	messages := make([]string, 100)
	for i := range messages {
		messages[i] = "<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - Test message"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = framer.FrameBatch(messages)
	}
}
