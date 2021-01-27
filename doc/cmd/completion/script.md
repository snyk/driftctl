# Driftctl completion script

Driftctl can output completion script (also known as *tab completion*) for you to use on your shell. Currently `bash`, `zsh`, `fish` and `powershell` shells are supported.

### Before you start
In order to generate the completion script required to make the completion work, you have to install driftctl CLI first.

### Generate the completion file
To generate the completion script you can use:

```shell
$ driftctl completion [bash|zsh|fish|powershell]
```

By default, this command will print on the standard output the content of the completion script. To make the completion work you will need to redirect it to the completion folder of your shell.

### Bash
```shell
# Linux:
$ driftctl completion bash | sudo tee /etc/bash_completion.d/driftctl

# MacOS:
$ driftctl completion bash > /usr/local/etc/bash_completion.d/driftctl
```

Remember to open a new shell to test the functionality.

### Zsh
If shell completion is not already enabled in your environment, you will need to enable it. You can execute the following once:

```shell
$ echo "autoload -U compinit; compinit" >> ~/.zshrc
```

At this point you can generate and place the completion script in your completion folder listed in your `fpath` if it already exists. Otherwise, you can create a directory, add it to your `fpath` and copy the file in it:

```shell
$ driftctl completion zsh > fpath/completion_folder/_driftctl
```

#### Oh-My-Zsh
```shell
$ mkdir -p ~/.oh-my-zsh/completions
$ driftctl completion zsh > ~/.oh-my-zsh/completions/_driftctl
```

You will need to start a new shell for this setup to take effect.

### Fish
```shell
$ driftctl completion fish > ~/.config/fish/completions/driftctl.fish
```

Remember to create the directory if it's not already there `mkdir -p ~/.config/fish/completions/`.

Remember to open a new shell to test the functionality.

### Powershell
```shell
$ driftctl completion powershell > driftctl.ps1
```

You will need to source this file from your powershell profile for this to work as expected.
