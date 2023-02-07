package tags

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-catupiry/catu"
	"github.com/go-catupiry/catu/helpers"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TermModel struct {
	ID             uint64    `gorm:"primaryKey;column:id" json:"id" filter:"param:id;type:number"`
	Text           string    `gorm:"column:text;type:varchar(255);not null" json:"text" filter:"param:text;type:string"`
	Description    string    `gorm:"column:description;type:text" json:"description" filter:"param:description;type:string"`
	VocabularyName string    `gorm:"column:vocabularyName;type:varchar(255);not null;default:Tags" json:"vocabularyName" filter:"param:vocabularyName;type:string"`
	CreatedAt      time.Time `gorm:"column:createdAt;type:datetime;not null" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updatedAt;type:datetime;not null" json:"updatedAt"`

	LinkPermanent string `gorm:"-" json:"linkPermanent"`
}

// TableName - Set db table name for term model
func (r *TermModel) TableName() string {
	return "terms"
}

func (r *TermModel) GetIDString() string {
	return strconv.FormatInt(int64(r.ID), 10)
}

func (r *TermModel) ToJSON() string {
	jsonString, _ := json.MarshalIndent(r, "", "  ")
	return string(jsonString)
}

// Save - Create if is new or update
func (m *TermModel) Save() error {
	var err error
	db := catu.GetDefaultDatabaseConnection()

	if m.ID == 0 {
		// create ....
		r := db.Create(m)
		if r.Error != nil {
			return r.Error
		}
	} else {
		// update ...
		err = db.Save(m).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TermModel) LoadTeaserData() error {
	r.LoadPath()
	return nil
}

func (r *TermModel) LoadData() error {
	return r.LoadTeaserData()
}

func (r *TermModel) GetPath() string {
	path := ""

	subPath := "/vocabulary/"

	if r.VocabularyName == "" {
		subPath += "Tags/"
	} else {
		subPath += r.VocabularyName + "/"
	}

	if r.ID != 0 {
		path += subPath + "term/" + r.GetIDString()
	}

	return path
}

func (r *TermModel) LoadPath() error {
	app := catu.GetApp()
	r.LinkPermanent = app.GetConfiguration().Get("APP_ORIGIN") + r.GetPath()
	return nil
}

func (r *TermModel) Delete() error {
	db := catu.GetDefaultDatabaseConnection()
	return db.Unscoped().Delete(r).Error
}

func NewTerm() (TermModel, error) {
	r := TermModel{}
	return r, nil
}

// Find One term by ID
func TermFindOne(id string, record *TermModel) error {
	db := catu.GetDefaultDatabaseConnection()

	err := db.First(&record, id).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	return err
}

// Find One term by vocabulary / term
func TermFindOneByText(text, vocabularyName string, record *TermModel) error {
	db := catu.GetDefaultDatabaseConnection()

	err := db.Where("text = ? AND vocabularyName = ?", text, vocabularyName).
		First(record).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

func TermFindManyByText(texts []string, vocabularyName string, records *[]TermModel) error {
	db := catu.GetDefaultDatabaseConnection()

	err := db.Where("text IN ? AND vocabularyName = ?", texts, vocabularyName).
		Find(records).Error
	if err != nil {
		return err
	}

	return nil
}

func FindTermTextsAndCount(ctx *catu.RequestContext) ([]string, int64, error) {
	var terms []string

	var count int64
	var records []TermModel
	err := TermQueryAndCountReq(&TermQueryOpts{
		Records: &records,
		Count:   &count,
		Limit:   ctx.GetLimit(),
		Offset:  ctx.GetOffset(),
		C:       ctx,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Debug("FindTermTextsAndCount Error on find terms")
		return nil, 0, err
	}

	for _, r := range records {
		terms = append(terms, r.Text)
	}

	return terms, count, nil
}

type TermQueryOpts struct {
	Records *[]TermModel
	Count   *int64
	Limit   int
	Offset  int
	C       echo.Context
	IsHTML  bool
}

func TermQueryAndCountReq(opts *TermQueryOpts) error {
	db := catu.GetDefaultDatabaseConnection()

	c := opts.C

	q := c.QueryParam("q")

	text := c.QueryParam("text")
	term := c.QueryParam("term")

	if text == "" && term != "" {
		text = term
	}

	vocabularyName := c.Param("vocabulary")

	query := db

	ctx := c.(*catu.RequestContext)

	queryI, err := ctx.Query.SetDatabaseQueryForModel(query, &TermModel{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%+v\n", err),
		}).Error("TermQueryAndCountReq error")
	}
	query = queryI.(*gorm.DB)

	if q != "" {
		query = query.Where(
			db.Where("text LIKE ?", "%"+q+"%").
				Or(db.Where("description LIKE ?", "%"+q+"%")),
		)
	}

	if vocabularyName != "" {
		query = query.Where("vocabularyName = ?", vocabularyName)
	}

	if text != "" {
		query = query.Where("text LIKE ?", text+"%")
	}

	orderColumn, orderIsDesc, orderValid := helpers.ParseUrlQueryOrder(c.QueryParam("order"), c.QueryParam("sort"), c.QueryParam("sortDirection"))

	if orderValid {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: clause.CurrentTable, Name: orderColumn},
			Desc:   orderIsDesc,
		})
	} else {
		query = query.
			Order("createdAt DESC").
			Order("id DESC")
	}

	query = query.Limit(opts.Limit).
		Offset(opts.Offset)

	r := query.Find(opts.Records)
	if r.Error != nil {
		return r.Error
	}

	return TermCountReq(opts)
}

func TermCountReq(opts *TermQueryOpts) error {
	db := catu.GetDefaultDatabaseConnection()

	c := opts.C

	q := c.QueryParam("q")

	ctx := c.(*catu.RequestContext)

	// Count ...
	queryCount := db

	if q != "" {
		queryCount = queryCount.Or(
			db.Where("text LIKE ?", "%"+q+"%"),
			db.Where("description LIKE ?", "%"+q+"%"),
		)
	}

	queryICount, err := ctx.Query.SetDatabaseQueryForModel(queryCount, &VocabularyModel{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%+v\n", err),
		}).Error("TermCountReq count error")
	}
	queryCount = queryICount.(*gorm.DB)

	return queryCount.
		Table("terms").
		Count(opts.Count).Error
}
