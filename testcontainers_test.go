package testcontainers

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"testing"
)

func TestSessionID(t *testing.T) {
	t.Run("SessionID() returns a non-empty string", func(t *testing.T) {
		sessionID := SessionID()
		if sessionID == "" {
			t.Error("SessionID() returned an empty string")
		}
	})

	t.Run("Multiple calls to SessionID() return the same value", func(t *testing.T) {
		sessionID1 := SessionID()
		sessionID2 := SessionID()
		if sessionID1 != sessionID2 {
			t.Errorf("SessionID() returned different values: %s != %s", sessionID1, sessionID2)
		}
	})

	t.Run("Multiple calls to SessionID() in multiple goroutines return the same value", func(t *testing.T) {
		sessionID1 := ""
		sessionID2 := ""

		done := make(chan bool)
		go func() {
			sessionID1 = SessionID()
			done <- true
		}()

		go func() {
			sessionID2 = SessionID()
			done <- true
		}()

		<-done
		<-done

		if sessionID1 != sessionID2 {
			t.Errorf("SessionID() returned different values: %s != %s", sessionID1, sessionID2)
		}
	})

	t.Run("SessionID() from different child processes returns the same value", func(t *testing.T) {
		args := []string{"test", "./...", "-v", "-run", "TestSessionIDHelper"}
		env := append(os.Environ(), "TESTCONTAINERS_SESSION_ID_HELPER=1")

		re := regexp.MustCompile(">>>(.*)<<<")

		cmd1 := exec.Command("go", args...)
		cmd1.Env = env
		stdoutStderr1, err := cmd1.CombinedOutput()
		if err != nil {
			t.Errorf("cmd1.Run() failed with %s", err)
		}
		sessionID1 := re.FindString(string(stdoutStderr1))

		cmd2 := exec.Command("go", args...)
		cmd2.Env = env
		stdoutStderr2, err := cmd2.CombinedOutput()
		if err != nil {
			t.Errorf("cmd2.Run() failed with %s", err)
		}
		sessionID2 := re.FindString(string(stdoutStderr2))

		if sessionID1 != sessionID2 {
			t.Errorf("SessionID() returned different values: %s != %s", sessionID1, sessionID2)
		}
	})
}

// Not a real test, used to print out the session ID
func TestSessionIDHelper(t *testing.T) {
	if os.Getenv("TESTCONTAINERS_SESSION_ID_HELPER") == "" {
		t.Skip("Not a real test, used as a test helper")
	}

	fmt.Printf(">>>%s<<<\n", SessionID())
}
