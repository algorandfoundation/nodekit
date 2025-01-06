package explanations

// SudoWarningMsg is a constant string displayed to warn users that they may be prompted for their password during execution.
const SudoWarningMsg = "(You may be prompted for your password)"

// PermissionErrorMsg is a constant string that indicates a command requires super-user privileges (sudo) to be executed.
const PermissionErrorMsg = "this command must be run with super-user privileges (sudo)"

// NotInstalledErrorMsg is the error message displayed when the algod software is not installed on the system.
const NotInstalledErrorMsg = "algod is not installed. please run the *install* command"

// RunningErrorMsg represents the error message displayed when algod is running and needs to be stopped before proceeding.
const RunningErrorMsg = "algod is running, please run the *stop* command"

// NotRunningErrorMsg is the error message displayed when the algod service is not currently running on the system.
const NotRunningErrorMsg = "algod is not running"
