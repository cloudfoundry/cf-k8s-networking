package matchers

import (
	"fmt"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/gexec"
)

// Copied from gexec.exit_matcher

/*
The Exit matcher operates on a session:

	Expect(session).Should(ExitSuccessfully())

Exit passes if the session has already exited and has exited with code 0.

Note that the process must have already exited.  To wait for a process to exit, use Eventually:

	Eventually(session, 3).Should(ExitSuccessfully())
*/
func ExitSuccessfully() *exitMatcher {
	return &exitMatcher{}
}

type exitMatcher struct {
	actualExitCode int
}

type Exiter interface {
	ExitCode() int
}

func (m *exitMatcher) Match(actual interface{}) (success bool, err error) {
	exiter, ok := actual.(Exiter)
	if !ok {
		return false, fmt.Errorf("ExitSuccessfully must be passed a gexec.Exiter (Missing method ExitCode() int) Got:\n%s", format.Object(actual, 1))
	}

	m.actualExitCode = exiter.ExitCode()

	return 0 == m.actualExitCode, nil
}

func (m *exitMatcher) FailureMessage(actual interface{}) (message string) {
	session, ok := actual.(*gexec.Session)
	if !ok {
		panic("ExitSuccessfully must be passed a gexec.Session")
	}

	if m.actualExitCode == -1 {
		return "Expected process to exit.  It did not."
	}
	stdout := string(session.Out.Contents())
	stderr := string(session.Err.Contents())

	return format.Message(m.actualExitCode, fmt.Sprintf("to match exit code 0.\nSTDOUT:\n%s\nSTDERR:\n%s\n", stdout, stderr))
}

func (m *exitMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(m.actualExitCode, "not to match exit code 0")
}

func (m *exitMatcher) MatchMayChangeInTheFuture(actual interface{}) bool {
	session, ok := actual.(*gexec.Session)
	if ok {
		return session.ExitCode() == -1
	}
	return true
}
