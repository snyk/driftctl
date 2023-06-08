package acceptance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/analyser"
	cmderrors "github.com/snyk/driftctl/pkg/cmd/errors"

	"github.com/sirupsen/logrus"

	"github.com/snyk/driftctl/test"

	"github.com/spf13/cobra"

	"github.com/snyk/driftctl/logger"
	"github.com/snyk/driftctl/pkg/cmd"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/hashicorp/terraform-exec/tfinstall"
)

type ShouldRetryFunc func(result *test.ScanResult, retryDuration time.Duration, retryCount uint8) bool

type AccCheck struct {
	PreExec     func()
	PostExec    func()
	Env         map[string]string
	Args        func() []string
	ShouldRetry ShouldRetryFunc
	Check       func(result *test.ScanResult, stdout string, err error)
}

type AccTestCase struct {
	DoNotRunTerraform          bool
	TerraformVersion           string
	WorkingDir                 string
	Paths                      []string
	Args                       []string
	OnStart                    func()
	OnEnd                      func()
	Checks                     []AccCheck
	tmpResultFilePath          string
	originalEnv                []string
	tf                         map[string]*tfexec.Terraform
	ShouldRefreshBeforeDestroy bool
}

func (c *AccTestCase) initTerraformExecutor() error {
	logrus.Debug("Initializing terraform...")
	installPath := path.Join(os.TempDir(), "terraform-bin", c.TerraformVersion)
	binPath := path.Join(installPath, "terraform")
	execPathFinderOptions := make([]tfinstall.ExecPathFinder, 0)

	err := os.MkdirAll(installPath, fs.ModePerm)
	if err != nil {
		return err
	}

	_, err = os.Stat(binPath)
	if os.IsNotExist(err) {
		execPathFinderOptions = append(execPathFinderOptions, tfinstall.ExactVersion(c.TerraformVersion, installPath))
	} else {
		execPathFinderOptions = append(execPathFinderOptions, tfinstall.ExactPath(binPath))
	}

	execPath, err := tfinstall.Find(context.Background(), execPathFinderOptions...)
	if err != nil {
		return err
	}

	c.tf = make(map[string]*tfexec.Terraform, 1)
	for _, p := range c.Paths {
		c.tf[p], err = tfexec.NewTerraform(p, execPath)
		if err != nil {
			return err
		}
		env := c.resolveTerraformEnv()
		if err := c.tf[p].SetEnv(env); err != nil {
			return err
		}
	}
	return nil
}

