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

// LogsNotFoundErrorMsg is the error message displayed when the node's log file does not exist.
// It takes the resolved path, which is not always inside the data directory.
const LogsNotFoundErrorMsg = "no log file at %s. algod creates it on first start, please run the *start* command"

// LogsPermissionErrorMsg is the error message displayed when the node's own log
// file, or the data directory holding it, cannot be read.
const LogsPermissionErrorMsg = "cannot read the node's log file: permission denied. The node's files are owned by the 'algorand' user, so run this command with super-user (sudo); on Linux you can instead add your account to the 'algorand' group and open a new session for that to take effect"

// LogsFilePermissionErrorMsg is the error message displayed when a file named by
// --file cannot be read. It takes the path: that file is the user's own choice
// and has nothing to do with how a node's data directory is owned.
const LogsFilePermissionErrorMsg = "cannot read %s: permission denied"

// LogsFollowCompressedErrorMsg is the error message displayed when --follow is asked for a
// compressed archive, which is a rotated file that nothing is appending to.
const LogsFollowCompressedErrorMsg = "cannot follow a compressed archive: it is a rotated log, not the file the node is writing to. Drop --file to follow the live log, or read this one without --follow"

// LogsToStdoutErrorMsg is the error message displayed when algod is configured to log to
// standard output instead of a file, which it does when config.json sets LogSizeLimit to 0.
const LogsToStdoutErrorMsg = "this node is configured to log to stdout, not to a log file (config.json sets LogSizeLimit to 0). On Linux the output goes to the systemd journal, on macOS to /tmp/algod.out"

// LogsEmptyMsg is displayed when the node's log file exists but has no content yet.
const LogsEmptyMsg = "the log file is empty, the node has not logged anything yet"

// LogsNoMatchMsg is displayed when the log has content but nothing passed the active filters.
const LogsNoMatchMsg = "no log entries matched. Try *--all* to include every level, or widen *--since*"
