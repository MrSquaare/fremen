package tests

type Report struct {
	Configuration Configuration `json:"configuration"`
	Results       []Result      `json:"results"`
	Summary       Summary       `json:"summary"`
}

type Configuration struct {
	Paths              []string `json:"paths"`
	Database           string   `json:"database"`
	Recursive          bool     `json:"recursive"`
	IncludeGit         bool     `json:"include_git"`
	IncludeNodeModules bool     `json:"include_node_modules"`
	ExcludeRegex       string   `json:"exclude_regex"`
	FullReport         bool     `json:"full_report"`
}

type Result struct {
	InfectedCount    int               `json:"infected_count"`
	InfectedPackages []InfectedPackage `json:"infected_packages"`
	Lockfiles        []string          `json:"lockfiles"`
	Project          string            `json:"project"`
}

type InfectedPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Summary struct {
	TotalProjects         int `json:"total_projects"`
	InfectedProjects      int `json:"infected_projects"`
	TotalInfectedPackages int `json:"total_infected_packages"`
}
