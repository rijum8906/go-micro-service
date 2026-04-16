package dto

type CompanyInfo struct {
	Name        string       `validate:"omitempty"`
	Emails      []string     `validate:"omitempty,dive,email"`
	Addresses   []string     `validate:"omitempty,dive,required"`
	WebsiteURL  string       `validate:"omitempty,url"`
	SocialLinks []SocialLink `validate:"omitempty,dive"`
}

type SocialLink struct {
	Label string `validate:"required"`
	URL   string `validate:"required,url"`
}
