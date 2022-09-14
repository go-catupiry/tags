package tags

import (
	"net/http"

	"github.com/go-catupiry/catu"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type VocabularyListJSONResponse struct {
	catu.BaseListReponse
	Records *[]VocabularyModel `json:"vocabulary"`
}

type VocabularyCountJSONResponse struct {
	catu.BaseMetaResponse
}

type VocabularyFindOneJSONResponse struct {
	Record *VocabularyModel `json:"vocabulary"`
}

type VocabularyBodyRequest struct {
	Record *VocabularyModel `json:"vocabulary"`
}

type VocabularyTeaserTPL struct {
	Ctx    *catu.RequestContext
	Record *VocabularyModel
}

// Http vocabulary controller | struct with http handlers
type VocabularyController struct {
	App catu.App
}

func (ctl *VocabularyController) Query(c echo.Context) error {
	var err error

	RequestContext := c.(*catu.RequestContext)

	var count int64
	var records []VocabularyModel
	err = VocabularyQueryAndCountReq(&VocabularyQueryOpts{
		Records: &records,
		Count:   &count,
		Limit:   RequestContext.GetLimit(),
		Offset:  RequestContext.GetOffset(),
		C:       c,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Debug("Query Error on find b3-news")
	}

	RequestContext.Pager.Count = count

	logrus.WithFields(logrus.Fields{
		"count":             count,
		"len_records_found": len(records),
	}).Debug("Query count result")

	for i := range records {
		records[i].LoadData()
	}

	resp := VocabularyListJSONResponse{
		Records: &records,
	}

	resp.Meta.Count = count

	return c.JSON(200, &resp)

}

func (ctl *VocabularyController) Create(c echo.Context) error {
	logrus.Debug("VocabularyController.Create running")
	var err error
	ctx := c.(*catu.RequestContext)

	can := ctx.Can("create_vocabulary")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var body VocabularyBodyRequest

	if err := c.Bind(&body); err != nil {
		if _, ok := err.(*echo.HTTPError); ok {
			return err
		}
		return c.NoContent(http.StatusNotFound)
	}

	record := body.Record
	record.ID = 0

	if err := c.Validate(record); err != nil {
		if _, ok := err.(*echo.HTTPError); ok {
			return err
		}
		return err
	}

	logrus.WithFields(logrus.Fields{
		"body": body,
	}).Info("VocabularyController.Create params")

	err = record.Save()
	if err != nil {
		return err
	}

	err = record.LoadData()
	if err != nil {
		return err
	}

	resp := VocabularyFindOneJSONResponse{
		Record: record,
	}

	return c.JSON(http.StatusCreated, &resp)
}

func (ctl *VocabularyController) Count(c echo.Context) error {
	var err error
	RequestContext := c.(*catu.RequestContext)

	var count int64
	err = VocabularyCountReq(&VocabularyQueryOpts{
		Count:  &count,
		Limit:  RequestContext.GetLimit(),
		Offset: RequestContext.GetOffset(),
		C:      c,
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Debug("VocabularyController.Count Error on find records")
	}

	RequestContext.Pager.Count = count

	resp := VocabularyCountJSONResponse{}
	resp.Count = count

	return c.JSON(200, &resp)
}

func (ctl *VocabularyController) FindOne(c echo.Context) error {
	id := c.Param("id")

	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("VocabularyController.FindOne id from params")

	record := VocabularyModel{}
	err := VocabularyFindOne(id, &record)
	if err != nil {
		return err
	}

	if record.ID == 0 {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Debug("VocabularyController.FindOne id record not found")

		return echo.NotFoundHandler(c)
	}

	record.LoadData()

	resp := VocabularyFindOneJSONResponse{
		Record: &record,
	}

	return c.JSON(200, &resp)
}

func (ctl *VocabularyController) Update(c echo.Context) error {
	var err error

	id := c.Param("id")

	RequestContext := c.(*catu.RequestContext)

	logrus.WithFields(logrus.Fields{
		"id":    id,
		"roles": RequestContext.GetAuthenticatedRoles(),
	}).Debug("VocabularyController.Update id from params")

	can := RequestContext.Can("update_vocabulary")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	record := VocabularyModel{}
	err = VocabularyFindOne(id, &record)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Debug("VocabularyController.Update error on find one")
		return errors.Wrap(err, "VocabularyController.Update error on find one")
	}

	record.LoadData()

	body := VocabularyFindOneJSONResponse{Record: &record}

	if err := c.Bind(&body); err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Debug("VocabularyController.Update error on bind")

		if _, ok := err.(*echo.HTTPError); ok {
			return err
		}
		return c.NoContent(http.StatusNotFound)
	}

	err = record.Save()
	if err != nil {
		return err
	}
	resp := VocabularyFindOneJSONResponse{
		Record: &record,
	}

	return c.JSON(http.StatusOK, &resp)
}

func (ctl *VocabularyController) Delete(c echo.Context) error {
	var err error

	id := c.Param("id")

	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("VocabularyController.Delete id from params")

	RequestContext := c.(*catu.RequestContext)

	can := RequestContext.Can("delete_vocabulary")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	record := VocabularyModel{}
	err = VocabularyFindOne(id, &record)
	if err != nil {
		return err
	}

	err = record.Delete()
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (ctl *VocabularyController) FindAllPageHandler(c echo.Context) error {
	panic("TODO!")
}

type VocabularyControllerCfg struct {
	App catu.App
}

func NewVocabularyController(cfg *VocabularyControllerCfg) *VocabularyController {
	ctx := VocabularyController{App: cfg.App}

	return &ctx
}
