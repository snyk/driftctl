package acceptance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	cmderrors "github.com/cloudskiff/driftctl/pkg/cmd/errors"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/test"

	"github.com/spf13/cobra"

	"github.com/cloudskiff/driftctl/logger"
	"github.com/cloudskiff/driftctl/pkg/cmd"
)

type AccCheck struct {
	PreExec  func()
	PostExec func()
	Env      map[string]string
	Check    func(result *ScanResult, stdout string, err error)
}

type AccTestCase struct {
	Path              string
	Args              []string
	OnStart           func()
	OnEnd             func()
	Checks            []AccCheck
	tmpResultFilePath string
	originalEnv       []string
}

func (c *AccTestCase) createResultFile(t *testing.T) error {
	tmpDir := t.TempDir()
	file, err := ioutil.TempFile(tmpDir, "result")
	if err != nil {
		return err
	}
	defer file.Close()
	c.tmpResultFilePath = file.Name()
	return nil
}

func (c *AccTestCase) validate() error {
	if c.Checks == nil || len(c.Checks) == 0 {
		return fmt.Errorf("checks attribute must be defined")
	}

	if c.Path == "" {
		return fmt.Errorf("path attribute must be defined")
	}

	for _, arg := range c.Args {
		if arg == "--output" || arg == "-o" {
			return fmt.Errorf("--output flag should not be defined in test case, it is automatically tested")
		}
	}

	return nil
}

func (c *AccTestCase) getResultFilePath() string {
	return c.tmpResultFilePath
}

func (c *AccTestCase) getResult(t *testing.T) *ScanResult {
	analysis := analyser.Analysis{}
	result, err := ioutil.ReadFile(c.getResultFilePath())
	if err != nil {
		return nil
	}

	if err := json.Unmarshal(result, &analysis); err != nil {
		return nil
	}

	return NewScanResult(t, analysis)
}

/**
 * Retrieve env from os.Environ() but override every variable prefixed with ACC_
 * e.g. ACC_AWS_PROFILE will override AWS_PROFILE
 */
func (c *AccTestCase) resolveTerraformEnv() []string {

	environMap := make(map[string]string, len(os.Environ()))

	const PREFIX string = "ACC_"

	for _, e := range os.Environ() {
		envKeyValue := strings.SplitN(e, "=", 2)
		if strings.HasPrefix(envKeyValue[0], PREFIX) {
			varName := strings.TrimPrefix(envKeyValue[0], PREFIX)
			environMap[varName] = envKeyValue[1]
			continue
		}
		if _, exist := environMap[envKeyValue[0]]; !exist {
			environMap[envKeyValue[0]] = envKeyValue[1]
		}
	}

	results := make([]string, 0, len(environMap))
	for k, v := range environMap {
		results = append(results, fmt.Sprintf("%s=%s", k, v))
	}

	return results
}

func (c *AccTestCase) terraformInit() error {
	_, err := os.Stat(path.Join(c.Path, ".terraform"))
	if os.IsNotExist(err) {
		logrus.Debug("Running terraform init ...")
		cmd := exec.Command("terraform", "init", "-input=false")
		cmd.Dir = c.Path
		cmd.Env = c.resolveTerraformEnv()
		out, err := cmd.CombinedOutput()
		if err != nil {
			return errors.Wrap(err, string(out))
		}
		logrus.Debug("Terraform init done")
	}

	return nil
}

func (c *AccTestCase) terraformApply() error {
	logrus.Debug("Running terraform apply ...")
	cmd := exec.Command("terraform", "apply", "-auto-approve")
	cmd.Dir = c.Path
	cmd.Env = c.resolveTerraformEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, string(out))
	}
	logrus.Debug("Terraform apply done")

	return nil
}

func (c *AccTestCase) terraformDestroy() error {
	logrus.Debug("Running terraform destroy ...")
	cmd := exec.Command("terraform", "destroy", "-auto-approve")
	cmd.Dir = c.Path
	cmd.Env = c.resolveTerraformEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, string(out))
	}
	logrus.Debug("Terraform destroy done")

	return nil
}

func runDriftCtlCmd(driftctlCmd *cmd.DriftctlCmd) (*cobra.Command, string, error) {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	cmd, cmdErr := driftctlCmd.ExecuteC()
	// Ignore not in sync errors in acceptance test context
	if _, isNotInSyncErr := cmdErr.(cmderrors.InfrastructureNotInSync); isNotInSyncErr {
		cmdErr = nil
	}
	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	out := <-outC
	return cmd, out, cmdErr
}

func (c *AccTestCase) useTerraformEnv() {
	c.originalEnv = os.Environ()
	c.setEnv(c.resolveTerraformEnv())
}

func (c *AccTestCase) restoreEnv() {
	if c.originalEnv != nil {
		logrus.Debug("Restoring original environment ...")
		os.Clearenv()
		c.setEnv(c.originalEnv)
		c.originalEnv = nil
	}
}

func (c *AccTestCase) setEnv(env []string) {
	os.Clearenv()
	for _, e := range env {
		envKeyValue := strings.SplitN(e, "=", 2)
		os.Setenv(envKeyValue[0], envKeyValue[1])
	}
}

func Run(t *testing.T, c AccTestCase) {

	if os.Getenv("DRIFTCTL_ACC") == "" {
		t.Skip()
	}

	if err := c.validate(); err != nil {
		t.Fatal(err)
	}

	if c.OnStart != nil {
		c.OnStart()
	}

	// Disable terraform version checks
	// @link https://www.terraform.io/docs/commands/index.html#upgrade-and-security-bulletin-checks
	checkpoint := os.Getenv("CHECKPOINT_DISABLE")
	os.Setenv("CHECKPOINT_DISABLE", "true")

	// Execute terraform init if .terraform folder is not found in test folder
	err := c.terraformInit()
	if err != nil {
		t.Fatal(err)
	}

	err = c.terraformApply()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		c.restoreEnv()
		err := c.terraformDestroy()
		os.Setenv("CHECKPOINT_DISABLE", checkpoint)
		if err != nil {
			t.Fatal(err)
		}
	}()

	logger.Init(logger.GetConfig())

	driftctlCmd := cmd.NewDriftctlCmd(test.Build{})

	err = c.createResultFile(t)
	if err != nil {
		t.Fatal(err)
	}
	if c.Args != nil {
		c.Args = append([]string{""}, c.Args...)
		c.Args = append(c.Args,
			"--from", fmt.Sprintf("tfstate://%s", path.Join(c.Path, "terraform.tfstate")),
			"--output", fmt.Sprintf("json://%s", c.getResultFilePath()),
		)
	}
	os.Args = c.Args

	for _, check := range c.Checks {
		if check.Check == nil {
			t.Fatal("Check attribute must be defined")
		}
		if len(check.Env) > 0 {
			for key, value := range check.Env {
				os.Setenv(key, value)
			}
		}
		if check.PreExec != nil {
			c.useTerraformEnv()
			check.PreExec()
			c.restoreEnv()
		}
		_, out, cmdErr := runDriftCtlCmd(driftctlCmd)
		if len(check.Env) > 0 {
			for key := range check.Env {
				_ = os.Unsetenv(key)
			}
		}
		check.Check(c.getResult(t), out, cmdErr)
		if check.PostExec != nil {
			check.PostExec()
		}
	}
	if c.OnEnd != nil {
		c.OnEnd()
	}
}
