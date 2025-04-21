package storage

type AuthStorage interface {
	SaveUser()
	GetUser()
	GetApp()
	IsAdmin()
}
