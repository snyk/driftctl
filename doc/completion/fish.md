# Fish Completion

Copy the `fish` completion file to your usual completion directory:

  ```shell
  cp completions/driftctl.fish ~/.config/fish/completions/
  ```

Now `fish` is autocompleting the `driftctl` commands:

```shell
$ driftctl scan 
--filter        (Filter Resources)  --help       (Print help)
--from  (Terraform State Location)  --output  (Output Format)
```