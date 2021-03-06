package helpers

import (
	"io/ioutil"
	"strconv"
	"os"
	"syscall"
	"os/exec"
	"log"
	"io"
	"github.com/PapiCZ/github-notifier/settings"
)

type Command struct {
	homeDir string
	srcRoot string
	goPath  string
}

func copyFile(src string, dest string) {
	from, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	defer from.Close()

	to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		panic(err)
	}
}

func NewCommand(homeDir string, goPath string, srcRoot string) (*Command) {
	command := Command{
		homeDir: homeDir,
		srcRoot: srcRoot,
		goPath:  goPath,
	}

	return &command
}

func (c *Command) Install() {
	// Prepare directory for config file
	configErr := os.MkdirAll(c.homeDir+settings.ConfigPath, 0700)

	if configErr != nil {
		panic(configErr)
	}

	// Prepare directory for static data
	octocatErr := os.MkdirAll(c.homeDir+settings.DataPath, 0700)
	if octocatErr != nil {
		panic(configErr)
	}

	// Copy config file
	copyFile(c.goPath+c.srcRoot+"/config.json.example", c.homeDir+settings.ConfigPath+"/"+settings.ConfigFileName)

	// Copy octocat
	copyFile(c.goPath+c.srcRoot+"/icons/octocat.png", c.homeDir+settings.DataPath+"/"+settings.IconFileName)
}

func (c *Command) Start(pidFileName string) {
	// Check if app doesn't already running
	if _, err := os.Stat(c.homeDir + "/" + pidFileName); err == nil {
		panic("GitHub notifier already running")
	}

	// Run this app again
	cmd := exec.Command(os.Args[0], os.Args[2:]...)

	curDir, osErr := os.Getwd()
	if osErr != nil {
		panic(osErr)
	}

	cmd.Dir = curDir
	cmd.Start()

	// Store PID in file
	pidFileErr := ioutil.WriteFile(c.homeDir+"/"+pidFileName, []byte(strconv.Itoa(cmd.Process.Pid)), 0600)
	if pidFileErr != nil {
		cmd.Process.Kill()
		panic(pidFileErr)
	}

	cmd.Process.Release()
}

func (c *Command) Stop(pidFileName string) {
	// Read PID file
	content, err := ioutil.ReadFile(c.homeDir + "/" + pidFileName)
	if err != nil {
		panic(err)
	}

	pid, err := strconv.Atoi(string(content))
	if err != nil {
		panic(err)
	}

	// Kill with PID from PID file
	syscall.Kill(pid, 15)

	// Remove PID file
	removeErr := os.Remove(c.homeDir + "/" + settings.PidFileName)
	if removeErr != nil {
		panic(removeErr)
	}
}
