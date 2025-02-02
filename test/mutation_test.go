package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/mutation"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

var configYaml = `
checks:
  pullPolicyNotAlways: warning
  hostIPCSet: danger
  hostPIDSet: danger
  hostNetworkSet: danger
  deploymentMissingReplicas: warning
  priorityClassNotSet: ignore
  runAsRootAllowed: danger
  cpuRequestsMissing: warning
  cpuLimitsMissing: warning
  memoryRequestsMissing: warning
  memoryLimitsMissing: warning
  readinessProbeMissing: warning
  livenessProbeMissing: warning
`

func TestMutations(t *testing.T) {
	c, err := config.Parse([]byte(configYaml))
	assert.NoError(t, err)
	assert.Len(t, c.Mutations, 0)
	mutations := []string{"hostIPCSet", "pullPolicyNotAlways", "hostPIDSet", "hostNetworkSet", "deploymentMissingReplicas", "runAsRootAllowed", "cpuRequestsMissing", "cpuLimitsMissing", "memoryRequestsMissing", "memoryLimitsMissing", "livenessProbeMissing", "readinessProbeMissing"}
	for _, mutationStr := range mutations {
		for _, tc := range failureTestCasesMap[mutationStr] {
			newConfig := c
			key := fmt.Sprintf("%s/%s", tc.check, strings.ReplaceAll(tc.filename, "failure", "mutated"))
			mutatedYamlContent, ok := mutatedYamlContentMap[key]
			assert.True(t, ok)
			assert.Len(t, tc.resources.Resources, 1)
			newConfig.Mutations = []string{mutationStr}
			results, err := validator.ApplyAllSchemaChecksToResourceProvider(&newConfig, tc.resources)
			assert.NoError(t, err)
			assert.Len(t, results, 1)
			comments, allMutations := mutation.GetMutationsAndCommentsFromResults(results)
			assert.Len(t, allMutations, 1)
			for _, resources := range tc.resources.Resources {
				assert.Len(t, resources, 1)
				key := fmt.Sprintf("%s/%s/%s", resources[0].Kind, resources[0].Resource.GetName(), resources[0].Resource.GetNamespace())
				mutations := allMutations[key]
				mutated, err := mutation.ApplyAllSchemaMutations(&c, tc.resources, resources[0], mutations)
				assert.NoError(t, err)
				yamlContent, err := yaml.JSONToYAML(mutated.OriginalObjectJSON)
				assert.NoError(t, err)
				contentStr := mutation.UpdateMutatedContentWithComments(string(yamlContent), comments)
				assert.EqualValues(t, mutatedYamlContent, contentStr)
			}
		}
	}
}
