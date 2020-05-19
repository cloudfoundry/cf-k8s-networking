package checker

import (
	"encoding/json"
	"os/exec"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
)

type Upgrade struct {
	PollInterval time.Duration

	upgradeName string
	stopChan    chan bool
	upgradeDone bool
	waitGroup   sync.WaitGroup
	mutex       sync.Mutex
}

type kappResponse struct {
	Tables []struct {
		Rows []struct {
			Name       string `json:"name"`
			FinishedAt string `json:"finished_at"`
		}
	}
}

func (u *Upgrade) Start() {
	u.stopChan = make(chan bool)

	u.waitGroup.Add(1)
	go func() {
		for {
			select {
			case <-u.stopChan:
				u.waitGroup.Done()
				return
			case <-time.After(u.PollInterval):
				u.checkUpgrade()
			}
		}
	}()
}

func (u *Upgrade) Stop() {
	close(u.stopChan)
	u.waitGroup.Wait()
}

func (u *Upgrade) HasFoundUpgrade() bool {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	return u.upgradeName != ""
}

func (u *Upgrade) IsUpgradeFinished() bool {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	return u.upgradeDone
}

func (u *Upgrade) checkUpgrade() {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.upgradeName == "" {
		u.upgradeName = u.discoverUpgradeName()
	} else {
		if u.checkIfUpgradeHasFinished(u.upgradeName) {
			u.upgradeDone = true
		}
	}
}

func (u *Upgrade) checkIfUpgradeHasFinished(upgradeName string) bool {
	resp := u.kappAppChangeLS()

	for _, table := range resp.Tables {
		for _, row := range table.Rows {
			if row.Name == upgradeName {
				return row.FinishedAt != ""
			}
		}
	}

	return false
}

func (u *Upgrade) discoverUpgradeName() string {
	resp := u.kappAppChangeLS()

	for _, table := range resp.Tables {
		for _, row := range table.Rows {
			if row.FinishedAt == "" {
				return row.Name
			}
		}
	}

	return ""
}

func (u *Upgrade) kappAppChangeLS() kappResponse {
	args := []string{"app-change", "ls", "-a", "cf", "--json"}
	cmd := exec.Command("kapp", args...)

	cmd.Stderr = GinkgoWriter

	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	resp := &kappResponse{}

	err = json.Unmarshal(output, resp)
	if err != nil {
		panic(err)
	}

	return *resp
}
