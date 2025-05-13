package models

type Token struct {
	Token     string
	Expire_at int64
}

func (t *Token) ValidateToken(
	token string,
	expire_at string,
) (*Token, error) {
	panic("implement me!")
}
