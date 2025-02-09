# WebTTY

WebTTY is used to use the terminal on a web page. This service supports maintaining the terminal connection in an unstable network environment.

## Usage

After downloading the executable file according to your system, you can use the following commands:

```shell
$ webtty -h
  -command string
        Command to run
  -host string
        Host to listen on (default "localhost")
  -index-file string
        Index file, if not set, use the default index.html
  -port int
        Port to listen on (default 8080)
  -prefix-path string
        Prefix path (default "/")
  -workdir string
        Workdir for the command, default is current directory
```

Running method:

```shell
$ webtty -command bash
...
time=2025-02-09T20:07:31.819+08:00 level=INFO msg="command -> bash"
time=2025-02-09T20:07:31.820+08:00 level=INFO msg="workdir -> "
time=2025-02-09T20:07:31.820+08:00 level=INFO msg="please visit http://localhost:8080/"
```

Then open the web page http://localhost:8080/ and you can start using it.

## Building

The framework used in the building process: https://taskfile.dev/

The following command is used to build all binary packages:

```shell
task build
```