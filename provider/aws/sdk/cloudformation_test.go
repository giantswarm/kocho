package sdk

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/giantswarm/kocho/provider/aws/types"
)

func pointer(s string) *string {
	return &s
}

// TestFromCloudFormationTags checks if fromCloudFormationTags is able to
// properly translate a list of cloudformation.Tag to a list of types.Tag.
func TestFromCloudFormationTags(t *testing.T) {
	testCases := []struct {
		CloudformationTags []*cloudformation.Tag
		KochoTags          []types.Tag
	}{
		{
			[]*cloudformation.Tag{},
			[]types.Tag{},
		},
		{
			[]*cloudformation.Tag{
				{Key: pointer("test-key-1"), Value: pointer("test-value-1")},
			},
			[]types.Tag{
				{Key: "test-key-1", Value: "test-value-1"},
			},
		},
		{
			[]*cloudformation.Tag{
				{Key: pointer("test-key-1"), Value: pointer("test-value-1")},
				{Key: pointer("test-key-2"), Value: pointer("test-value-2")},
				{Key: pointer("test-key-3"), Value: pointer("test-value-3")},
			},
			[]types.Tag{
				{Key: "test-key-1", Value: "test-value-1"},
				{Key: "test-key-2", Value: "test-value-2"},
				{Key: "test-key-3", Value: "test-value-3"},
			},
		},
	}

	for _, testCase := range testCases {
		tags := fromCloudFormationTags(testCase.CloudformationTags)
		if !reflect.DeepEqual(testCase.KochoTags, tags) {
			t.Fatalf("expected tags '%#v' to equal '%#v'", tags, testCase.KochoTags)
		}
	}
}
