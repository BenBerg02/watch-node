package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	process         []processInfo
	status          string
	prevProcessInfo map[int]processInfo
	prevSystemCPU   float64

	validKill bool
	cursor    int
}

type processInfo struct {
	PID  int
	Name string
	RAM  int
	CPU  float64

	cpuTime float64
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "u", "up":
			{
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.process) - 1
				}
			}
		case "j", "down":
			m.cursor++
			if m.cursor > len(m.process)-1 {
				m.cursor = 0
			}
		case " ":
			if len(m.process) == 0 {
				break
			}
			process := m.process[m.cursor]
			proc, err := os.FindProcess(process.PID)
			if err != nil {
				break
			}

			err = proc.Kill()
			if err != nil {
				break
			}
		}

	case tickMsg:
		m.process = m.getProcess()
		m.status = fmt.Sprintf("Time: %s", time.Now().Format("15:04:05"))
		return m, tickCmd()
	}
	return m, nil
}

func (m model) View() string {

	var s strings.Builder

	s.WriteString(m.status + "\n\n")

	s.WriteString(fmt.Sprintf("    %-10s %-40s %-10s %s\n", "PID", "Name", "CPU(%)", "RAM (kB)"))
	s.WriteString("------------------------------------------------------------------------------\n")

	for i, p := range m.process {
		const maxNameLength = 40
		truncatedName := p.Name
		if len(truncatedName) > maxNameLength {
			truncatedNameIf := truncatedName[maxNameLength-15:]
			if len(truncatedNameIf) < len(truncatedName) {
				truncatedName = "..." + truncatedNameIf
			}
		}

		var check string

		if i == m.cursor {
			check = "x"
		} else {
			check = " "
		}

		s.WriteString(fmt.Sprintf("[%s] %-10d %-40s %-10.2f %d\n", check, p.PID, truncatedName, p.CPU, p.RAM))
	}
	return s.String()
}

func getSysyemCPUInfo() (float64, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	cpuStats := strings.Fields(lines[0])
	totalTime := 0.0
	for i := 1; i < len(cpuStats); i++ {
		val, err := strconv.ParseFloat(cpuStats[i], 64)
		if err == nil {
			totalTime += val
		}
	}
	return totalTime, nil
}

func (m *model) getProcess() []processInfo {
	var process []processInfo

	seenPids := make(map[int]bool)

	files, err := os.ReadDir("/proc")
	if err != nil {
		log.Fatal(err)
	}

	systemCPUTime, err := getSysyemCPUInfo()

	if err != nil {
		m.status = fmt.Sprintf("Error reading system CPU info: %v", err)
		return nil
	}

	for _, file := range files {
		if file.IsDir() {
			pid, err := strconv.Atoi(file.Name())

			if err == nil {
				if !seenPids[pid] {
					cmdline := fmt.Sprintf("/proc/%d/cmdline", pid)
					cmdlineRaw, err := os.ReadFile(cmdline)
					if err != nil {
						continue
					}
					args := strings.Split(string(cmdlineRaw), "\x00")
					if len(args) == 0 {
						continue
					}
					initCommand := args[0]
					pathInit := strings.Split(initCommand, "/")
					command := pathInit[len(pathInit)-1]
					if command != "node" {
						continue
					}

					statusfile := fmt.Sprintf("/proc/%d/status", pid)
					data, err := os.ReadFile(statusfile)

					ramUsage := 0

					if err != nil {
						continue
					}

					lines := strings.Split(string(data), "\n")

					for _, line := range lines {
						if strings.HasPrefix(line, "VmRSS:") {
							fields := strings.Fields(line)
							if len(fields) >= 2 {
								ramUsage, _ = strconv.Atoi(fields[1])
								break
							}
						}
					}

					cpuPercent := 0.0

					statFile := fmt.Sprintf("/proc/%d/stat", pid)

					statData, err := os.ReadFile(statFile)

					if err == nil {
						fields := strings.Fields(string(statData))

						if len(fields) >= 22 {
							utime, _ := strconv.ParseFloat(fields[13], 64)
							stime, _ := strconv.ParseFloat(fields[14], 64)

							// startTime, _ := strconv.Atoi(fields[21])
							currentCPUTime := utime + stime

							if prev, ok := m.prevProcessInfo[pid]; ok && m.prevSystemCPU > 0 {
								diffCPUTime := currentCPUTime - prev.cpuTime
								diffSystemTime := systemCPUTime - m.prevSystemCPU

								if diffSystemTime > 0 {
									cpuPercent = (diffCPUTime / diffSystemTime) * 100
								}
							}

							if m.prevProcessInfo == nil {
								m.prevProcessInfo = make(map[int]processInfo)
							}
							m.prevProcessInfo[pid] = processInfo{
								cpuTime: currentCPUTime,
								// startTime: startTime,
							}

						}
					}

					process = append(process, processInfo{
						PID:  pid,
						Name: args[2],
						RAM:  ramUsage,
						CPU:  cpuPercent,
					})

				}
			}
		}
	}

	m.prevSystemCPU = systemCPUTime

	return process
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initialModel() model {
	return model{
		status:    "starting",
		cursor:    0,
		validKill: false,
	}
}

func cleanScreen() {
	cmd := exec.Command("clear")

	cmd.Stdout = os.Stdout

	cmd.Run()
}

func main() {

	p := tea.NewProgram(initialModel())

	cleanScreen()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error init watch-node: %v", err)
		os.Exit(1)
	}

}
