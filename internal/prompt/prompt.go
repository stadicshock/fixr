package prompt

import "fmt"

// BuildGrammarPrompt creates the system prompt for grammar correction.
func BuildGrammarPrompt(tone, language string) string {
	if tone == "" {
		tone = "professional"
	}
	if language == "" {
		language = "en-US"
	}

	return fmt.Sprintf(`You are a grammar and writing assistant. Your job is to fix grammar, spelling, and punctuation errors in the user's text.

Rules:
1. Fix all grammar, spelling, and punctuation errors.
2. Maintain the original meaning and intent.
3. Use a %s tone.
4. Use %s language conventions.
5. Preserve the original formatting (line breaks, bullet points, etc.).
6. Do NOT add explanations or commentary — return ONLY the corrected text.
7. If the text is already correct, return it unchanged.
8. If a screenshot is provided, use it as context to understand the conversation tone and setting, but only fix the provided text.
9. Keep the corrections minimal — don't rewrite sentences unless grammatically necessary.`, tone, language)
}

// BuildUserMessage creates the user message with the text to fix.
func BuildUserMessage(text string, hasScreenshot bool) string {
	if hasScreenshot {
		return fmt.Sprintf("Fix the grammar in the following text. A screenshot of the active window is attached for context about the conversation setting and tone.\n\nText to fix:\n%s", text)
	}
	return fmt.Sprintf("Fix the grammar in the following text:\n\n%s", text)
}
