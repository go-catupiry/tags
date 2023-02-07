package tags

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-catupiry/catu"
	"github.com/go-catupiry/catu/helpers"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VocabularyConfiguration struct {
	VocabularyName    string
	CanCreate         bool
	FormFieldMultiple bool
	OnlyLowercase     bool
}

// Create a new VocabularyConfiguration with default settings
func NewVocabularyConfiguration() VocabularyConfiguration {
	return VocabularyConfiguration{}
}

// Vocabulary SQL model
type VocabularyModel struct {
	ID          uint64    `gorm:"primaryKey;column:id;type:int(11);not null" json:"id"`
	Name        string    `gorm:"unique;unique;column:name;type:varchar(255)" json:"name"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	CreatedAt   time.Time `gorm:"column:createdAt;type:datetime;not null" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updatedAt;type:datetime;not null" json:"updatedAt"`
	CreatorID   *uint64   `gorm:"index:creatorId;column:creatorId;type:int(11)" json:"creatorId,omitempty"`
	// Users       User      `gorm:"joinForeignKey:creatorId;foreignKey:id" json:"usersList"` // We.js users table

	LinkPermanent string `gorm:"-" json:"linkPermanent"`
}

// TableName - Set db table name for vocabulary table
func (r *VocabularyModel) TableName() string {
	return "vocabularies"
}

func (r *VocabularyModel) GetIDString() string {
	return strconv.FormatUint(r.ID, 10)
}

func (m *VocabularyModel) ToJSON() string {
	jsonString, _ := json.Marshal(m)
	return string(jsonString)
}

func (m *VocabularyModel) Save() error {
	var err error
	db := catu.GetDefaultDatabaseConnection()

	if m.ID == 0 {
		// create ....
		err = db.Create(&m).Error
		if err != nil {
			return err
		}
	} else {
		// update ...
		err = db.Save(&m).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *VocabularyModel) LoadTeaserData() error {
	r.LoadPath()
	return nil
}

func (r *VocabularyModel) LoadData() error {
	return r.LoadTeaserData()
}

func (r *VocabularyModel) GetPath() string {
	path := ""

	if r.ID != 0 {
		path += "/vocabulary/" + r.GetIDString()
	}

	return path
}

func (r *VocabularyModel) LoadPath() error {
	app := catu.GetApp()
	r.LinkPermanent = app.GetConfiguration().Get("APP_ORIGIN") + r.GetPath()
	return nil
}

type VocabularyQueryOpts struct {
	Records *[]VocabularyModel
	Count   *int64
	Limit   int
	Offset  int
	C       echo.Context
	IsHTML  bool
}

// FindOne - Find one b3news.Content record
func VocabularyFindOne(id string, record *VocabularyModel) error {
	db := catu.GetDefaultDatabaseConnection()

	return db.First(&record, id).Error
}

func (r *VocabularyModel) Delete() error {
	db := catu.GetDefaultDatabaseConnection()
	return db.Unscoped().Delete(&r).Error
}

func VocabularyQueryAndCountReq(opts *VocabularyQueryOpts) error {
	db := catu.GetDefaultDatabaseConnection()

	c := opts.C

	q := c.QueryParam("q")

	query := db

	ctx := c.(*catu.RequestContext)

	queryI, err := ctx.Query.SetDatabaseQueryForModel(query, &VocabularyModel{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%+v\n", err),
		}).Error("QueryAndCountReq error")
	}
	query = queryI.(*gorm.DB)

	if q != "" {
		query = query.Where(
			db.Where("name LIKE ?", "%"+q+"%").Or(db.Where("description LIKE ?", "%"+q+"%")),
		)
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

	return VocabularyCountReq(opts)
}

func VocabularyCountReq(opts *VocabularyQueryOpts) error {
	db := catu.GetDefaultDatabaseConnection()

	c := opts.C

	q := c.QueryParam("q")

	ctx := c.(*catu.RequestContext)

	// Count ...
	queryCount := db

	if q != "" {
		queryCount = queryCount.Or(
			db.Where("name LIKE ?", "%"+q+"%"),
			db.Where("description LIKE ?", "%"+q+"%"),
		)
	}

	queryICount, err := ctx.Query.SetDatabaseQueryForModel(queryCount, &VocabularyModel{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%+v\n", err),
		}).Error("vocabulary count error")
	}
	queryCount = queryICount.(*gorm.DB)

	return queryCount.
		Table("vocabularies").
		Count(opts.Count).Error
}
