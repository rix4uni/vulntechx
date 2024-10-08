## vulntechx

vulntechx finds vulnerabilities based on tech stack using nuclei tags or fuzzing with ffuf.

## Installation
```
go install github.com/rix4uni/vulntechx@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/vulntechx/releases/download/v0.0.1/vulntechx-linux-amd64-0.0.1.tgz
tar -xvzf vulntechx-linux-amd64-0.0.1.tgz
rm -rf vulntechx-linux-amd64-0.0.1.tgz
mv vulntechx ~/go/bin/vulntechx
```
Or download [binary release](https://github.com/rix4uni/vulntechx/releases) for your platform.

## Compile from source
```
git clone --depth 1 github.com/rix4uni/vulntechx.git
cd vulntechx; go install
```

## Usage
```
                __        __               __
 _   __ __  __ / /____   / /_ ___   _____ / /_   _  __
| | / // / / // // __ \ / __// _ \ / ___// __ \ | |/_/
| |/ // /_/ // // / / // /_ /  __// /__ / / / /_>  <
|___/ \__,_//_//_/ /_/ \__/ \___/ \___//_/ /_//_/|_|

                            Current vulntechx version v0.0.1

A longer description of your application.

Usage:
  vulntechx [flags]
  vulntechx [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  ffuf        A brief description of your command
  help        Help about any command
  httpxjson   Parse httpx output to extract hosts and technologies, and format as JSON.
  nuclei      Run Nuclei scans on multiple hosts in parallel, filtering by technology stack.

Flags:
  -h, --help      help for vulntechx
  -u, --update    update vulntechx to latest version
  -v, --version   Print the version of the tool and exit.

Use "vulntechx [command] --help" for more information about a command.
```

## Usage Example
```bash
# Step 1, subdomain enumeration and subdomain probing and find tech stack
subfinder -d hackerone.com -all -duc -silent | httpx -duc -silent -nc -mc 200 -t 300 -td | unew httpx.txt

# Step 2, convert httpx output to json
cat httpx.txt | vulntechx httpxjson -o httpxjson-output.json

# Step 3, find vulnerabilities based on tech using nuclei
vulntechx nuclei --file httpxjson-output.json --nucleicmd "nuclei -tags {tech} -es unknown,info,low" --parallel 10 --process --append nuclei-output.txt

# Step 4, find vulnerabilities based on tech using fuzzing with ffuf
```