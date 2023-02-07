package tags

import (
	"bytes"
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

// ModelstermsModel - Stores terms associations with other models
type ModelstermsModel struct {
	ID             uint64    `gorm:"primaryKey;column:id" json:"id"`
	ModelName      string    `gorm:"index:modelsterms_modelName_IDX;index:modelName_modelId;column:modelName;type:varchar(255);not null" json:"modelName"`
	ModelID        uint64    `gorm:"index:modelName_modelId;column:modelId;type:int(11);not null" json:"modelId"`
	Field          string    `gorm:"index:modelsterms_modelName_IDX;column:field;type:varchar(255);not null" json:"field"`
	IsTag          string    `gorm:"column:isTag;type:varchar(255)" json:"isTag"`
	Order          int       `gorm:"column:order;type:tinyint(1);default:0" json:"order"`
	VocabularyName string    `gorm:"column:vocabularyName;type:varchar(255);not null;default:Tags" json:"vocabularyName"`
	CreatedAt      time.Time `gorm:"column:createdAt" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updatedAt" json:"updatedAt"`
	TermID         *uint64   `gorm:"column:termId;type:int(11)" json:"termId"`
}

func NewModelsterms(vocabularyName, modelName, field string, modelId, termId uint64) (ModelstermsModel, error) {
	r := ModelstermsModel{
		VocabularyName: vocabularyName,
		ModelName:      modelName,
		ModelID:        modelId,
		Field:          field,
		TermID:         &termId,
	}
	return r, nil
}

// TableName get sql table name
func (m *ModelstermsModel) TableName() string {
	return "modelsterms"
}

func (r *ModelstermsModel) GetIDString() string {
	return strconv.FormatUint(r.ID, 10)

}

func (r *ModelstermsModel) GetModelIDString() string {
	return strconv.FormatUint(r.ModelID, 10)

}

func (r *ModelstermsModel) ToJSON() string {
	jsonString, _ := json.MarshalIndent(r, "", "  ")
	return string(jsonString)
}

// Save - Create if is new or update
func (m *ModelstermsModel) Save() error {
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

func (r *ModelstermsModel) RenderRelatedRecord(ctx *catu.RequestContext, app catu.App) (bytes.Buffer, error) {
	tPlugin := app.GetPlugin("taxonomy").(*Plugin)
	return tPlugin.RenderRelatedRecord(r, ctx)
}

func (r *ModelstermsModel) Delete() error {
	db := catu.GetDefaultDatabaseConnection()
	return db.Unscoped().Delete(&r).Error
}

type ModelstermQueryOpts struct {
	Records *[]ModelstermsModel
	Count   *int64
	Limit   int
	Offset  int
	C       echo.Context
	IsHTML  bool
}

func ModelstermQueryAndCountReq(opts *ModelstermQueryOpts) error {
	db := catu.GetDefaultDatabaseConnection()

	c := opts.C
	vocabularyName := c.Param("vocabulary")
	termId := c.Param("id")

	query := db

	ctx := c.(*catu.RequestContext)

	queryI, err := ctx.Query.SetDatabaseQueryForModel(query, &VocabularyModel{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%+v\n", err),
		}).Error("ModelstermQueryAndCountReq error")
	}
	query = queryI.(*gorm.DB)

	if termId != "" {
		query = query.Where("termId = ?", termId)
	}

	if vocabularyName != "" {
		query = query.Where("vocabularyName = ?", vocabularyName)
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

	return ModelstermsCountReq(opts)
}

func ModelstermsCountReq(opts *ModelstermQueryOpts) error {
	db := catu.GetDefaultDatabaseConnection()

	c := opts.C

	ctx := c.(*catu.RequestContext)

	// Count ...
	queryCount := db

	queryICount, err := ctx.Query.SetDatabaseQueryForModel(queryCount, &VocabularyModel{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%+v\n", err),
		}).Error("ModelstermsCountReq count error")
	}
	queryCount = queryICount.(*gorm.DB)

	return queryCount.
		Table("modelsterms").
		Count(opts.Count).Error
}
