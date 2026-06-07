// Example: phone number utilities and SMS segment calculation.
//
// Run: go run .
package main

import (
	"fmt"
	"strings"

	"github.com/KARTIKrocks/gosms"
)

func main() {
	// --- E.164 validation ---
	numbers := []string{"+15551234567", "5551234567", "+1", "+442071838750", "invalid"}
	for _, n := range numbers {
		fmt.Printf("ValidateE164(%q) = %t\n", n, gosms.ValidateE164(n))
	}

	fmt.Println()

	// --- Phone normalization ---
	normalized := gosms.NormalizePhone("555-123-4567", "+1")
	fmt.Printf("NormalizePhone(\"555-123-4567\", \"+1\") = %s\n", normalized)

	normalized = gosms.NormalizePhone("+44 20 7183 8750", "")
	fmt.Printf("NormalizePhone(\"+44 20 7183 8750\", \"\") = %s\n", normalized)

	fmt.Println()

	// --- GSM encoding detection ---
	messages := []string{
		"Hello, world!",
		"Hello 🌍",
		"Price: €10 {sale}",
	}
	for _, m := range messages {
		gsm := gosms.IsGSMEncoding(m)
		segments := gosms.CalculateSegments(m)
		encoding := "Unicode"
		if gsm {
			encoding = fmt.Sprintf("GSM (len=%d)", gosms.GSMLen(m))
		}
		fmt.Printf("%q → %s, %d segment(s)\n", m, encoding, segments)
	}

	fmt.Println()

	// --- Segment calculation for long messages ---
	short := "Hi there!"
	long := strings.Repeat("a", 200)
	longUnicode := strings.Repeat("🎉", 80)

	fmt.Printf("Short GSM (%d chars): %d segment(s)\n", len(short), gosms.CalculateSegments(short))
	fmt.Printf("Long GSM (%d chars): %d segment(s)\n", len(long), gosms.CalculateSegments(long))
	fmt.Printf("Long Unicode (80 emoji): %d segment(s)\n", gosms.CalculateSegments(longUnicode))
}
