package pomodoro_test

import (
	"testing"

	"obigvee.com/pomo_cli/interactiveTool/pomo/obigvee.com/pomo_cli/interactiveTool/pomodoro"
	"obigvee.com/pomo_cli/interactiveTool/pomo/obigvee.com/pomo_cli/interactiveTool/pomodoro/repository"
)

func getRepo(t *testing.T) pomodoro.Repository {
	t.Helper()

	return repository.NewInMemoryRepo(), func() {}
}
