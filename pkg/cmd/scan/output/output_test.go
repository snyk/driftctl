package output

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/r3labs/diff/v2"
)

func fakeAnalysis() *analyser.Analysis {
	a := analyser.Analysis{}
	a.AddUnmanaged(
		&testresource.FakeResource{
			Id:   "unmanaged-id-1",
			Type: "aws_unmanaged_resource",
		},
		&testresource.FakeResource{
			Id:   "unmanaged-id-2",
			Type: "aws_unmanaged_resource",
		},
	)
	a.AddDeleted(
		&testresource.FakeResource{
			Id:   "deleted-id-1",
			Type: "aws_deleted_resource",
		}, &testresource.FakeResource{
			Id:   "deleted-id-2",
			Type: "aws_deleted_resource",
		},
	)
	a.AddManaged(
		&testresource.FakeResource{
			Id:   "diff-id-1",
			Type: "aws_diff_resource",
		},
		&testresource.FakeResource{
			Id:   "no-diff-id-1",
			Type: "aws_no_diff_resource",
		},
	)
	a.AddDifference(analyser.Difference{Res: &testresource.FakeResource{
		Id:   "diff-id-1",
		Type: "aws_diff_resource",
	}, Changelog: []diff.Change{
		{
			Type: diff.UPDATE,
			Path: []string{"updated", "field"},
			From: "foobar",
			To:   "barfoo",
		},
		{
			Type: diff.CREATE,
			Path: []string{"new", "field"},
			From: nil,
			To:   "newValue",
		},
		{
			Type: diff.DELETE,
			Path: []string{"a"},
			From: "oldValue",
			To:   nil,
		},
	}})
	return &a
}

func fakeAnalysisNoDrift() *analyser.Analysis {
	a := analyser.Analysis{}
	for i := 0; i < 5; i++ {
		a.AddManaged(&testresource.FakeResource{
			Id:   "managed-id-" + fmt.Sprintf("%d", i),
			Type: "aws_managed_resource",
		})
	}
	return &a
}

func fakeAnalysisWithJsonFields() *analyser.Analysis {
	a := analyser.Analysis{}
	a.AddManaged(
		&testresource.FakeResource{
			Id:   "diff-id-1",
			Type: "aws_diff_resource",
		},
	)
	a.AddManaged(
		&testresource.FakeResource{
			Id:   "diff-id-2",
			Type: "aws_diff_resource",
		},
	)
	a.AddDifference(analyser.Difference{Res: &testresource.FakeResource{
		Id:   "diff-id-1",
		Type: "aws_diff_resource",
	}, Changelog: []diff.Change{
		{
			Type: diff.UPDATE,
			Path: []string{"Json"},
			From: "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Removed\":\"Added\",\"Changed\":[\"ec2:DescribeInstances\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}",
			To:   "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Changed\":[\"ec2:*\"],\"NewField\":[\"foobar\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}",
		},
	}})
	a.AddDifference(analyser.Difference{Res: &testresource.FakeResource{
		Id:   "diff-id-2",
		Type: "aws_diff_resource",
	}, Changelog: []diff.Change{
		{
			Type: diff.UPDATE,
			Path: []string{"Json"},
			From: "{\"foo\":\"bar\"}",
			To:   "{\"bar\":\"foo\"}",
		},
	}})
	return &a
}

func fakeAnalysisWithStringerResources() *analyser.Analysis {
	a := analyser.Analysis{}
	a.AddDeleted(
		&testresource.FakeResourceStringer{
			Id:   "dfjkgnbsgj",
			Name: "deleted resource",
		},
	)
	a.AddManaged(
		&testresource.FakeResourceStringer{
			Id:   "usqyfsdbgjsdgjkdfg",
			Name: "managed resource",
		},
	)
	a.AddUnmanaged(
		&testresource.FakeResourceStringer{
			Id:   "duysgkfdjfdgfhd",
			Name: "unmanaged resource",
		},
	)
	a.AddDifference(analyser.Difference{Res: &testresource.FakeResourceStringer{
		Id:   "gdsfhgkbn",
		Name: "resource with diff",
	}, Changelog: []diff.Change{
		{
			Type: diff.UPDATE,
			Path: []string{"Name"},
			From: "",
			To:   "resource with diff",
		},
	}})
	return &a
}
