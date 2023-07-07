package transformer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitCompose_computeRepoAndBuildPath(t *testing.T) {
	testCases := []struct {
		name        string
		projectPath string
		composePath string
		buildPath   string
		defaultRepo string
		want        []string
	}{
		{
			name:        "BuildPathInsideProjectPath",
			projectPath: "/parent/path/myproject/",
			composePath: "/parent/path/myproject/docker-compose.yml",
			buildPath:   ".",
			defaultRepo: "defaultRepo",
			want:        []string{"defaultRepo", "."},
		},
		{
			name:        "BuildPathInsideProjectPathDeepCompose",
			projectPath: "/parent/path/myproject/",
			composePath: "/parent/path/myproject/.ergomake/docker-compose.yml",
			buildPath:   "..",
			defaultRepo: "defaultRepo",
			want:        []string{"defaultRepo", ".."},
		},
		{
			name:        "BuildPathOutsideProjectPath",
			projectPath: "/parent/path/myproject/",
			composePath: "/parent/path/myproject/docker-compose.yml",
			buildPath:   "../otherproject",
			defaultRepo: "defaultRepo",
			want:        []string{"otherproject", "."},
		},
		{
			name:        "BuildPathOutsideProjectPathDeepCompose",
			projectPath: "/parent/path/myproject/",
			composePath: "/parent/path/myproject/.ergomake/docker-compose.yml",
			buildPath:   "../../otherproject",
			defaultRepo: "defaultRepo",
			want:        []string{"otherproject", "."},
		},
		{
			name:        "BuildPathOutsideProjectDeepPath",
			projectPath: "/parent/path/myproject/",
			composePath: "/parent/path/myproject/docker-compose.yml",
			buildPath:   "../otherproject/build",
			defaultRepo: "defaultRepo",
			want:        []string{"otherproject", "build"},
		},
		{
			name:        "BuildPathOutsideProjectDeepPathDeepCompose",
			projectPath: "/parent/path/myproject/",
			composePath: "/parent/path/myproject/.ergomake/docker-compose.yml",
			buildPath:   "../../otherproject/build",
			defaultRepo: "defaultRepo",
			want:        []string{"otherproject", "build"},
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := &gitCompose{
				projectPath:    tc.projectPath,
				configFilePath: tc.composePath,
			}

			repo, buildPath := c.computeRepoAndBuildPath(tc.buildPath, tc.defaultRepo)
			assert.Equal(t, tc.want[0], repo)
			assert.Equal(t, tc.want[1], buildPath)
		})
	}
}
