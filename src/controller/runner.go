package controller

import (
	"bytes"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/hiumee/addack/src/model"
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
	Notify         chan string
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
	r.Controller.Logger.Println("Runner started", r.Exploit.Name, r.Target.Name)
	tickTicker := time.NewTicker(time.Duration(r.Controller.Config.TickTime) * time.Millisecond)
	defer tickTicker.Stop()
	for {
		var output bytes.Buffer

		cmd := exec.Command("bash", "-c", r.Exploit.Command)
		cmd.Dir = r.Controller.Config.ExploitsPath + r.Exploit.Path
		cmd.Env = append(cmd.Env, "TARGET="+r.Target.Ip)
		writer := io.Writer(&output)
		cmd.Stdout = writer
		cmd.Stderr = writer

		if err := cmd.Start(); err != nil {
			r.Controller.Logger.Println("Runner error", r.Exploit.Name, r.Target.Name, err)
			<-tickTicker.C
			continue
		}
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		timer := time.NewTimer(time.Duration(r.Exploit.Timeout) * time.Millisecond)
		select {
		case <-r.Notify:
			r.Controller.Logger.Println("Runner stopped", r.Exploit.Name, r.Target.Name)
			cmd.Process.Kill()
			close(r.Notify)
			timer.Stop()
			return
		case <-done:
			result := string(output.Bytes())

			flagStrings := r.Controller.Config.FlagRegex.FindAllString(result, 100)

			for _, flagString := range flagStrings {
				r.Controller.Logger.Println("Runner result", r.Exploit.Name, r.Target.Name, flagString)

				var validation string
				if flagString != "" {
					validation = "matched"
				} else {
					validation = "not matched"
				}

				flag := &model.Flag{
					ExploitId: r.Exploit.Id,
					TargetId:  r.Target.Id,
					Result:    result,
					Valid:     validation,
					Flag:      flagString,
				}
				r.Flagger <- flag
			}
		case <-timer.C:
			<-tickTicker.C
		}

		timer.Stop()

		<-tickTicker.C
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
		Flagger:        make(chan *model.Flag, 3000),
		controller:     controller,
		Notify:         make(chan string),
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

	er.controller.Logger.Println("ExploitRunner added exploit", exploit.Name)
}

func (er *ExploitRunner) removeExploit(exploit *model.Exploit) {
	er.runnerLock.Lock()
	defer er.runnerLock.Unlock()

	if _, ok := er.exploits[exploit.Id]; !ok {
		return
	}

	delete(er.exploits, exploit.Id)

	runners := []*Runner{}

	for _, runner := range er.runner[exploit.Id] {
		runners = append(runners, runner)
	}

	for _, runner := range runners {
		runner.Notify <- "stop"
	}

	delete(er.runner, exploit.Id)

	er.controller.Logger.Println("ExploitRunner removed exploit", exploit.Id)
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

	er.controller.Logger.Println("ExploitRunner added target", target.Name)
}

func (er *ExploitRunner) removeTarget(target *model.Target) {
	er.runnerLock.Lock()
	defer er.runnerLock.Unlock()

	if _, ok := er.targets[target.Id]; !ok {
		return
	}
	delete(er.targets, target.Id)

	for id, exploit := range er.runner {
		runner := exploit[target.Id]
		runner.Notify <- "stop"
		delete(er.runner[id], runner.Target.Id)
	}

	er.controller.Logger.Println("ExploitRunner removed target", target.Id)
}

func (er *ExploitRunner) Run() {
	er.controller.Logger.Println("ExploitRunner started")
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
			id, err := er.controller.DB.CreateFlag(*flag)
			if err != nil {
				er.controller.Logger.Println("ExploitRunner error", "Could not save flag", err)
				continue
			}
			if er.controller.Config.FlaggerCommand != "" && flag.Valid == "matched" {
				flag.Id = id
				go SendFlag(*flag, er.controller)
			}
		case <-er.Notify:
			er.controller.Logger.Println("ExploitRunner stopped")
			return
		}
	}
}

func (er *ExploitRunner) Stop() {
	for _, exploit := range er.exploits {
		for _, runner := range er.runner[exploit.Id] {
			runner.Notify <- "stop"
		}
	}
	er.Notify <- "stop"
}
