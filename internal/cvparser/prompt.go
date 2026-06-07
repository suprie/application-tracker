package cvparser

import (
	"fmt"
)

func BuildCVParserPrompt(cvText string) string {
	return fmt.Sprintf(`
You are an expert career profile extraction engine.

Convert this raw CV text into valid YAML.

Rules:
- Extract facts only.
- Never invent information.
- Use null when missing.
- Output YAML. Markdown fences are unacceptable

Schema:
name:
headline:
location:
email:
phone:
linkedin:
summary:
total_years_experience:
domains:
  -
skills:
  languages:
    -
  mobile:
    -
  architecture:
    -
  testing:
    -
  ci_cd:
    -
  tools:
    -
  leadership:
    -
experience:
  -
    title:
    company:
    start_date:
    end_date:
    team_size:
    highlights:
      -
education:
  -
    school:
    degree:
    start_date:
    end_date:
parsing_warnings:
  -

CV TEXT:
%s
`, cvText)
}
