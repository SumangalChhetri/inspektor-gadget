// Copyright 2024 The Inspektor Gadget authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package image

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	commonImage "github.com/inspektor-gadget/inspektor-gadget/cmd/common/image"
	. "github.com/inspektor-gadget/inspektor-gadget/integration"
)

func TestImage(t *testing.T) {
	// ensure that experimental is enabled
	os.Setenv("IG_EXPERIMENTAL", "true")

	// start registry
	r := StartRegistry(t, "test-image-registry")
	t.Cleanup(func() {
		r.Stop(t)
	})
	if len(r.PortBindings()) == 0 || r.PortBindings()["5000/tcp"] == nil {
		t.Fatal("registry port not exposed")
	}
	pb := r.PortBindings()["5000/tcp"][0]
	registryAddr := fmt.Sprintf("%s:%s", pb.HostIP, pb.HostPort)

	testReg := "reg.com"
	testRepo := "repo1"
	testTag := "tag1"
	testLocalImage := fmt.Sprintf("%s/%s:%s", testReg, testRepo, testTag)
	testRegistryImage := fmt.Sprintf("%s/%s:%s", registryAddr, testRepo, testTag)

	tmpFolder := t.TempDir()
	exportPath := path.Join(tmpFolder, "export.tar")

	// ensure all images are removed
	t.Cleanup(func() {
		// remove local image
		cmd := commonImage.NewRemoveCmd()
		cmd.SetArgs([]string{testLocalImage})
		cmd.Execute()

		// remove registry image
		cmd = commonImage.NewRemoveCmd()
		cmd.SetArgs([]string{testRegistryImage})
		cmd.Execute()
	})

	type testCase struct {
		name           string
		cmd            *cobra.Command
		args           []string
		expectedStdout []string
		expectedStderr []string
		negateExpected bool
	}

	// The order of these tests is important as one test depends on the previous
	// one
	testCases := []testCase{
		{
			name: "build",
			cmd:  commonImage.NewBuildCmd(),
			args: []string{
				"--builder-image", *testBuilderImage, "--tag", testLocalImage, "../../../gadgets/trace_open",
			},
			expectedStdout: []string{
				fmt.Sprintf("Successfully built %s", testLocalImage),
			},
		},
		{
			name: "list",
			cmd:  commonImage.NewListCmd(),
			args: []string{},
			expectedStdout: []string{
				testRepo,
				testTag,
			},
		},
		{
			name: "tag",
			cmd:  commonImage.NewTagCmd(),
			args: []string{testLocalImage, testRegistryImage},
			expectedStdout: []string{
				fmt.Sprintf("Successfully tagged with %s", testRegistryImage),
			},
		},
		{
			name: "push",
			cmd:  commonImage.NewPushCmd(),
			args: []string{testRegistryImage, "--insecure"},
			expectedStdout: []string{
				fmt.Sprintf("Successfully pushed %s", testRegistryImage),
			},
		},
		{
			name: "push-invalid-image",
			cmd:  commonImage.NewPushCmd(),
			args: []string{"unknown", "--insecure"},
			expectedStderr: []string{
				"failed to resolve ghcr.io/inspektor-gadget/gadget/unknown:latest: not found",
			},
		},
		{
			name: "push-unknown-tag",
			cmd:  commonImage.NewPushCmd(),
			args: []string{fmt.Sprintf("%s/%s:%s", registryAddr, testRepo, "unknown"), "--insecure"},
			expectedStderr: []string{
				fmt.Sprintf("%s/%s:%s: not found", registryAddr, testRepo, "unknown"),
			},
		},
		{
			name: "remove-local-image",
			cmd:  commonImage.NewRemoveCmd(),
			args: []string{testLocalImage},
			expectedStdout: []string{
				fmt.Sprintf("Successfully removed %s", testLocalImage),
			},
		},
		{
			name: "remove-registry-image",
			cmd:  commonImage.NewRemoveCmd(),
			args: []string{testRegistryImage},
			expectedStdout: []string{
				fmt.Sprintf("Successfully removed %s", testRegistryImage),
			},
		},
		{
			name:           "validate-remove",
			cmd:            commonImage.NewListCmd(),
			args:           []string{},
			negateExpected: true,
			// can't use combined repository here as REPOSITORY column can be truncated
			expectedStdout: []string{
				registryAddr,
				testRepo,
			},
		},
		{
			name: "pull",
			cmd:  commonImage.NewPullCmd(),
			args: []string{testRegistryImage, "--insecure"},
			expectedStdout: []string{
				fmt.Sprintf("Successfully pulled %s", testRegistryImage),
			},
		},
		{
			name: "validate-pull",
			cmd:  commonImage.NewListCmd(),
			args: []string{},
			expectedStdout: []string{
				registryAddr,
				testTag,
			},
		},
		{
			name: "pull-invalid-image",
			cmd:  commonImage.NewPullCmd(),
			args: []string{"unknown", "--insecure"},
			expectedStderr: []string{
				"failed to resolve ghcr.io/inspektor-gadget/gadget/unknown:latest",
			},
		},
		{
			name: "pull-unknown-tag",
			cmd:  commonImage.NewPullCmd(),
			args: []string{fmt.Sprintf("%s/%s:%s", registryAddr, testRepo, "unknown"), "--insecure"},
			expectedStderr: []string{
				fmt.Sprintf("%s/%s:%s: not found", registryAddr, testRepo, "unknown"),
			},
		},
		{
			name: "export",
			cmd:  commonImage.NewExportCmd(),
			args: []string{testRegistryImage, exportPath},
			expectedStdout: []string{
				fmt.Sprintf("Successfully exported images to %s", exportPath),
			},
		},
		{
			name: "import",
			cmd:  commonImage.NewImportCmd(),
			args: []string{exportPath},
			expectedStdout: []string{
				fmt.Sprintf("Successfully imported images:\n  %s", testRegistryImage),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			tc.cmd.SetArgs(tc.args)
			tc.cmd.SetOut(&stdout)
			tc.cmd.SetErr(&stderr)
			err := tc.cmd.Execute()
			if len(tc.expectedStderr) > 0 {
				require.NotNil(t, err)
				for _, expected := range tc.expectedStderr {
					if tc.negateExpected {
						require.NotContains(t, stderr.String(), expected)
						continue
					}
					require.Contains(t, stderr.String(), expected)
				}
				return
			}
			require.Nil(t, err)
			require.Empty(t, stderr.String())
			for _, expected := range tc.expectedStdout {
				if tc.negateExpected {
					require.NotContains(t, stdout.String(), expected)
					continue
				}
				require.Contains(t, stdout.String(), expected)
			}
		})
	}
}
