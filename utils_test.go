package tags

import (
	"os"

	"github.com/brianvoe/gofakeit"
	"github.com/go-catupiry/catu"
	"github.com/pkg/errors"
)

var appInstance catu.App

func GetAppInstance() catu.App {
	if appInstance != nil {
		return appInstance
	}

	os.Setenv("DB_URI", "file::memory:?cache=shared")
	os.Setenv("DB_ENGINE", "sqlite")
	// os.Setenv("LOG_QUERY", "1")

	app := catu.Init(&catu.AppOptions{})
	err := app.Bootstrap()
	if err != nil {
		panic(err)
	}
	// fake content stub for tests:
	err = app.GetDB().AutoMigrate(
		&ContentModelStub{},
		&VocabularyModel{},
		&TermModel{},
		&ModelstermsModel{},
	)

	if err != nil {
		panic(errors.Wrap(err, "taxonomy.GetAppInstance Error on run auto migration"))
	}

	return app
}

type ContentModelStub struct {
	ID         uint64 `json:"id"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Published  bool   `json:"published"`
	ClickCount int64  `json:"clickCount"`
	Secret     string `json:"-"`
	Email      string `json:"email"`
	Email2     string `json:"email2"`
	PrivateBio string `json:"-"`
}

func GetContentModelStub() ContentModelStub {
	return ContentModelStub{
		// ID:         gofakeit.Uint64(),
		Title:      gofakeit.Paragraph(1, 4, 4, " "),
		Body:       gofakeit.Paragraph(1, 3, 5, " "),
		Published:  true,
		Secret:     gofakeit.Word(),
		Email:      gofakeit.Email(),
		Email2:     gofakeit.Email(),
		PrivateBio: gofakeit.Paragraph(1, 4, 4, ""),
	}
}
