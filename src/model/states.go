package model

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/adax/adax-api"
	"github.com/thingsplex/adax/utils"
)

type States struct {
	path         string
	WorkDir      string `json:"-"`
	ConfiguredAt string `json:"configuret_at"`
	ConfiguredBy string `json:"configures_by"`

	HomesAndRooms *adax.HomesAndRooms `json:"homesAndRooms"`
	States        *adax.State         `json:"states"`
}

func NewStates(workDir string) *States {
	state := &States{WorkDir: workDir}
	//state.path = filepath.Join(workDir, "data", "state.json")
	//if !utils.FileExists(state.path) {
	//	log.Info("State file doesn't exist.Loading default state")
	//	defaultStateFile := filepath.Join(workDir, "defaults", "state.json")
	//	err := utils.CopyFile(defaultStateFile, state.path)
	//	if err != nil {
	//		fmt.Print(err)
	//		panic("Can't copy state file.")
	//	}
	//}
	return state
}

func (st *States) LoadFromFile() error {
	stateFileBody, err := ioutil.ReadFile(st.path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(stateFileBody, st)
	if err != nil {
		return err
	}
	return nil
}

func (st *States) SaveToFile() error {
	st.ConfiguredBy = "auto"
	st.ConfiguredAt = time.Now().Format(time.RFC3339)
	bpayload, err := json.Marshal(st)
	err = ioutil.WriteFile(st.path, bpayload, 0664)
	if err != nil {
		return err
	}
	return err
}

func (st *States) LoadDefaults() error {
	stateFile := filepath.Join(st.WorkDir, "data", "state.json")
	os.Remove(stateFile)
	log.Info("State file doesn't exist.Loading default state")
	defaultStateFile := filepath.Join(st.WorkDir, "defaults", "state.json")
	return utils.CopyFile(defaultStateFile, stateFile)
}

func (st *States) IsConfigured() bool {
	if true {
		return true
	}
	return false
}

type StateReport struct {
	OpStatus string    `json:"op_status"`
	AppState AppStates `json:"app_state"`
}