func (c *AccTestCase) createResultFile(t *testing.T) error {
	tmpDir := t.TempDir()
	file, err := os.CreateTemp(tmpDir, "result")
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

	if len(c.Paths) < 1 && !c.DoNotRunTerraform {
		return fmt.Errorf("Paths attribute must be defined")
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

func (c *AccTestCase) getResult(t *testing.T) *test.ScanResult {
	analysis := &analyser.Analysis{}
	result, err := os.ReadFile(c.getResultFilePath())
	if err != nil {
		return nil
	}

	if err := json.Unmarshal(result, analysis); err != nil {
		return nil
	}

	return test.NewScanResult(t, analysis)
}

/**
 * Retrieve env from os.Environ() but override every variable prefixed with ACC_
 * e.g. ACC_AWS_PROFILE will override AWS_PROFILE
 */
func (c *AccTestCase) resolveTerraformEnv() map[string]string {

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

	return environMap
}

func (c *AccTestCase) terraformInit() error {
	if err := c.initTerraformExecutor(); err != nil {
		return err
	}
	for _, p := range c.Paths {
		_, err := os.Stat(path.Join(p, ".terraform"))
		if os.IsNotExist(err) {
			logrus.WithFields(logrus.Fields{
				"path": p,
			}).Debug("Running terraform init ...")
			stderr := new(bytes.Buffer)
			c.tf[p].SetStderr(stderr)
			if err := c.tf[p].Init(context.Background()); err != nil {
				return errors.Wrap(err, stderr.String())
			}
			logrus.WithFields(logrus.Fields{
				"path": p,
			}).Debug("Terraform init done")
		}
	}

	return nil
}

func (c *AccTestCase) terraformApply() error {
	for _, p := range c.Paths {
		logrus.WithFields(logrus.Fields{
			"p": p,
		}).Debug("Running terraform apply ...")
		stderr := new(bytes.Buffer)
		c.tf[p].SetStderr(stderr)
		if err := c.tf[p].Apply(context.Background()); err != nil {
			return errors.Wrap(err, stderr.String())
		}
		logrus.WithFields(logrus.Fields{
			"p": p,
		}).Debug("Terraform apply done")
	}

	return nil
}

func (c *AccTestCase) terraformDestroy() error {
	if c.ShouldRefreshBeforeDestroy {
		logrus.Debug("Running terraform refresh...")
		if err := c.terraformRefresh(); err != nil {
			return err
		}
	}

	for _, p := range c.Paths {
		logrus.WithFields(logrus.Fields{
			"p": p,
		}).Debug("Running terraform destroy ...")
		stderr := new(bytes.Buffer)
		c.tf[p].SetStderr(stderr)
		if err := c.tf[p].Destroy(context.Background()); err != nil {
			return errors.Wrap(err, stderr.String())
		}
		logrus.WithFields(logrus.Fields{
			"p": p,
		}).Debug("Terraform destroy done")
	}

	return nil
}

func (c *AccTestCase) terraformRefresh() error {
	for _, p := range c.Paths {
		logrus.WithFields(logrus.Fields{
			"p": p,
		}).Debug("Running terraform refresh ...")
		stderr := new(bytes.Buffer)
		c.tf[p].SetStderr(stderr)
		if err := c.tf[p].Refresh(context.Background()); err != nil {
			return errors.Wrap(err, stderr.String())
		}
		logrus.WithFields(logrus.Fields{
			"p": p,
		}).Debug("Terraform refresh done")
	}

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
	environMap := c.resolveTerraformEnv()
	env := make([]string, 0, len(environMap))
	for k, v := range environMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	c.setEnv(env)
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

	logger.Init()

	if os.Getenv("DRIFTCTL_ACC") == "" {
		t.Skip()
	}

	if err := c.validate(); err != nil {
		t.Fatal(err)
	}

	if c.OnStart != nil {
		c.useTerraformEnv()
		c.OnStart()
		if c.OnEnd != nil {
			defer func() {
				c.useTerraformEnv()
				c.OnEnd()
				c.restoreEnv()
			}()
		}
		c.restoreEnv()
	}

	// Disable terraform version checks
	// @link https://www.terraform.io/docs/commands/index.html#upgrade-and-security-bulletin-checks
	checkpoint := os.Getenv("CHECKPOINT_DISABLE")
	os.Setenv("CHECKPOINT_DISABLE", "true")

	// Retry after 2s, 4s, 8s, 16s, 32s, 64s, 2m, 2m, 2m, 2m
	// Try tweaking the backoff interval limit and/or the retry count limit in
	// response to flaky tests.
	limitedExponentialBackoff := retrier.New(retrier.LimitedExponentialBackoff(10, time.Second*2, time.Minute*2), nil)

	if !c.DoNotRunTerraform {
		// Execute terraform init if .terraform folder is not found in test folder
		err := limitedExponentialBackoff.Run(c.terraformInit)
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			c.restoreEnv()
			err := limitedExponentialBackoff.Run(c.terraformDestroy)
			os.Setenv("CHECKPOINT_DISABLE", checkpoint)
			if err != nil {
				t.Fatal(err)
			}
		}()

		err = limitedExponentialBackoff.Run(c.terraformApply)
		if err != nil {
			t.Fatal(err)
		}

		err = c.createResultFile(t)
		if err != nil {
			t.Fatal(err)
		}
	}

	// If the path contains only one element, we switch to this directory for driftctl execution
	// We can override this logic by passing a WorkingDir argument in test
	if c.WorkingDir == "" && len(c.Paths) == 1 {
		c.WorkingDir = c.Paths[0]
	}

	if c.Args != nil {
		c.Args = append([]string{""}, c.Args...)
		isFromSet := false
		for _, arg := range c.Args {
			if arg == "--from" || arg == "-f" {
				isFromSet = true
				break
			}
		}
		// If any --from flag was manually provided OR if a working dir is specified,
		// do not setup any --from flags
		if !isFromSet && c.WorkingDir == "" {
			for _, p := range c.Paths {
				c.Args = append(c.Args,
					"--from", fmt.Sprintf("tfstate://%s", path.Join(p, "terraform.tfstate")),
				)
			}
		}
		if c.getResultFilePath() != "" {
			c.Args = append(c.Args,
				"--output", fmt.Sprintf("json://%s", c.getResultFilePath()),
			)
		}
	}

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
		os.Args = c.Args
		if check.Args != nil {
			os.Args = append(os.Args, check.Args()...)
		}

		wd, _ := os.Getwd()
		if c.WorkingDir != "" {
			logrus.WithField("dir", c.WorkingDir).Debug("Switching working directory for driftctl execution")
			err := os.Chdir(c.WorkingDir)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"from": wd,
					"to":   c.WorkingDir,
				}).Errorf("Unable to switch to testing working dir: %s", err.Error())
			}
		}
		logrus.WithField("args", fmt.Sprintf("%+v", os.Args)).Debug("Running driftctl")
		driftctlCmd := cmd.NewDriftctlCmd(test.Build{})
		_, out, cmdErr := runDriftCtlCmd(driftctlCmd)
		result := c.getResult(t)
		var retryCount uint8 = 0
		timeBeforeRetry := time.Now()
		for {
			if check.ShouldRetry == nil || !check.ShouldRetry(result, time.Since(timeBeforeRetry), retryCount) {
				break
			}
			logrus.
				WithField("count", fmt.Sprintf("%d", retryCount)).
				WithField("retry_duration", time.Since(timeBeforeRetry).Round(time.Second)).
				Debug("Retrying scan ...")
			_, _, _ = runDriftCtlCmd(driftctlCmd)
			result = c.getResult(t)
			retryCount++
		}
		// Restore original working directory
		if c.WorkingDir != "" {
			err := os.Chdir(wd)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"to":   wd,
					"from": c.WorkingDir,
				}).Errorf("Unable to switch back to original working dir: %s", err.Error())
			}
		}

		if len(check.Env) > 0 {
			for key := range check.Env {
				_ = os.Unsetenv(key)
			}
		}
		check.Check(result, out, cmdErr)
		if check.PostExec != nil {
			check.PostExec()
		}
	}
}

// LinearBackoff returns a function that retries using
// a back-off strategy of retrying 'n' times and doubling the
// amount of time waited after each one.
func LinearBackoff(limit time.Duration) ShouldRetryFunc {
	return func(result *test.ScanResult, retryDuration time.Duration, retryCount uint8) bool {
		if result.IsSync() || retryDuration > limit {
			return false
		}
		time.Sleep((2 * time.Duration(retryCount)) * time.Minute)
		return true
	}
}
