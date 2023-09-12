package model

type Target struct {
	Id      int64
	Name    string
	Ip      string
	Tag     string
	Enabled bool
	Flags   int64
}
