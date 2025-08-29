package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// --- Container Kernel Framework ---

type ProcessState int

const (
	Running ProcessState = iota
	Stopped
	Completed
)

type Process struct {
	Name     string
	Priority int
	Action   func()
	State    ProcessState
}

type Container struct {
	ID        string
	Name      string
	MemoryMB  int
	CPULoad   float64
	Processes []*Process
	mu        sync.Mutex
}

func (c *Container) AddProcess(p *Process) {
	c.mu.Lock()
	defer c.mu.Unlock()
	p.State = Running
	c.Processes = append(c.Processes, p)
}

func (c *Container) StartProcesses() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, p := range c.Processes {
		if p.State == Running {
			go func(proc *Process) {
				proc.Action()
				proc.State = Completed
			}(p)
		}
	}
}

func (c *Container) StopProcesses() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, p := range c.Processes {
		if p.State == Running {
			p.State = Stopped
		}
	}
}

// --- Kernel ---
type Kernel struct {
	Containers map[string]*Container
	mu         sync.Mutex
}

func NewKernel() *Kernel {
	return &Kernel{
		Containers: make(map[string]*Container),
	}
}

func (k *Kernel) CreateContainer(id, name string, memory int) *Container {
	k.mu.Lock()
	defer k.mu.Unlock()
	c := &Container{
		ID:        id,
		Name:      name,
		MemoryMB:  memory,
		CPULoad:   0,
		Processes: []*Process{},
	}
	k.Containers[id] = c
	fmt.Printf("[Kernel] Created container: %s\n", name)
	return c
}

func (k *Kernel) StartAll() {
	k.mu.Lock()
	defer k.mu.Unlock()
	for _, c := range k.Containers {
		fmt.Printf("[Kernel] Starting container: %s\n", c.Name)
		c.StartProcesses()
	}
}

func (k *Kernel) StopAll() {
	k.mu.Lock()
	defer k.mu.Unlock()
	for _, c := range k.Containers {
		fmt.Printf("[Kernel] Stopping container: %s\n", c.Name)
		c.StopProcesses()
	}
}

func (k *Kernel) Monitor(interval time.Duration, cycles int) {
	for i := 0; i < cycles; i++ {
		fmt.Println("=== Kernel Monitoring ===")
		k.mu.Lock()
		for _, c := range k.Containers {
			active := 0
			for _, p := range c.Processes {
				if p.State == Running {
					active++
				}
			}
			fmt.Printf("Container %s | Memory: %dMB | CPU: %.2f%% | Running Processes: %d\n",
				c.Name, c.MemoryMB, c.CPULoad, active)
		}
		k.mu.Unlock()
		time.Sleep(interval)
	}
}

// Inter-container messaging
func (k *Kernel) SendMessage(fromID, toID, msg string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	from, ok1 := k.Containers[fromID]
	to, ok2 := k.Containers[toID]
	if ok1 && ok2 {
		fmt.Printf("[Kernel] %s -> %s : %s\n", from.Name, to.Name, msg)
	} else {
		fmt.Println("[Kernel] Messaging error: container not found")
	}
}

// --- Example Process ---
func exampleProcess(name string, duration time.Duration) *Process {
	return &Process{
		Name:     name,
		Priority: rand.Intn(10),
		Action: func() {
			fmt.Printf("Process %s started\n", name)
			time.Sleep(duration)
			fmt.Printf("Process %s completed\n", name)
		},
		State: Running,
	}
}

// --- Main ---
func main() {
	kernel := NewKernel()

	// Create containers
	c1 := kernel.CreateContainer("c1", "WebServer", 512)
	c2 := kernel.CreateContainer("c2", "Database", 1024)

	// Add processes
	c1.AddProcess(exampleProcess("HTTP Server", 2*time.Second))
	c1.AddProcess(exampleProcess("Worker", 3*time.Second))
	c2.AddProcess(exampleProcess("DB Engine", 4*time.Second))
	c2.AddProcess(exampleProcess("Backup", 5*time.Second))

	// Start all containers
	kernel.StartAll()

	// Inter-container messaging
	kernel.SendMessage("c1", "c2", "Query: SELECT * FROM users;")
	kernel.SendMessage("c2", "c1", "Response: 42 records returned.")

	// Dynamic CPU/Memory simulation
	go func() {
		for i := 0; i < 5; i++ {
			kernel.mu.Lock()
			for _, c := range kernel.Containers {
				c.CPULoad = rand.Float64() * 100
				c.MemoryMB = c.MemoryMB + rand.Intn(50) - 25
			}
			kernel.mu.Unlock()
			time.Sleep(1 * time.Second)
		}
	}()

	// Monitor kernel for 5 cycles
	kernel.Monitor(1*time.Second, 5)

	// Stop all containers
	kernel.StopAll()
	fmt.Println("[Kernel] All containers stopped.")
}
