package testtools

type TestScenario struct {
	name   string
	input  []testData
	output []testData
}

type testData struct {
	name    string
	content string
}
