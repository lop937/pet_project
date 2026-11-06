package tests

import (
	"fmt"
	"github.com/SlavaNovikov/grif/pkg/database"
	"github.com/SlavaNovikov/grif/internal/ds"
	"github.com/SlavaNovikov/grif/internal/pager"
	"github.com/SlavaNovikov/grif/internal/product"
	"github.com/SlavaNovikov/grif/internal/user"
	"github.com/SlavaNovikov/grif/internal/work"
	"github.com/SlavaNovikov/grif/migrations"
	"github.com/gorilla/schema"
	_ "github.com/mattn/go-sqlite3"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

var fullNames = []string{
	"Иванов Иван Иванович",
	"Петрова Анна Сергеевна",
	"Смирнов Дмитрий Владимирович",
	"Кузнецова Елена Павловна",
	"Соколов Максим Игоревич",
	"Михайлова Ольга Александровна",
	"Новиков Артём Николаевич",
	"Фёдорова Мария Андреевна",
	"Васильев Кирилл Викторович",
	"Зайцева Ксения Григорьевна",
}

var encoderSchema = schema.NewEncoder()

func TestGenerateContent(t *testing.T) {
	// создадим новую базу данных и наполним её
	// todo тут нужно убирать удаление, а то...
	_ = os.Remove("../store.db")
	db := database.NewSQLite("../store.db", migrations.InitSQL)
	defer db.Close()

	getRequest := func(dst interface{}) *http.Request {
		form := url.Values{}
		if err := encoderSchema.Encode(dst, form); err != nil {
			t.Fatal(err)
		}
		return &http.Request{Method: http.MethodPost, PostForm: form}
	}

	h := pager.New(db, nil, pager.Nav, pager.Registry)

	// создадим юзеров
	for i := range fullNames {
		if _, err := h.FactoryPager(ds.UsersRoute, db).Upsert(t.Context(), getRequest(user.User{
			FIO:  fullNames[i],
			Rate: 1,
		})); err != nil {
			t.Fatal(err)
		}
	}

	// создадим продукты
	for i := 0; i < 100; i++ {
		if _, err := h.FactoryPager(ds.ProductRoute, db).Upsert(t.Context(), getRequest(product.Product{
			Name:  fmt.Sprintf("изделие №%d", i),
			Tag:   "tag",
			Price: int64(i * 100),
		})); err != nil {
			t.Fatal(err)
		}
	}

	now := time.Now()
	startDate := now.AddDate(0, -1, 0)
	for d := startDate; !d.After(now); d = d.AddDate(0, 0, 1) {
		for k := 0; k < 5; k++ {
			dayStr := d.Format(ds.LayoutEventDate)
			if _, err := h.FactoryPager(ds.WorksRoute, db).Upsert(t.Context(), getRequest(work.Work{
				Date:      dayStr,
				Name:      "work " + dayStr,
				Price:     1000,
				UserID:    ds.ID(rand.Intn(len(fullNames))),
				ProductID: ds.ID(rand.Intn(100)),
				Quantity:  1,
			})); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func generateEnglishNickname() string {
	// Списки частей никнеймов
	adjectives := []string{"Fast", "Silent", "Shadow", "Crazy", "Lucky", "Steel"}
	nouns := []string{"Wolf", "Fox", "Bear", "Eagle", "Ghost", "Tiger"}
	numbers := []string{"", "42", "777", "99", "2k25", "101"}

	// Выбираем случайные части из списков
	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	num := numbers[rand.Intn(len(numbers))]

	// Собираем никнейм
	return fmt.Sprintf("%s%s%s", adj, noun, num)
}
