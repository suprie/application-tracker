package jdextractor

import "fmt"

func BuildJDParserPrompt(jdText string) string {
	return fmt.Sprintf(`
You are an expert job description parsing engine.

Your task is to convert a raw Job Description into a structured JSON document.

The JSON will be used by an AI-powered Application Tracking System to compare the job requirements against a candidate's master_profile.yaml using semantic matching.

Your goal is accuracy, not creativity.

# CRITICAL RULES

1. Extract facts only.
2. Never invent information.
3. Never infer requirements that are not present.
4. Use null when information is missing.
5. Preserve important wording from the job description.
6. Normalize skills into the provided taxonomy.
7. Deduplicate repeated skills.
8. Output valid JSON only.
9. Do not use markdown.
10. Do not use triple backticks.
11. Do not add explanations.
12. If uncertain, add an item to parsing_warnings.
13. Prefer missing data over incorrect data.

## Must Have vs Nice To Have

must_have:
Use only skills explicitly described as required, essential, must-have, minimum qualification, or core responsibility.

nice_to_have:
Use skills described as preferred, bonus, plus, desirable, advantage, or nice-to-have.

If the JD does not clearly separate required and preferred skills, place core repeated requirements under must_have and add a parsing warning.

# INPUT JOB DESCRIPTION

%s

# OUTPUT

Return raw JSON only, no triple tick for markdown format.
	`, jdText)
}
