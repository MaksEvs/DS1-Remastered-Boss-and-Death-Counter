//go:build windows

package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
)

type bossFlag struct {
	Name string
	ID   int
}

var bossFlags = []bossFlag{
	{"Gaping Dragon", 2}, {"Bell Gargoyles", 3}, {"Crossbreed Priscilla", 4}, {"Great Grey Wolf Sif", 5},
	{"Pinwheel", 6}, {"Gravelord Nito", 7}, {"Chaos Witch Quelaag", 9}, {"Bed of Chaos", 10},
	{"Iron Golem", 11}, {"Ornstein and Smough", 12}, {"Four Kings", 13}, {"Seath the Scaleless", 14},
	{"Gwyn, Lord of Cinder", 15}, {"Asylum Demon", 16}, {"Taurus Demon", 11010901}, {"Capra Demon", 11010902},
	{"Moonlight Butterfly", 11200900}, {"Sanctuary Guardian", 11210000}, {"Knight Artorias", 11210001},
	{"Manus, Father of the Abyss", 11210002}, {"Black Dragon Kalameet", 11210004}, {"Demon Firesage", 11410410},
	{"Ceaseless Discharge", 11410900}, {"Centipede Demon", 11410901}, {"Dark Sun Gwyndolin", 11510900}, {"Stray Demon", 11810900},
}

func runDiagnostic() error {
	processID, err := findGameProcessID()
	if err != nil {
		return err
	}
	reader, err := connectToGame(processID)
	if err != nil {
		return err
	}
	defer reader.close()

	pattern := []byte{0x48, 0x8B, 0x0D, 0, 0, 0, 0, 0x99, 0x33, 0xC2, 0x45, 0x33, 0xC0, 0x2B, 0xC2, 0x8D, 0x50, 0xF6}
	wildcards := map[int]bool{3: true, 4: true, 5: true, 6: true}
	instructionAddress, err := reader.scan(pattern, wildcards)
	if err != nil {
		return fmt.Errorf("module base 0x%x: %w", reader.moduleBase, err)
	}

	rel32Bytes, err := reader.read(instructionAddress+3, 4)
	if err != nil {
		return err
	}
	rel32 := int64(int32(binary.LittleEndian.Uint32(rel32Bytes)))
	staticPointerAddress := uintptr(int64(instructionAddress+7) + rel32)
	staticPointerBytes, err := reader.read(staticPointerAddress, 8)
	if err != nil {
		return err
	}
	eventManagerAddress := uintptr(binary.LittleEndian.Uint64(staticPointerBytes))
	managerBytes, err := reader.read(eventManagerAddress, 4)
	if err != nil {
		return err
	}
	flagArrayAddress := uintptr(binary.LittleEndian.Uint32(managerBytes))

	fmt.Printf("Connected to DarkSoulsRemastered.exe (PID %d)\n", processID)
	fmt.Printf("Module base: 0x%x\n", reader.moduleBase)
	fmt.Printf("Event manager: 0x%x\n", eventManagerAddress)
	fmt.Printf("Flag array: 0x%x\n\n", flagArrayAddress)

	killedCount := 0
	for _, boss := range bossFlags {
		killed, err := readEventFlag(reader, flagArrayAddress, boss.ID)
		if err != nil {
			return fmt.Errorf("read %s: %w", boss.Name, err)
		}
		if killed {
			killedCount++
		}
		fmt.Printf("%-28s %d  %t\n", boss.Name, boss.ID, killed)
	}
	fmt.Printf("\nResult: %d/26\n", killedCount)
	return nil
}

func findGameProcessID() (uint32, error) {
	return findProcessIDByName("DarkSoulsRemastered.exe")
}

func readEventFlag(reader *processMemoryReader, flagArrayAddress uintptr, flagID int) (bool, error) {
	byteOffset, bitIndex, err := eventFlagOffset(flagID)
	if err != nil {
		return false, err
	}
	data, err := reader.read(flagArrayAddress+uintptr(byteOffset), 1)
	if err != nil {
		return false, err
	}
	return data[0]&(1<<bitIndex) != 0, nil
}

func eventFlagOffset(flagID int) (int, uint, error) {
	id := fmt.Sprintf("%08d", flagID)
	if len(id) != 8 {
		return 0, 0, fmt.Errorf("unsupported flag ID %d", flagID)
	}
	primaryOffsets := map[byte]int{'0': 0x00000, '1': 0x00500, '5': 0x05F00, '6': 0x0B900, '7': 0x11300}
	secondaryIndexes := map[string]int{"000": 0, "100": 1, "101": 2, "102": 3, "110": 4, "120": 5, "121": 6, "130": 7, "131": 8, "132": 9, "140": 10, "141": 11, "150": 12, "151": 13, "160": 14, "170": 15, "180": 16, "181": 17}
	primary, ok := primaryOffsets[id[0]]
	if !ok {
		return 0, 0, fmt.Errorf("unsupported primary flag range %d", flagID)
	}
	secondaryIndex, ok := secondaryIndexes[id[1:4]]
	if !ok {
		return 0, 0, fmt.Errorf("unsupported secondary flag range %d", flagID)
	}
	tail, _ := strconv.Atoi(id[5:8])
	mask := uint32(0x80000000) >> uint(tail%32)
	significantByte := 0
	for mask>>(uint(significantByte)*8)&0xFF == 0 {
		significantByte++
	}
	byteValue := byte(mask >> (uint(significantByte) * 8))
	bitIndex := uint(0)
	for byteValue&(1<<bitIndex) == 0 {
		bitIndex++
	}
	offset := primary + secondaryIndex*1280 + int(id[4]-'0')*128 + (tail-tail%32)/8 + significantByte
	return offset, bitIndex, nil
}

func diagnosticMain() {
	if err := runDiagnostic(); err != nil {
		fmt.Fprintln(os.Stderr, "Diagnostic failed:", err)
		os.Exit(1)
	}
}
