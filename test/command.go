package test

import (
	"bytes"

	"github.com/spf13/cobra"
)

func Execute(cmd *cobra.Command, args ...string) (output string, err error) {
	_, output, err = ExecuteC(cmd, args...)
	return output, err
}

func ExecuteC(cmd *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	c, err = cmd.ExecuteC()

	return c, buf.String(), err
}
