# Python Runner

One of many Python wrappers on the internet. Its only purpose is to simplify the local Python development process.

*Before:*

```
python -m venv .venv
./.venv/scripts/activate
python app.py
```

*After:*

```
pyrun app.py
```

*Python Runner* uses the `python` and `pip` executables from an existing virtual environment (`venv/`). If the directory does not exist, then it creates one with `venv`. Dependencies are thereafter installed if `requirements.txt` is present.

## Installation
The easiest way to install `pyrun` is by downloading the executable for your platform from the [releases](https://github.com/asiangoldfish/python-runner/releases). After downloading the file, rename it to `pyrun` and add it to PATH.

Alternatively, it can be installed with the Go tooling:

```
go install github.com/asiangoldfish/pyrun@latest
```

## Usage
To build and use `pyrun`:
1. The project is built with Go version 1.18.
2. Build the executable:
   ```
   go build -o bin/
   ```

To execute the program as a command, you can move the executable (`pyrun` or `pyrun.exe` depending on the target platform) to a directory in PATH.
