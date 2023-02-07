package tags

import (
	"net/http"

	"github.com/go-catupiry/catu"
	"github.com/go-catupiry/metatags"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type TermTextsResponse struct {
	catu.BaseListReponse
	Terms []string `json:"term"`
}

type TermListJSONResponse struct {
	catu.BaseListReponse
	Records *[]TermModel `json:"term"`
}

type TermCountJSONResponse struct {
	catu.BaseMetaResponse
}

type TermFindOneJSONResponse struct {
	Record *TermModel `json:"term"`
}

type TermBodyRequest struct {
	Record *TermModel `json:"term"`
}

type TermTeaserTPL struct {
	Ctx    *catu.RequestContext
	Record *TermModel
}

type RelatedRecordTeaserTPL struct {
	Ctx    *catu.RequestContext
	Record interface{}
}

// Http term controller | struct with http handlers
type TermController struct {
	App catu.App
}

func (ctl *TermController) Query(c echo.Context) error {
	var err error

	RequestContext := c.(*catu.RequestContext)

	var count int64
	var records []TermModel
	err = TermQueryAndCountReq(&TermQueryOpts{
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

	resp := TermListJSONResponse{
		Records: &records,
	}

	resp.Meta.Count = count

	return c.JSON(200, &resp)

}

func (ctl *TermController) Create(c echo.Context) error {
	logrus.Debug("TermController.Create running")
	var err error
	ctx := c.(*catu.RequestContext)

	can := ctx.Can("create_term")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var body TermBodyRequest

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
	}).Info("TermController.Create params")

	err = record.Save()
	if err != nil {
		return err
	}

	err = record.LoadData()
	if err != nil {
		return err
	}

	resp := TermFindOneJSONResponse{
		Record: record,
	}

	return c.JSON(http.StatusCreated, &resp)
}

func (ctl *TermController) Count(c echo.Context) error {
	var err error
	RequestContext := c.(*catu.RequestContext)

	var count int64
	err = TermCountReq(&TermQueryOpts{
		Count:  &count,
		Limit:  RequestContext.GetLimit(),
		Offset: RequestContext.GetOffset(),
		C:      c,
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Debug("TermController.Count Error on find records")
	}

	RequestContext.Pager.Count = count

	resp := TermCountJSONResponse{}
	resp.Count = count

	return c.JSON(200, &resp)
}

func (ctl *TermController) FindOne(c echo.Context) error {
	id := c.Param("id")
	vocabulary := c.Param("vocabulary")

	logrus.WithFields(logrus.Fields{
		"id":         id,
		"vocabulary": vocabulary,
	}).Debug("TermController.FindOne id from params")

	record := TermModel{}
	err := TermFindOne(id, &record)
	if err != nil {
		return err
	}

	if record.ID == 0 {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Debug("TermController.FindOne id record not found")

		return &catu.HTTPError{
			Code:    404,
			Message: "not found",
		}
	}

	record.LoadData()

	resp := TermFindOneJSONResponse{
		Record: &record,
	}

	return c.JSON(200, &resp)
}

func (ctl *TermController) Update(c echo.Context) error {
	var err error

	id := c.Param("id")

	RequestContext := c.(*catu.RequestContext)

	logrus.WithFields(logrus.Fields{
		"id":    id,
		"roles": RequestContext.GetAuthenticatedRoles(),
	}).Debug("TermController.Update id from params")

	can := RequestContext.Can("update_term")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	record := TermModel{}
	err = TermFindOne(id, &record)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Debug("TermController.Update error on find one")
		return &catu.HTTPError{
			Code:     404,
			Message:  "not found",
			Internal: errors.Wrap(err, "TermController.Update error on find one"),
		}
	}

	record.LoadData()

	body := TermFindOneJSONResponse{Record: &record}

	if err := c.Bind(&body); err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Debug("TermController.Update error on bind")

		if _, ok := err.(*echo.HTTPError); ok {
			return err
		}
		return c.NoContent(http.StatusNotFound)
	}

	err = record.Save()
	if err != nil {
		return err
	}
	resp := TermFindOneJSONResponse{
		Record: &record,
	}

	return c.JSON(http.StatusOK, &resp)
}

func (ctl *TermController) Delete(c echo.Context) error {
	var err error

	id := c.Param("id")

	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("TermController.Delete id from params")

	RequestContext := c.(*catu.RequestContext)

	can := RequestContext.Can("delete_term")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	record := TermModel{}
	err = TermFindOne(id, &record)
	if err != nil {
		return err
	}

	err = record.Delete()
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (ctl *TermController) FindAllPageHandler(c echo.Context) error {
	panic("TODO!")
}

func (ctl *TermController) FindOnePageHandler(c echo.Context) error {
	var err error
	ctx := c.(*catu.RequestContext)

	switch ctx.GetResponseContentType() {
	case "application/json":
		return ctl.FindOne(c)
	}

	id := c.Param("id")
	vocabulary := c.Param("vocabulary")

	logrus.WithFields(logrus.Fields{
		"id":         id,
		"vocabulary": vocabulary,
	}).Debug("TermController.FindOnePagehandler id from params")

	record := TermModel{}

	err = TermFindOne(id, &record)
	if err != nil {
		return err
	}

	if record.ID == 0 {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Debug("TermController.FindOnePagehandler id record not found")
		return echo.NotFoundHandler(c)
	}

	record.LoadData()

	var count int64
	var records []ModelstermsModel
	err = ModelstermQueryAndCountReq(&ModelstermQueryOpts{
		Records: &records,
		Count:   &count,
		Limit:   ctx.GetLimit(),
		Offset:  ctx.GetOffset(),
		C:       c,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Debug("FindOnePageHandler Error on find term modelsterms")
	}

	// load related record
	var teaserList []string
	var hasRecords bool

	for i := range records {
		teaserHTML, err := records[i].RenderRelatedRecord(ctx, ctl.App)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err.Error(),
			}).Error("FindOnePageHandler.FindOnePagehandler error on render teaser")
		} else {
			teaserList = append(teaserList, teaserHTML.String())
			hasRecords = true
		}
	}

	ctx.Set("hasRecords", hasRecords)
	ctx.Set("records", teaserList)
	ctx.Set("RequestPath", ctx.Request().URL.String())

	ctx.Title = record.Text
	ctx.BodyClass = append(ctx.BodyClass, "body-content-findOne")

	mt := c.Get("metatags").(*metatags.HTMLMetaTags)
	mt.Title = record.Text
	mt.Description = record.Description

	ctx.Pager.Count = count

	return c.Render(http.StatusOK, "taxonomy/term/findOne", &catu.TemplateCTX{
		Ctx:     ctx,
		Record:  &record,
		Records: &records,
	})
}

func (ctl *TermController) TermTexts(c echo.Context) error {
	var err error
	ctx := c.(*catu.RequestContext)

	terms, count, err := FindTermTextsAndCount(ctx)
	if err != nil {
		return errors.Wrap(err, "TermTexts error on FindTermTextsAndCount")
	}

	res := TermTextsResponse{
		Terms: terms,
	}

	res.Meta.Count = count

	return c.JSON(200, res)
}

func (ctl *TermController) TagClound(c echo.Context) error {
	panic("TODO!")
}

type TermControllerCfg struct {
	App catu.App
}

func NewTermController(cfg *TermControllerCfg) *TermController {
	ctx := TermController{App: cfg.App}

	return &ctx
}
