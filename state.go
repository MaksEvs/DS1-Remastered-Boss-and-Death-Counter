package main

import (
	"fmt"
	"strings"
	"sync"
)

const totalBossCount = 26

type applicationState struct {
	mu                   sync.RWMutex
	Label                string `json:"label"`
	Killed               int    `json:"killed"`
	Deaths               int    `json:"deaths"`
	LastGameDeaths       uint32 `json:"lastGameDeaths"`
	HasGameDeathBaseline bool   `json:"hasGameDeathBaseline"`
	GameConnected        bool   `json:"connected"`
}

func newApplicationState() *applicationState {
	return &applicationState{Label: "Боссов убито"}
}

func (s *applicationState) deathText() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return fmt.Sprintf("Смерти: %d", s.Deaths)
}

func (s *applicationState) obsText() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	label := strings.TrimSpace(s.Label)
	if label != "" {
		label += ": "
	}

	return fmt.Sprintf("%s%d/%d", label, s.Killed, totalBossCount)
}

func (s *applicationState) snapshot() stateSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return stateSnapshot{
		Label:                s.Label,
		Killed:               s.Killed,
		Deaths:               s.Deaths,
		LastGameDeaths:       s.LastGameDeaths,
		HasGameDeathBaseline: s.HasGameDeathBaseline,
		Total:                totalBossCount,
		GameConnected:        s.GameConnected,
	}
}

func (s *applicationState) updateKilledCount(killed int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Killed = killed
}

func (s *applicationState) updateGameStatus(connected bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.GameConnected = connected
}

type stateSnapshot struct {
	Label                string `json:"label"`
	Killed               int    `json:"killed"`
	Total                int    `json:"total"`
	Deaths               int    `json:"deaths"`
	LastGameDeaths       uint32 `json:"lastGameDeaths"`
	HasGameDeathBaseline bool   `json:"hasGameDeathBaseline"`
	GameConnected        bool   `json:"connected"`
}
