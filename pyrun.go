package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing script name")
		return
	}

	venvDir := ".venv"
	reqFile := "requirements.txt"
	_, err := os.Stat(venvDir)
	if err != nil {
		// venv not found. Create it and install deps if necessary.
		fmt.Print("Project Python interpreter not found. ")
		fmt.Println("Creating virtual environment...")

		cmd := exec.Command("python", "-m", "venv", ".venv")
		if err := cmd.Run(); err != nil {
			// Python not installed?
			panic(err)
		}

		// Install deps with requirements.txt
		fmt.Println("Installing dependencies...")

		_, err = os.Stat(reqFile)
		cmd = exec.Command(venvDir+"/Scripts/pip.exe", "install", "-r", reqFile)
		if err == nil {
			// requirements.txt was found. Install deps.
			if err := cmd.Run(); err != nil {
				panic(err)
			}
		}
	}

	// TODO check if venv is a file. If so, abort.

	scriptName := os.Args[1]

	cmd := exec.Command(
		venvDir+"/Scripts/python.exe",
		scriptName,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
