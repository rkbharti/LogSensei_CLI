package analyzer

import "strings"

// Explanation holds a human-readable reason and fix suggestion.
type Explanation struct {
	Reason     string
	Suggestion string
}

type explanationRule struct {
	keyword    string
	reason     string
	suggestion string
}

var explanationRegistry = []explanationRule{

	// ── Oracle / Network OS errors ────────────────────────────────────────────
	{
		keyword:    "setsockopt",
		reason:     "A socket option (multicast/network) could not be set on the interface.",
		suggestion: "Check if the network adapter supports multicast. Try disabling MCAST or binding to a different interface.",
	},
	{
		keyword:    "mcast_add",
		reason:     "Failed to join a multicast group — the address may not be available on this interface.",
		suggestion: "Verify the network interface supports multicast. Check with 'netstat -g' or review adapter settings.",
	},
	{
		keyword:    "edc8116",
		reason:     "EDC8116I — Address not available. The requested IP or socket address cannot be used.",
		suggestion: "Ensure the IP address is assigned to a valid network interface. Check TCP/IP stack configuration.",
	},
	{
		keyword:    "edc",
		reason:     "An EDC (TCP/IP) error code was raised by the network stack.",
		suggestion: "Look up the specific EDC code in IBM TCP/IP documentation. Check socket and network configuration.",
	},
	{
		keyword:    "proterr",
		reason:     "A protocol state machine received an unexpected event for its current state.",
		suggestion: "Check RSVP/network signaling configuration. The flow may have been torn down unexpectedly.",
	},
	{
		keyword:    "rsvp",
		reason:     "RSVP (Resource Reservation Protocol) encountered a state or signaling error.",
		suggestion: "Review RSVP session configuration. Check if PATHTEAR/RESVERR events are expected in your setup.",
	},
	{
		keyword:    "pathtear",
		reason:     "An RSVP PATH teardown was received in an unexpected session state.",
		suggestion: "Investigate whether the upstream RSVP node terminated the session prematurely.",
	},
	{
		keyword:    "mailslot",
		reason:     "A mailslot (IPC communication channel) failed during creation or socket setup.",
		suggestion: "Check system IPC limits and network interface binding. Review mailslot service configuration.",
	},
	{
		keyword:    "ora-",
		reason:     "An Oracle database error occurred.",
		suggestion: "Look up the ORA- error code on Oracle docs. Common causes: missing table, bad credentials, or connection issue.",
	},
	{
		keyword:    "tns:",
		reason:     "Oracle TNS listener error — could not resolve service name or connect to listener.",
		suggestion: "Check tnsnames.ora and listener.ora. Verify Oracle listener is running with 'lsnrctl status'.",
	},

	// ── Go runtime errors ─────────────────────────────────────────────────────
	{
		keyword:    "nil pointer",
		reason:     "A variable is being used before it was initialized (nil pointer dereference).",
		suggestion: "Check if the variable is nil before using it. Use guard clauses: if x == nil { return }",
	},
	{
		keyword:    "invalid memory address",
		reason:     "A variable is being used before it was initialized (nil pointer dereference).",
		suggestion: "Check if the variable is nil before using it. Use guard clauses: if x == nil { return }",
	},
	{
		keyword:    "index out of range",
		reason:     "You are accessing a slice or array at an index that does not exist.",
		suggestion: "Check the slice length before indexing. Use len(slice) > index as a guard.",
	},
	{
		keyword:    "stack overflow",
		reason:     "A function is calling itself infinitely (infinite recursion).",
		suggestion: "Check your recursive function for a proper base/termination case.",
	},
	{
		keyword:    "deadlock",
		reason:     "All goroutines are waiting on each other — the program is stuck.",
		suggestion: "Check for circular mutex locks or channels that are never read/written.",
	},
	{
		keyword:    "goroutine leak",
		reason:     "A goroutine was started but never terminated, causing memory to grow.",
		suggestion: "Ensure every goroutine has a clear exit path. Use context cancellation.",
	},

	// ── network / timeout ─────────────────────────────────────────────────────
	{
		keyword:    "connection refused",
		reason:     "The target server is not running or is not accepting connections.",
		suggestion: "Verify the server is running and the port is correct. Check firewall rules.",
	},
	{
		keyword:    "connection reset",
		reason:     "The connection was forcibly closed by the remote server.",
		suggestion: "Check server-side logs for why the connection was dropped. May be a crash or restart.",
	},
	{
		keyword:    "timeout",
		reason:     "The operation took too long and timed out.",
		suggestion: "Check network latency, server load, or increase the timeout limit.",
	},
	{
		keyword:    "no such host",
		reason:     "The hostname could not be resolved via DNS.",
		suggestion: "Check the hostname for typos. Verify DNS settings and network connectivity.",
	},
	{
		keyword:    "tls",
		reason:     "A TLS/SSL handshake or certificate error occurred.",
		suggestion: "Check certificate validity, expiry, and whether the CA is trusted.",
	},

	// ── database errors ───────────────────────────────────────────────────────
	{
		keyword:    "connection failed",
		reason:     "The application failed to connect to the database.",
		suggestion: "Check database host, port, credentials, and whether the DB server is running.",
	},
	{
		keyword:    "database",
		reason:     "A database operation failed.",
		suggestion: "Check database server status, credentials, and network connection.",
	},
	{
		keyword:    "deadlock found",
		reason:     "Two database transactions are blocking each other.",
		suggestion: "Review transaction order and add retry logic for deadlock errors.",
	},
	{
		keyword:    "duplicate entry",
		reason:     "A unique constraint was violated in the database.",
		suggestion: "Check for duplicate data before inserting. Use INSERT IGNORE or ON CONFLICT.",
	},

	// ── authentication / authorization ────────────────────────────────────────
	{
		keyword:    "unauthorized",
		reason:     "The request lacks valid authentication credentials.",
		suggestion: "Check API keys, JWT tokens, or session credentials. Ensure they are not expired.",
	},
	{
		keyword:    "forbidden",
		reason:     "The authenticated user does not have permission for this action.",
		suggestion: "Check role/permission configuration. Verify the user has required access level.",
	},
	{
		keyword:    "token expired",
		reason:     "The authentication token has expired.",
		suggestion: "Implement token refresh logic. Check token TTL configuration.",
	},

	// ── file system errors ────────────────────────────────────────────────────
	{
		keyword:    "no such file or directory",
		reason:     "A required file or directory does not exist at the expected path.",
		suggestion: "Verify the file path. Check working directory and relative vs absolute paths.",
	},
	{
		keyword:    "permission denied",
		reason:     "The process does not have permission to access the file or resource.",
		suggestion: "Check file permissions with ls -la. Run with appropriate privileges or fix ownership.",
	},
	{
		keyword:    "disk full",
		reason:     "The disk has run out of space.",
		suggestion: "Free up disk space. Check with df -h. Consider log rotation or storage expansion.",
	},

	// ── JavaScript / frontend ─────────────────────────────────────────────────
	{
		keyword:    "cannot read property",
		reason:     "You are accessing a property on an undefined or null object.",
		suggestion: "Ensure the object exists before accessing its properties. Use optional chaining: obj?.property",
	},
	{
		keyword:    "is not a function",
		reason:     "You are calling something that is not a function.",
		suggestion: "Check the variable type before calling it. It may be undefined or overwritten.",
	},
	{
		keyword:    "syntaxerror",
		reason:     "There is a JavaScript syntax error in the code.",
		suggestion: "Check for missing brackets, quotes, or semicolons near the reported line.",
	},

	// ── Python errors ─────────────────────────────────────────────────────────
	{
		keyword:    "traceback",
		reason:     "A Python exception occurred — see the traceback for the call stack.",
		suggestion: "Read the last line of the traceback for the actual error. Fix the root cause shown there.",
	},
	{
		keyword:    "attributeerror",
		reason:     "You are accessing an attribute that does not exist on the object.",
		suggestion: "Check the object type with type(obj). Verify the attribute name is correct.",
	},
	{
		keyword:    "importerror",
		reason:     "A Python module could not be imported.",
		suggestion: "Run pip install <module> or check your virtual environment is activated.",
	},
	{
		keyword:    "keyerror",
		reason:     "A dictionary key does not exist.",
		suggestion: "Use dict.get(key, default) instead of dict[key] to safely access keys.",
	},

	// ── generic fallback rules ────────────────────────────────────────────────
	{
		keyword:    "failed",
		reason:     "An operation failed during execution.",
		suggestion: "Check the surrounding log lines for the root cause. Look for resource or permission issues.",
	},
	{
		keyword:    "fatal",
		reason:     "A fatal error caused the process to terminate.",
		suggestion: "Investigate the root cause immediately. Check system resources and application state.",
	},
	{
		keyword:    "out of memory",
		reason:     "The process ran out of available memory.",
		suggestion: "Increase memory limits, optimize memory usage, or check for memory leaks.",
	},
}

func ExplainError(message string) Explanation {
	lower := strings.ToLower(message)

	for _, rule := range explanationRegistry {
		if strings.Contains(lower, rule.keyword) {
			return Explanation{
				Reason:     rule.reason,
				Suggestion: rule.suggestion,
			}
		}
	}

	return Explanation{
		Reason:     "Unknown error occurred.",
		Suggestion: "Check logs and debug manually.",
	}
}
