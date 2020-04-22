package lib

type apiLoginResponse struct {
	Status           string `json:"status"`
	DeveloperMessage string `json:"_developerMessage"`
	Created          string `json:"_created"`
	URI              string `json:"_uri"`
	RequestID        string `json:"_requestId"`
	User             struct {
		LastName  string `json:"lastName"`
		UserEmail string `json:"userEmail"`
		UserID    string `json:"userId"`
		FirstName string `json:"firstName"`
	} `json:"user"`
	Token      string `json:"token"`
	IdentityID string `json:"identityId"`
}
