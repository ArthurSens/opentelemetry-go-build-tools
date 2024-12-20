package main

import (
	"testing"

	cdowners "github.com/hairyhenderson/go-codeowners"
	"github.com/joshdk/go-junit"
	"github.com/stretchr/testify/require"
)

func TestIngestArtifacts(t *testing.T) {
	rg := newReportGenerator()
	rg.ingestArtifacts("./testdata/junit", "./testdata/codeowners")

	expectedTestResults := map[string]junit.Suite{
		"package1": junit.Suite{
			Name:       "package1",
			Package:    "",
			Properties: map[string]string{"go.version": "go1.23.1 darwin/arm64"},
			Tests: []junit.Test{
				junit.Test{
					Name:      "TestFailure",
					Classname: "package1",
					Duration:  0,
					Status:    "failed",
					Message:   "Failed",
					Error: junit.Error{
						Message: "Failed",
						Type:    "",
						Body:    "=== RUN   TestFailure\n--- FAIL: TestFailure (0.00s)\n",
					},
					Properties: map[string]string{"classname": "package1", "name": "TestFailure", "time": "0.000000"},
					SystemOut:  "",
					SystemErr:  "",
				},
				junit.Test{
					Name:       "TestSucess",
					Classname:  "package1",
					Duration:   0,
					Status:     "passed",
					Message:    "",
					Properties: map[string]string{"classname": "package1", "name": "TestSucess", "time": "0.000000"},
					SystemOut:  "",
					SystemErr:  ""},
			},
			SystemOut: "",
			SystemErr: "",
			Totals: junit.Totals{
				Tests:    2,
				Passed:   1,
				Skipped:  0,
				Failed:   1,
				Error:    0,
				Duration: 0,
			},
		}, "package2": junit.Suite{
			Name:       "package2",
			Package:    "",
			Properties: map[string]string{"go.version": "go1.23.1 darwin/arm64"},
			Tests: []junit.Test{
				junit.Test{
					Name:      "TestFailure",
					Classname: "package2",
					Duration:  0,
					Status:    "failed",
					Message:   "Failed",
					Error: junit.Error{
						Message: "Failed",
						Type:    "",
						Body:    "=== RUN   TestFailure\n--- FAIL: TestFailure (0.00s)\n",
					}, Properties: map[string]string{"classname": "package2", "name": "TestFailure", "time": "0.000000"},
					SystemOut: "",
					SystemErr: ""},
				junit.Test{
					Name:       "TestSucess",
					Classname:  "package2",
					Duration:   0,
					Status:     "passed",
					Message:    "",
					Properties: map[string]string{"classname": "package2", "name": "TestSucess", "time": "0.000000"},
					SystemOut:  "",
					SystemErr:  ""},
			},
			SystemOut: "",
			SystemErr: "",
			Totals: junit.Totals{
				Tests:    2,
				Passed:   1,
				Skipped:  0,
				Failed:   1,
				Error:    0,
				Duration: 0},
		},
	}
	require.Equal(t, expectedTestResults, rg.testSuites)

	expectedCodeowners1, _ := cdowners.NewCodeowner("package1/*", []string{"@ArthurSens"})
	expectedCodeowners2, _ := cdowners.NewCodeowner("package2/*", []string{"@ArthurSens"})
	require.Equal(t, expectedCodeowners1, rg.codeowners.Patterns[0])
	require.Equal(t, expectedCodeowners2, rg.codeowners.Patterns[1])
}

func TestProcessTestResults(t *testing.T) {
	testCases := []struct {
		name       string
		testSuite  map[string]junit.Suite
		codeowners *cdowners.Codeowners

		expectedReports []report
	}{
		{
			name: "Default codeowners",
			testSuite: map[string]junit.Suite{
				"package1": junit.Suite{
					Name:  "package1",
					Tests: []junit.Test{{Name: "TestFailure", Status: junit.StatusFailed}},
					Totals: junit.Totals{
						Failed: 1,
					},
				},
				"package2": junit.Suite{
					Name:  "package2",
					Tests: []junit.Test{{Name: "TestFailure", Status: junit.StatusFailed}},
					Totals: junit.Totals{
						Failed: 1,
					},
				},
			},
			codeowners: func() *cdowners.Codeowners {
				c, _ := cdowners.NewCodeowner("*", []string{"@User1", "@User2"})
				c2, _ := cdowners.NewCodeowner("nonExistingPath/*", []string{"@NonExistentUser"})
				return &cdowners.Codeowners{Patterns: []cdowners.Codeowner{c, c2}}
			}(),

			expectedReports: []report{
				{module: "package1", codeOwners: "@User1, @User2", failedTests: []string{"TestFailure"}},
				{module: "package2", codeOwners: "@User1, @User2", failedTests: []string{"TestFailure"}},
			},
		},
		{
			name: "Overlapping codeowners",
			testSuite: map[string]junit.Suite{
				"package1": junit.Suite{
					Name:  "package1",
					Tests: []junit.Test{{Name: "TestFailure", Status: junit.StatusFailed}},
					Totals: junit.Totals{
						Failed: 1,
					},
				},
			},
			codeowners: func() *cdowners.Codeowners {
				c, _ := cdowners.NewCodeowner("*", []string{"@User1"})
				c2, _ := cdowners.NewCodeowner("package1/*", []string{"@User2"})
				return &cdowners.Codeowners{Patterns: []cdowners.Codeowner{c, c2}}
			}(),
			expectedReports: []report{
				{module: "package1", codeOwners: "@User2", failedTests: []string{"TestFailure"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rg := newReportGenerator()
			rg.testSuites = tc.testSuite
			rg.codeowners = tc.codeowners
			rg.processTestResults()

			require.Equal(t, tc.expectedReports, rg.reports)
		})
	}
}
