package compose

import (
	"context"
	_ "embed"
	"os"
	"strings"
	"testing"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/google/go-cmp/cmp"
)

func TestOutputString(t *testing.T) {
	projectName := "testdata"
	project, err := loader.LoadWithContext(
		context.Background(), types.ConfigDetails{
			WorkingDir: "./testdata",
			ConfigFiles: []types.ConfigFile{
				{Filename: "./testdata/compose.yml"},
				{Filename: "./testdata/additional.yml"},
			},
			Environment: types.NewMapping(os.Environ()),
		},
		func(o *loader.Options) {
			o.SetProjectName(projectName, true)
		},
	)
	if err != nil {
		panic(err)
	}

	type testCase struct {
		lines     string
		expected  []ComposeOutputLine
		shouldErr bool
	}

	for _, tc := range []testCase{
		{
			lines: createDryrunTxt,
			expected: []ComposeOutputLine{
				{DryRunMode: true, ResourceType: Network, Name: "sample network", StateType: Creating},
				{DryRunMode: true, ResourceType: Network, Name: "sample network", StateType: Created},
				{DryRunMode: true, ResourceType: Volume, Name: "sample-volume", StateType: Creating},
				{DryRunMode: true, ResourceType: Volume, Name: "sample-volume", StateType: Created},
				{DryRunMode: true, ResourceType: Container, Name: "sample_service", Num: 1, StateType: Creating},
				{DryRunMode: true, ResourceType: Container, Name: "additional", Num: 1, StateType: Creating},
				{DryRunMode: true, ResourceType: Container, Name: "sample_service", Num: 1, StateType: Created},
				{DryRunMode: true, ResourceType: Container, Name: "additional", Num: 1, StateType: Created},
			},
		},
		{
			lines: create,
			expected: []ComposeOutputLine{
				{ResourceType: Network, Name: "sample network", StateType: Creating},
				{ResourceType: Network, Name: "sample network", StateType: Created},
				{ResourceType: Volume, Name: "sample-volume", StateType: Creating},
				{ResourceType: Volume, Name: "sample-volume", StateType: Created},
				{ResourceType: Container, Name: "sample_service", Num: 1, StateType: Creating},
				{ResourceType: Container, Name: "additional", Num: 1, StateType: Creating},
				{ResourceType: Container, Name: "additional", Num: 1, StateType: Created},
				{ResourceType: Container, Name: "sample_service", Num: 1, StateType: Created},
			},
		},
		{
			lines: start,
			expected: []ComposeOutputLine{
				{ResourceType: Container, Name: "sample_service", Num: 1, StateType: Starting},
				{ResourceType: Container, Name: "additional", Num: 1, StateType: Starting},
				{ResourceType: Container, Name: "sample_service", Num: 1, StateType: Started},
				{ResourceType: Container, Name: "additional", Num: 1, StateType: Started},
			},
		},
		{
			lines: recreateDryrun,
			expected: []ComposeOutputLine{
				{DryRunMode: true, ResourceType: Container, Name: "additional", Num: 1, StateType: Stopping},
				{DryRunMode: true, ResourceType: Container, Name: "additional", Num: 1, StateType: Stopped},
				{DryRunMode: true, ResourceType: Container, Name: "additional", Num: 1, StateType: Removing},
				{DryRunMode: true, ResourceType: Container, Name: "additional", Num: 1, StateType: Removed},
				{DryRunMode: true, ResourceType: Container, Name: "sample_service", Num: 1, StateType: Recreate},
				{DryRunMode: true, ResourceType: Container, Name: "sample_service", Num: 1, StateType: Recreated},
			},
		},
		{
			lines: recreate,
			expected: []ComposeOutputLine{
				{ResourceType: Container, Name: "additional", Num: 1, StateType: Stopping},
				{ResourceType: Container, Name: "additional", Num: 1, StateType: Stopped},
				{ResourceType: Container, Name: "additional", Num: 1, StateType: Removing},
				{ResourceType: Container, Name: "additional", Num: 1, StateType: Removed},
				{ResourceType: Container, Name: "sample_service", Num: 1, StateType: Recreate},
				{ResourceType: Container, Name: "sample_service", Num: 1, StateType: Recreated},
			},
		},
		{
			lines:     restartDryrun,
			shouldErr: true,
		},
		{
			lines: restart,
			expected: []ComposeOutputLine{
				{ResourceType: Container, Name: "sample_service", Num: 1, StateType: Starting},
				{ResourceType: Container, Name: "sample_service", Num: 1, StateType: Started},
			},
		},
		{
			lines: down,
			expected: []ComposeOutputLine{
				{ResourceType: Container, Name: "sample_service", Num: 1, StateType: Stopping},
				{ResourceType: Container, Name: "sample_service", Num: 1, StateType: Stopped},
				{ResourceType: Container, Name: "sample_service", Num: 1, StateType: Removing},
				{ResourceType: Container, Name: "sample_service", Num: 1, StateType: Removed},
				{ResourceType: Volume, Name: "sample-volume", StateType: Removing},
				{ResourceType: Volume, Name: "sample-volume", StateType: Removed},
				{ResourceType: Network, Name: "sample network", StateType: Removing},
				{ResourceType: Network, Name: "sample network", StateType: Removed},
			},
		},
		{
			lines:     nonexistentComposeYml,
			shouldErr: true,
		},
	} {
		for idx, line := range strings.Split(tc.lines, "\n") {
			if line == "" {
				break
			}
			decoded, err := DecodeComposeOutputLine(line, "testdata", project)
			if tc.shouldErr {
				if err == nil {
					t.Errorf("decoding should cause an error but is nil")
				}
				continue
			}
			if err != nil {
				t.Errorf("decode err = %s", err)
				continue
			}
			if diff := cmp.Diff(decoded, tc.expected[idx]); diff != "" {
				t.Errorf("not equal. diff =%s", diff)
			}
		}

	}
}

//go:embed  testdata/00_create-dryrun.txt
var createDryrunTxt string

//go:embed  testdata/01_create.txt
var create string

//go:embed  testdata/02_start.txt
var start string

//go:embed  testdata/03_recreate-dryrun.txt
var recreateDryrun string

//go:embed  testdata/04_recreate.txt
var recreate string

//go:embed  testdata/05_restart-dryrun.txt
var restartDryrun string

//go:embed  testdata/06_restart.txt
var restart string

//go:embed  testdata/07_down.txt
var down string

//go:embed  testdata/08_nonexistent_compose_yml.txt
var nonexistentComposeYml string
