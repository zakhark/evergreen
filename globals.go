package evergreen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tychoish/grip"
	"github.com/tychoish/grip/send"
	"github.com/tychoish/grip/slogger"
)

const (
	User = "mci"

	TestDir      = "config_test"
	TestSettings = "evg_settings.yml"

	HostRunning         = "running"
	HostTerminated      = "terminated"
	HostUninitialized   = "starting"
	HostInitializing    = "provisioning"
	HostProvisionFailed = "provision failed"
	HostUnreachable     = "unreachable"
	HostQuarantined     = "quarantined"
	HostDecommissioned  = "decommissioned"

	HostStatusSuccess = "success"
	HostStatusFailed  = "failed"

	TaskStarted      = "started"
	TaskUndispatched = "undispatched"
	TaskDispatched   = "dispatched"
	TaskFailed       = "failed"
	TaskSucceeded    = "success"
	TaskInactive     = "inactive"

	TestFailedStatus    = "fail"
	TestSkippedStatus   = "skip"
	TestSucceededStatus = "pass"

	BuildStarted   = "started"
	BuildCreated   = "created"
	BuildFailed    = "failed"
	BuildSucceeded = "success"

	VersionStarted   = "started"
	VersionCreated   = "created"
	VersionFailed    = "failed"
	VersionSucceeded = "success"

	PatchCreated   = "created"
	PatchStarted   = "started"
	PatchSucceeded = "succeeded"
	PatchFailed    = "failed"

	PushLogPushing = "pushing"
	PushLogSuccess = "success"

	HostTypeStatic = "static"

	CompileStage = "compile"
	TestStage    = "single_test"
	SanityStage  = "smokeCppUnitTests"
	PushStage    = "push"

	// maximum task (zero based) execution number
	MaxTaskExecution = 3

	// maximum task priority
	MaxTaskPriority = 100

	// LogMessage struct versions
	LogmessageFormatTimestamp = 1
	LogmessageCurrentVersion  = LogmessageFormatTimestamp

	EvergreenHome = "EVGHOME"

	DefaultTaskActivator   = ""
	StepbackTaskActivator  = "stepback"
	APIServerTaskActivator = "apiserver"

	AgentAPIVersion = 2
)

// evergreen package names
const (
	UIPackage = "EVERGREEN_UI"
)

const (
	AuthTokenCookie  = "mci-token"
	TaskSecretHeader = "Task-Secret"
	HostHeader       = "Host-Id"
	HostSecretHeader = "Host-Secret"
)

// HTTP constants. Added after Go1.4. Here for compatibility with GCCGO
// compatibility. Copied from: https://golang.org/pkg/net/http/#pkg-constants
const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH" // RFC 5789
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

var (
	// UphostStatus is a list of all host statuses that are considered "up."
	// This is used for query building.
	UphostStatus = []string{
		HostRunning,
		HostUninitialized,
		HostInitializing,
		HostProvisionFailed,
	}

	// Logger is our global logger. It can be changed for testing.
	Logger slogger.Logger

	// database and config directory, set to the testing version by default for safety
	NotificationsFile = "mci-notifications.yml"
	ClientDirectory   = "clients"

	// version requester types
	PatchVersionRequester       = "patch_request"
	RepotrackerVersionRequester = "gitter_request"

	// constant arrays for db update logic
	AbortableStatuses = []string{TaskStarted, TaskDispatched}
	CompletedStatuses = []string{TaskSucceeded, TaskFailed}
)

// SetLegacyLogger sets the global (s)logger instance to wrap the current grip Logger.
func SetLegacyLogger() {
	Logger = slogger.Logger{
		Name:      fmt.Sprintf("evg-gloal.%s", grip.Name()),
		Appenders: []send.Sender{grip.GetSender()},
	}
}

// FindEvergreenHome finds the directory of the EVGHOME environment variable.
func FindEvergreenHome() string {
	// check if env var is set
	root := os.Getenv(EvergreenHome)
	if len(root) > 0 {
		return root
	}

	Logger.Logf(slogger.ERROR, "%s is unset", EvergreenHome)
	return ""
}

// TestConfig creates test settings from a test config.
func TestConfig() *Settings {
	evgHome := FindEvergreenHome()
	file := filepath.Join(evgHome, TestDir, TestSettings)
	settings, err := NewSettings(file)
	if err != nil {
		panic(err)
	}
	return settings
}

// IsSystemActivator returns true when the task activator is Evergreen.
func IsSystemActivator(caller string) bool {
	return caller == DefaultTaskActivator ||
		caller == APIServerTaskActivator
}
