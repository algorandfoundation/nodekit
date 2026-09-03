package explanations

// SudoWarningMsg is a constant string displayed to warn users that they may be prompted for their password during execution.
const SudoWarningMsg = "(You may be prompted for your password)"

// PermissionErrorMsg is a constant string that indicates a command requires super-user privileges (sudo) to be executed.
const PermissionErrorMsg = "this command must be run with super-user privileges (sudo)"

// AlgorandPermissionErrorMsg is a constant string that indicates a command requires additional permissions to be executed.
const AlgorandPermissionErrorMsg = "this command requires additional permissions, run with super-user (sudo) and if you're on Linux consider adding your account to the 'algorand' group after"

// NotInstalledErrorMsg is the error message displayed when the algod software is not installed on the system.
const NotInstalledErrorMsg = "algod is not installed. please run the *install* command"

// RunningErrorMsg represents the error message displayed when algod is running and needs to be stopped before proceeding.
const RunningErrorMsg = "algod is running, please run the *stop* command"

// NotRunningErrorMsg is the error message displayed when the algod service is not currently running on the system.
const NotRunningErrorMsg = "algod is not running"

// NotRunningStartErrorMsg is the error message displayed when algod needs to already be running and has to be started first.
const NotRunningStartErrorMsg = "algod is not running, please run the *start* command"

// NotSuperUserErrorMsg is the error message displayed when a non-superuser tries to execute a command requiring root privileges.
const NotSuperUserErrorMsg = "you need to be root to run this command. Please run this command with sudo"
