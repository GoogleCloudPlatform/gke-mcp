# Troubleshooting

## gke-mcp: command not found on macOS or Linux

If you run `gke-mcp` after using the manual install method and get an error like `-bash: gke-mcp: command not found`, it usually means the directory where Go places compiled programs is not included in your shell's `PATH` environment variable.

Here are the steps to fix this:

### Identify the Go Binary Location

The `go install` command places binaries in the directory specified by your `GOBIN` environment variable. If `GOBIN` is not set, it defaults to the `bin` subdirectory inside your `GOPATH` ([source](https://go.dev/doc/install)).

To find your `GOPATH`, run:

```sh
go env GOPATH
```

The default installation directory will be the path from that command, with `/bin` appended (e.g., `/Users/your-user/go/bin`).

### Update Your Shell Configuration File

You need to add the Go binary directory to your `PATH`. The configuration file you edit depends on the shell you use.

- **For Bash** (the default on many Linux distributions and older versions of macOS), add the following line to your `~/.bash_profile` or `~/.bashrc` file:

  ```sh
  export PATH=$PATH:$(go env GOPATH)/bin
  ```

- **For Zsh** (the default shell on newer versions of macOS), add the same line to your `~/.zshrc` file:

  ```sh
  export PATH=$PATH:$(go env GOPATH)/bin
  ```

### Apply the Changes

For the changes to take effect in your current terminal session, you must reload the configuration file using the `source` command.

- If you edited `~/.bash_profile`:

  ```sh
  source ~/.bash_profile
  ```

- If you edited `~/.zshrc`:

  ```sh
  source ~/.zshrc
  ```

After completing these steps, you should be able to run the `gke-mcp` command successfully.
