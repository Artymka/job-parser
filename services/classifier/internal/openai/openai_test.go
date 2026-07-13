package openai

import (
	"fmt"
	"os"
	"testing"

	"github.com/artymka/jobparser/services/classifier/internal/config"
)

func TestPrompt(t *testing.T) {
	os.Setenv("KAFKA_BROKER", "test")
	config, err := config.New("../../.env")
	if err != nil {
		panic(err)
	}

	client := NewClient(config)
	resp, err := client.Request(`Собрали все интервью стажёров в одном посте — читаем и вдохновляемся🥰

1️⃣ (https://t.me/start_v_rwb/193)Анастасия Ларина (https://t.me/start_v_rwb/193) поделилась впечатлениями, как прошла её стажировка в отделе внутренних процессов и аналитики.

2️⃣Будущий юрист Варвара Шерстяных (https://t.me/start_v_rwb/252) рассказала, как попала в договорной отдел.

3️⃣А в этом посте (https://t.me/start_v_rwb/271) стажёры из разных отделов поделились, какими задачами занимались.

4️⃣Две разные истории успеха в одной компании — интервью с близняшками (https://t.me/start_v_rwb/287) Аней и Машей.

5️⃣От стажёра до команды RWB: Андрей Сучков (https://t.me/start_v_rwb/532), комьюнити-менеджер, рассказал, как попал к нам в штат.`)

	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}
