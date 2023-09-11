package controller

import "addack/src/model"

type ExploitRunner struct {
	challenges       map[string]*model.Challenge
	ChallengeAdder   chan *model.Challenge
	ChallengeRemover chan *model.Challenge
	targets          map[string]*model.Target
	TargetAdder      chan *model.Target
	TargetRemover    chan *model.Target
}

func NewExploitRunner() *ExploitRunner {
	return &ExploitRunner{
		challenges:       make(map[string]*model.Challenge),
		ChallengeAdder:   make(chan *model.Challenge),
		ChallengeRemover: make(chan *model.Challenge),
		targets:          make(map[string]*model.Target),
		TargetAdder:      make(chan *model.Target),
		TargetRemover:    make(chan *model.Target),
	}
}

func (er *ExploitRunner) Run() {
	for {
		select {
		case challenge := <-er.ChallengeAdder:
			er.challenges[challenge.Path] = challenge
		case challenge := <-er.ChallengeRemover:
			delete(er.challenges, challenge.Path)
		case target := <-er.TargetAdder:
			er.targets[target.Ip] = target
		case target := <-er.TargetRemover:
			delete(er.targets, target.Ip)
		}
	}
}
