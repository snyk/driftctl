# ~/.config/fish/completions

# the only driftctl subcommand is "scan", but more will come
set -l commands scan version

# Disable file completions for the entire command
complete -c driftctl -f

# This line offers the subcommands
# - "scan",
# - "version"
# if no subcommand has been given so far.

# complete -f -c driftctl -n "not __fish_seen_subcommand_from $commands" -a $commands

complete -f -c driftctl -n "not __fish_seen_subcommand_from $commands" -a version -d "display version"
complete -f -c driftctl -n "not __fish_seen_subcommand_from $commands" -a scan -d "scan for drifts"

# complete -c driftctl -n "not __fish_seen_subcommand_from $commands" -a "--help --error-reporting"

# offers driftctl scan subcommands
complete -c driftctl -n "__fish_seen_subcommand_from scan" -a "--from" -d "Terraform State Location"
complete -c driftctl -n "__fish_seen_subcommand_from scan" -a "--output" -d "Output Format"
complete -c driftctl -n "__fish_seen_subcommand_from scan" -a "--filter" -d "Filter Resources"


# Options for source Terraform State
complete -c driftctl -n "__fish_seen_subcommand_from --from" -a "tfstate://terraform.tfstate" -d "A Local Terraform State"
complete -c driftctl -n "__fish_seen_subcommand_from --from" -a "tfstate+s3://my-bucket/terraform.tfstate" -d "An S3 Remote Terraform State"

# Options for output
complete -c driftctl -n "__fish_seen_subcommand_from --output" -a "console://" -d "Console Output"
complete -c driftctl -n "__fish_seen_subcommand_from --output" -a "json:///dev/stdout" -d "JSON Output"

# options that can be used anywhere
complete -c driftctl -s h -l help -d 'Print help'
