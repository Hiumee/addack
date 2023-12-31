package model

type Flag struct {
	Id        int64
	Flag      string
	ExploitId int64
	TargetId  int64
	Result    string
	Valid     string
}

type FlagDTO struct {
	Id          int64
	Flag        string
	ExploitName string
	TargetName  string
	Result      string
	Valid       string
	Timestamp   string
}
