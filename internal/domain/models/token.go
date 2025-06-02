package models

type Tokens struct {
	AccessToken  Token
	RefreshToken Token
}

type Token struct {
	Token     string
	Expire_at int64
}
