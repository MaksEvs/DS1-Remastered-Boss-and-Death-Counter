//go:build windows

package main

import (
	"encoding/binary"
	"fmt"
	"time"
)

func startGameMonitor(applicationState *applicationState) {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			currentFlags, err := readBossFlagsFromGame()
			if err != nil {
				applicationState.updateGameStatus(false)
				continue
			}

			applicationState.updateGameStatus(true)
			applicationState.updateKilledCount(countTrueFlags(currentFlags))
			if currentDeaths, err := readGameDeathCount(); err == nil {
				applicationState.observeGameDeaths(currentDeaths)
			}
			_ = saveApplicationState(applicationState)
		}
	}()
}

func (s *applicationState) observeGameDeaths(current uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.HasGameDeathBaseline && current == s.LastGameDeaths+1 {
		s.Deaths++
	}
	s.LastGameDeaths = current
	s.HasGameDeathBaseline = true
}

func readGameDeathCount() (uint32, error) {
	processID, err := findGameProcessID()
	if err != nil {
		return 0, err
	}

	reader, err := connectToGame(processID)
	if err != nil {
		return 0, err
	}
	defer reader.close()

	// Locate GameDataMan through its code signature.
	pattern := []byte{0x48, 0x8B, 0x05, 0, 0, 0, 0, 0x48, 0x8B, 0x50, 0x10, 0x48, 0x89, 0x54, 0x24, 0x60}
	wildcards := map[int]bool{
		3: true,
		4: true,
		5: true,
		6: true,
	}
	instructionAddress, err := reader.scan(pattern, wildcards)
	if err != nil {
		return 0, fmt.Errorf("GameDataMan signature: %w", err)
	}

	displacementBytes, err := reader.read(instructionAddress+3, 4)
	if err != nil {
		return 0, err
	}
	displacement := int64(int32(binary.LittleEndian.Uint32(displacementBytes)))
	globalPointerAddress := uintptr(int64(instructionAddress+7) + displacement)
	gameDataManagerBytes, err := reader.read(globalPointerAddress, 8)
	if err != nil {
		return 0, err
	}
	gameDataManagerAddress := uintptr(binary.LittleEndian.Uint64(gameDataManagerBytes))
	if gameDataManagerAddress < 0x10000 {
		return 0, fmt.Errorf("invalid GameDataMan pointer 0x%x", gameDataManagerAddress)
	}

	deathCountBytes, err := reader.read(gameDataManagerAddress+0x98, 4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(deathCountBytes), nil
}

func readBossFlagsFromGame() (map[int]bool, error) {
	processID, err := findGameProcessID()
	if err != nil {
		return nil, err
	}

	reader, err := connectToGame(processID)
	if err != nil {
		return nil, err
	}
	defer reader.close()

	pattern := []byte{0x48, 0x8B, 0x0D, 0, 0, 0, 0, 0x99, 0x33, 0xC2, 0x45, 0x33, 0xC0, 0x2B, 0xC2, 0x8D, 0x50, 0xF6}
	wildcards := map[int]bool{3: true, 4: true, 5: true, 6: true}
	instructionAddress, err := reader.scan(pattern, wildcards)
	if err != nil {
		return nil, err
	}

	rel32Bytes, err := reader.read(instructionAddress+3, 4)
	if err != nil {
		return nil, err
	}
	rel32 := int64(int32(binary.LittleEndian.Uint32(rel32Bytes)))
	staticPointerAddress := uintptr(int64(instructionAddress+7) + rel32)

	staticPointerBytes, err := reader.read(staticPointerAddress, 8)
	if err != nil {
		return nil, err
	}
	eventManagerAddress := uintptr(binary.LittleEndian.Uint64(staticPointerBytes))

	managerBytes, err := reader.read(eventManagerAddress, 4)
	if err != nil {
		return nil, err
	}
	flagArrayAddress := uintptr(binary.LittleEndian.Uint32(managerBytes))

	flags := make(map[int]bool, len(bossFlags))
	for _, boss := range bossFlags {
		isSet, err := readEventFlag(reader, flagArrayAddress, boss.ID)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", boss.Name, err)
		}
		flags[boss.ID] = isSet
	}

	return flags, nil
}

func countTrueFlags(flags map[int]bool) int {
	count := 0
	for _, isSet := range flags {
		if isSet {
			count++
		}
	}
	return count
}
