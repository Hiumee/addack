package controller

import (
	"addack/src/model"
	"bytes"
	"io"
	"log"
	"os/exec"
	"regexp"
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
	ExploitAdder   chan *model.Exploit
	ExploitRemover chan *model.Exploit
	targets        map[int64]*model.Target
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
			result := string(output.Bytes())
			flagRegex := r.Controller.Config.FlagRegex
			flagString := ""

			re, err := regexp.Compile(flagRegex)

			if err != nil {
				log.Default().Println("Runner error", "Can't compile regex", flagRegex)
			} else {
				flagString = re.FindString(result)
			}

			flag := &model.Flag{
				ExploitId: r.Exploit.Id,
				TargetId:  r.Target.Id,
				Result:    result,
				Valid:     flagString != "",
				Flag:      flagString,
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
		ExploitAdder:   make(chan *model.Exploit, 5),
		ExploitRemover: make(chan *model.Exploit, 5),
		targets:        make(map[int64]*model.Target),
		TargetAdder:    make(chan *model.Target, 5),
		TargetRemover:  make(chan *model.Target, 5),
		runner:         make(map[int64]map[int64]*Runner),
		runnerLock:     sync.RWMutex{},
		Flagger:        make(chan *model.Flag, 30),
		controller:     controller,
	}
}

func (er *ExploitRunner) addExploit(exploit *model.Exploit) {
	er.runnerLock.RLock()
	defer er.runnerLock.RUnlock()

	er.exploits[exploit.Id] = exploit
	er.runner[exploit.Id] = make(map[int64]*Runner)

	for _, target := range er.targets {
		if exploit.Tag == "" || target.Tag == "" || exploit.Tag == target.Tag {
			er.runner[exploit.Id][target.Id] = er.NewRunner(exploit, target)
			go er.runner[exploit.Id][target.Id].Run()
		}
	}

	log.Default().Println("ExploitRunner added exploit", exploit.Name)
}

func (er *ExploitRunner) removeExploit(exploit *model.Exploit) {
	er.runnerLock.Lock()
	defer er.runnerLock.Unlock()

	delete(er.exploits, exploit.Id)

	runners := []*Runner{}

	for _, runner := range er.runner[exploit.Id] {
		runners = append(runners, runner)
	}

	for _, runner := range runners {
		runner.Notify <- "stop"
	}

	delete(er.runner, exploit.Id)

	log.Default().Println("ExploitRunner removed exploit", exploit.Id)
}

func (er *ExploitRunner) addTarget(target *model.Target) {
	er.runnerLock.RLock()
	defer er.runnerLock.RUnlock()

	er.targets[target.Id] = target

	for _, exploit := range er.exploits {
		if exploit.Tag == "" || target.Tag == "" || exploit.Tag == target.Tag {
			er.runner[exploit.Id][target.Id] = er.NewRunner(exploit, target)
			go er.runner[exploit.Id][target.Id].Run()
		}
	}

	log.Default().Println("ExploitRunner added target", target.Name)
}

func (er *ExploitRunner) removeTarget(target *model.Target) {
	er.runnerLock.Lock()
	defer er.runnerLock.Unlock()

	delete(er.targets, target.Id)

	for id, exploit := range er.runner {
		runner := exploit[target.Id]
		runner.Notify <- "stop"
		delete(er.runner[id], runner.Target.Id)
	}

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
