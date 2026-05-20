package common

import (
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
)

type TestingSuite struct {
	suite.Suite
	Logger       Logger // Logger for test output
	TestCounters TestCounters
	Resources    Resources
}

type TestCounters struct {
	ProposalCounter           int
	ContractsCounter          int
	ContractsCounterPerSender map[string]uint64
	IBCV2PacketSequence       int
}

type Resources struct {
	TmpDirs        []string
	ChainA         *Chain
	ChainB         *Chain
	DkrPool        *dockertest.Pool
	DkrNet         *dockertest.Network
	HermesResource *dockertest.Resource

	ValResources map[string][]*dockertest.Resource
}

// ============================================================================
// LOGGER INTERFACE - ADD THIS SECTION
// ============================================================================

// Logger is the interface for test logging
type Logger interface {
	StartTest(testName string)
	PassTest(testName string)
	FailTest(testName string, err error)
	LogStep(stepName string, details ...interface{})
	LogInfo(format string, args ...interface{})
	LogError(format string, args ...interface{})
	LogSubTest(subTestName string)
	LogDebug(format string, args ...interface{})
	LogWarning(format string, args ...interface{})
	LogSuccess(format string, args ...interface{})
	LogFailure(format string, args ...interface{})
	LogSection(sectionName string)
	LogSeparator()
	LogWithTime(format string, args ...interface{})
	SetVerbose(verbose bool)
	IsVerbose() bool
	GetCurrentTest() string
	GetElapsedTime() time.Duration
}

// NewTestLogger is a variable that will be set by the e2e package
var NewTestLogger func(t testing.TB, verbose bool) Logger
