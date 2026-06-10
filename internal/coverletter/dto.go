package coverletter

// Response is the structured output from the AI cover letter generation.
type Response struct {
	YourName         string `json:"your_name"`
	YourAddress      string `json:"your_address"`
	YourEmail        string `json:"your_email"`
	YourPhone        string `json:"your_phone"`
	RecipientName    string `json:"recipient_name"`
	RecipientTitle   string `json:"recipient_title"`
	CompanyName      string `json:"company_name"`
	CompanyAddress   string `json:"company_address"`
	Subject          string `json:"subject"`
	Opening          string `json:"opening"`
	BodyParagraphs   []string `json:"body_paragraphs"`
	Closing          string `json:"closing"`
}
