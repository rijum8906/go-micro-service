package utils

func ValidateOrgnaziationMembershipStatus(status string) bool {
	return status == "active" || status == "suspended" || status == "left"
}
