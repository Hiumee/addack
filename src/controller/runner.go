package controller

import "addack/src/model"

type ExploitRunner struct {
	exploits       map[string]*model.Exploit
	ExploitAdder   chan *model.Exploit
	ExploitRemover chan *model.Exploit
	targets        map[string]*model.Target
	TargetAdder    chan *model.Target
	TargetRemover  chan *model.Target
}

func NewExploitRunner() *ExploitRunner {
	return &ExploitRunner{
		exploits:       make(map[string]*model.Exploit),
		ExploitAdder:   make(chan *model.Exploit),
		ExploitRemover: make(chan *model.Exploit),
		targets:        make(map[string]*model.Target),
		TargetAdder:    make(chan *model.Target),
		TargetRemover:  make(chan *model.Target),
	}
}

func (er *ExploitRunner) Run() {
	for {
		select {
		case exploit := <-er.ExploitAdder:
			er.exploits[exploit.Path] = exploit
		case exploit := <-er.ExploitRemover:
			delete(er.exploits, exploit.Path)
		case target := <-er.TargetAdder:
			er.targets[target.Ip] = target
		case target := <-er.TargetRemover:
			delete(er.targets, target.Ip)
		}
	}
}
