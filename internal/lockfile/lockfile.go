package lockfile

import (
	"encoding/json"
	"os"
	"time"

	"github.com/wayyoungboy/mcpcanary/internal/model"
	"github.com/wayyoungboy/mcpcanary/internal/semantic"
)

const ChangeAdded = "added"
const ChangeRemoved = "removed"
const ChangeDrifted = "drifted"

type Lockfile struct {
	SchemaVersion string       `json:"schema_version"`
	GeneratedAt   time.Time    `json:"generated_at"`
	Servers       []LockServer `json:"servers"`
}

type LockServer struct {
	Name        string    `json:"name"`
	Command     string    `json:"command,omitempty"`
	Package     string    `json:"package,omitempty"`
	Source      string    `json:"source,omitempty"`
	Fingerprint string    `json:"fingerprint"`
	Vector      []float64 `json:"vector"`
	RiskScore   int       `json:"risk_score"`
}

type CompareOptions struct {
	DriftThreshold float64
}

type Diff struct {
	Changes []Change `json:"changes"`
}

type Change struct {
	Kind       string  `json:"kind"`
	ServerName string  `json:"server_name"`
	Distance   float64 `json:"distance,omitempty"`
}

func FromServers(servers []model.Server, riskScore int) Lockfile {
	locked := make([]LockServer, 0, len(servers))
	for _, server := range servers {
		locked = append(locked, LockServer{
			Name:        server.Name,
			Command:     server.Command,
			Package:     server.Package,
			Source:      server.Source,
			Fingerprint: server.Fingerprint(),
			Vector:      semantic.Embed(server.DescriptorText(), 128),
			RiskScore:   riskScore,
		})
	}
	return Lockfile{SchemaVersion: model.SchemaVersion, GeneratedAt: time.Now().UTC(), Servers: locked}
}

func Write(path string, lock Lockfile) error {
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func Read(path string) (Lockfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Lockfile{}, err
	}
	var lock Lockfile
	if err := json.Unmarshal(data, &lock); err != nil {
		return Lockfile{}, err
	}
	return lock, nil
}

func Compare(lock Lockfile, current []model.Server, opts CompareOptions) Diff {
	threshold := opts.DriftThreshold
	if threshold <= 0 {
		threshold = 0.18
	}
	locked := make(map[string]LockServer, len(lock.Servers))
	for _, server := range lock.Servers {
		locked[server.Name] = server
	}
	seen := make(map[string]bool, len(current))
	var changes []Change
	for _, server := range current {
		seen[server.Name] = true
		prior, ok := locked[server.Name]
		if !ok {
			changes = append(changes, Change{Kind: ChangeAdded, ServerName: server.Name})
			continue
		}
		currentVector := semantic.Embed(server.DescriptorText(), 128)
		distance := 1 - semantic.Cosine(prior.Vector, currentVector)
		if prior.Fingerprint != server.Fingerprint() && distance >= threshold {
			changes = append(changes, Change{Kind: ChangeDrifted, ServerName: server.Name, Distance: distance})
		}
	}
	for _, prior := range lock.Servers {
		if !seen[prior.Name] {
			changes = append(changes, Change{Kind: ChangeRemoved, ServerName: prior.Name})
		}
	}
	return Diff{Changes: changes}
}
