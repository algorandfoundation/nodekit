package system

type Interface interface {
	IsInstalled() bool
	IsRunning() bool
	IsService() bool
	SetNetwork(network string) error
	Install() error
	Update() error
	Uninstall() error
	Start() error
	Stop() error
	Restart() error
	UpdateService(dataDirectoryPath string) error
	EnsureService() error
}
