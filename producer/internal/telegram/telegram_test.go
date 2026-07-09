package telegram

import (
	"testing"

	"github.com/artymka/jobparser-producer/internal/config"
)

func main() {
}

func TestAuth(t *testing.T) {
	config, err := config.New("../../.env")
	config.SessionPath = "../../sessions/session.json"
	if err != nil {
		panic(err)
	}

	t.Run("auth test", func(t *testing.T) {
		GetLastMessages(config)
	})
}
