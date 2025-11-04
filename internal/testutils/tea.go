package testutils

import (
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

// WithTermSize adjust terminal size for tests
func WithTermSize(width, height int) teatest.TestOption {
	return teatest.WithInitialTermSize(width, height)
}

// RunModel runs a Bubble Tea model until it completes, and returns final output.
func RunModel(t *testing.T, m tea.Model, opts ...teatest.TestOption) io.Reader {
	t.Helper()
	tm := teatest.NewTestModel(t, m, opts...)
	return tm.FinalOutput(t)
}

// RunModelWithInteraction runs a model, sends messages, waits for finish, returns final model.
func RunModelWithInteraction(t *testing.T, m tea.Model, sendMsgs []tea.Msg, opts ...teatest.TestOption) tea.Model {
	t.Helper()
	tm := teatest.NewTestModel(t, m, opts...)

	for _, msg := range sendMsgs {
		tm.Send(msg)
	}

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	return tm.FinalModel(t)
}

// AssertModelField asserts a field on the final model using a selector function
func AssertModelField[T comparable](t *testing.T, finalModel tea.Model, fieldName string, getVal func(tea.Model) T, expected T) {
	t.Helper()
	actual := getVal(finalModel)
	if actual != expected {
		t.Errorf("model field %s: expected %v, got %v", fieldName, expected, actual)
	}
}

// WaitUntil waits for a condition on the output reader within timeout
func WaitUntil(t *testing.T, reader io.Reader, condition func([]byte) bool, timeout time.Duration) {
	t.Helper()
	teatest.WaitFor(t, reader, condition, teatest.WithDuration(timeout))
}
