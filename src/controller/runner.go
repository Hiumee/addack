package controller

import (
	"addack/src/model"
	"bytes"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"
)

type Runner struct {
	Exploit    *model.Exploit
	Target     *model.Target
	Notify     chan string
	Flagger    chan *model.Flag
	Controller *Controller
}

type ExploitRunner struct {
	exploits       map[int64]*model.Exploit
	exploitsLock   sync.RWMutex
	ExploitAdder   chan *model.Exploit
	ExploitRemover chan *model.Exploit
	targets        map[int64]*model.Target
	targetsLock    sync.RWMutex
	TargetAdder    chan *model.Target
	TargetRemover  chan *model.Target
	runner         map[int64]map[int64]*Runner
	runnerLock     sync.RWMutex
	Flagger        chan *model.Flag
	controller     *Controller
}

func (er *ExploitRunner) NewRunner(exploit *model.Exploit, target *model.Target) *Runner {
	return &Runner{
		Exploit:    exploit,
		Target:     target,
		Notify:     make(chan string, 1),
		Flagger:    er.Flagger,
		Controller: er.controller,
	}
}

func (r *Runner) Run() {
	log.Default().Println("Runner started", r.Exploit.Name, r.Target.Name)
	for {
		var output bytes.Buffer

		cmd := exec.Command("bash", "-c", r.Exploit.Command)
		cmd.Dir = r.Controller.Config.ExploitsPath + r.Exploit.Path
		cmd.Env = append(cmd.Env, "TARGET="+r.Target.Ip)
		writer := io.Writer(&output)
		cmd.Stdout = writer

		if err := cmd.Start(); err != nil {
			log.Default().Println("Runner error", r.Exploit.Name, r.Target.Name, err)
			time.Sleep(time.Duration(r.Controller.Config.TickTime) * time.Millisecond)
			continue
		}
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		timer := time.NewTimer(time.Duration(r.Exploit.Timeout) * time.Millisecond)
		select {
		case <-r.Notify:
			log.Default().Println("Runner stopped", r.Exploit.Name, r.Target.Name)
			cmd.Process.Kill()
			close(r.Notify)
			return
		case <-done:
			flag := &model.Flag{
				ExploitId: r.Exploit.Id,
				TargetId:  r.Target.Id,
				Result:    string(output.Bytes()),
				Valid:     true,
			}
			r.Flagger <- flag
		case <-timer.C:
			log.Default().Println("Runner timeout", r.Exploit.Name, r.Target.Name)
		}

		timer.Stop()

		time.Sleep(time.Duration(r.Controller.Config.TickTime) * time.Millisecond)
	}
}

func NewExploitRunner(controller *Controller) *ExploitRunner {
	return &ExploitRunner{
		exploits:       make(map[int64]*model.Exploit),
		exploitsLock:   sync.RWMutex{},
		ExploitAdder:   make(chan *model.Exploit, 5),
		ExploitRemover: make(chan *model.Exploit, 5),
		targets:        make(map[int64]*model.Target),
		targetsLock:    sync.RWMutex{},
		TargetAdder:    make(chan *model.Target, 5),
		TargetRemover:  make(chan *model.Target, 5),
		runner:         make(map[int64]map[int64]*Runner),
		runnerLock:     sync.RWMutex{},
		Flagger:        make(chan *model.Flag, 30),
		controller:     controller,
	}
}

func (er *ExploitRunner) addExploit(exploit *model.Exploit) {
	er.exploitsLock.Lock()
	er.exploits[exploit.Id] = exploit
	er.exploitsLock.Unlock()

	er.targetsLock.RLock()
	er.runnerLock.Lock()
	er.runner[exploit.Id] = make(map[int64]*Runner)

	for _, target := range er.targets {
		er.runner[exploit.Id][target.Id] = er.NewRunner(exploit, target)
		go er.runner[exploit.Id][target.Id].Run()
	}
	er.runnerLock.Unlock()
	er.targetsLock.RUnlock()

	log.Default().Println("ExploitRunner added exploit", exploit.Name)
}

func (er *ExploitRunner) removeExploit(exploit *model.Exploit) {
	er.exploitsLock.Lock()
	delete(er.exploits, exploit.Id)
	er.exploitsLock.Unlock()

	er.runnerLock.Lock()
	runners := []*Runner{}

	for _, runner := range er.runner[exploit.Id] {
		runners = append(runners, runner)
	}

	for _, runner := range runners {
		runner.Notify <- "stop"
		delete(er.runner[exploit.Id], runner.Target.Id)
	}
	er.runnerLock.Unlock()

	log.Default().Println("ExploitRunner removed exploit", exploit.Id)
	log.Default().Println(er.runner)
}

func (er *ExploitRunner) addTarget(target *model.Target) {
	er.targetsLock.Lock()
	er.targets[target.Id] = target
	er.targetsLock.Unlock()
	er.exploitsLock.RLock()
	er.runnerLock.Lock()
	for _, exploit := range er.exploits {
		er.runner[exploit.Id][target.Id] = er.NewRunner(exploit, target)
		go er.runner[exploit.Id][target.Id].Run()
	}
	er.runnerLock.Unlock()
	er.exploitsLock.RUnlock()

	log.Default().Println("ExploitRunner added target", target.Name)
}

func (er *ExploitRunner) removeTarget(target *model.Target) {
	er.targetsLock.Lock()
	delete(er.targets, target.Id)
	er.targetsLock.Unlock()

	er.runnerLock.Lock()

	for id, exploit := range er.runner {
		runner := exploit[target.Id]
		runner.Notify <- "stop"
		delete(er.runner[id], runner.Target.Id)
	}
	er.runnerLock.Unlock()

	log.Default().Println("ExploitRunner removed target", target.Id)
}

func (er *ExploitRunner) Run() {
	log.Default().Println("ExploitRunner started")
	for {
		select {
		case exploit := <-er.ExploitAdder:
			er.addExploit(exploit)
		case exploit := <-er.ExploitRemover:
			er.removeExploit(exploit)
		case target := <-er.TargetAdder:
			er.addTarget(target)
		case target := <-er.TargetRemover:
			er.removeTarget(target)
		case flag := <-er.Flagger:
			er.controller.DB.CreateFlag(*flag)
		}
	}
}
