package tea

import (
	"errors"
	"strings"
	"testing"
)

// A failed insertion must not be acknowledged as completed: the program has to
// surface the terminal write error and stop before the model believes the
// native output landed.
func TestProgramDoesNotAcknowledgeFailedInsert(t *testing.T) {
	failure := errors.New("native output failed")
	model := &insertFailureModel{}
	program := NewProgram(model, WithInput(nil), WithWindowSize(40, 10),
		WithOutput(insertFailureWriter{err: failure}), WithoutSignalHandler())
	_, err := program.Run()
	if !errors.Is(err, failure) {
		t.Fatalf("lost terminal write error: %v", err)
	}
	if model.acknowledged {
		t.Fatal("acknowledged native output after terminal write failure")
	}
}

type insertDoneMsg struct{}

type insertFailureModel struct{ acknowledged bool }

func (*insertFailureModel) Init() Cmd {
	return Sequence(Println("FAILED-INSERT"), func() Msg { return insertDoneMsg{} })
}

func (m *insertFailureModel) Update(message Msg) (Model, Cmd) {
	if _, ok := message.(insertDoneMsg); ok {
		m.acknowledged = true
		return m, Quit
	}
	return m, nil
}

func (*insertFailureModel) View() View { return NewView("COMPOSER\nSTATUS") }

type insertFailureWriter struct{ err error }

func (w insertFailureWriter) Write(p []byte) (int, error) {
	if strings.Contains(string(p), "FAILED-INSERT") {
		return 0, w.err
	}
	return len(p), nil
}
