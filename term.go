package tags

import (
	"strconv"

	"github.com/go-catupiry/catu"
	"github.com/go-catupiry/catu/helpers"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Field configuration interface implements basic term fields logic
type FieldConfigurationInterface interface {
	CanCreateTerm() bool
	SetCanCreate(v bool) error
	IsFormFieldMultiple() bool
	SetFormFieldMultiple(v bool) error
	GetModelName() string
	SetModelName(name string) error
	GetFieldName() string
	SetFieldName(name string) error
	GetVocabularyName() string
	SetVocabularyName(name string) error
	// Methods changing the DB
	FindOneTerm(modelId string, target *TermModel) error
	FindManyTerm(modelId string, target *[]TermModel) error
	FindOneAssoc(modelId, termId string, target *ModelstermsModel) error
	Add(modelId, termText string) (*TermModel, *ModelstermsModel, error)
	AddMany(modelId string, texts []string) error
	Update(modelId string, termsText []string) error
	RemoveMany(modelId string, terms []string) error
	Clear(modelId string) error
	ClearField(modelId string) error
}

// Term field configuration to associate contents with terms
type FieldConfiguration struct {
	DB               *gorm.DB
	AssociationModel interface{}
	ModelToAssociate interface{}

	VocabularyName    string
	CanCreate         bool
	FormFieldMultiple bool
	OnlyLowercase     bool
	ModelName         string
	FieldName         string
}

func (f *FieldConfiguration) IsFormFieldMultiple() bool {
	return f.FormFieldMultiple
}

func (f *FieldConfiguration) SetFormFieldMultiple(v bool) error {
	f.FormFieldMultiple = v
	return nil
}

func (f *FieldConfiguration) CanCreateTerm() bool {
	return f.CanCreate
}

func (f *FieldConfiguration) SetCanCreate(v bool) error {
	f.CanCreate = v
	return nil
}

func (f *FieldConfiguration) GetModelName() string {
	return f.ModelName
}

func (f *FieldConfiguration) SetModelName(name string) error {
	f.ModelName = name
	return nil
}

func (f *FieldConfiguration) GetFieldName() string {
	return f.FieldName
}

func (f *FieldConfiguration) SetFieldName(name string) error {
	f.FieldName = name
	return nil
}

func (f *FieldConfiguration) GetVocabularyName() string {
	return f.VocabularyName
}

func (f *FieldConfiguration) SetVocabularyName(name string) error {
	f.VocabularyName = name
	return nil
}

func (f *FieldConfiguration) FindOneTerm(modelId string, target *TermModel) error {
	err := f.DB.
		Joins(`INNER JOIN modelsterms AS A on
			A.field = ? AND
			A.modelName = ? AND
			A.modelId = ? AND
			A.termId = terms.id`, f.GetFieldName(), f.GetModelName(), modelId).
		// Where("modelName = ? AND modelId = ?", "modelName", "modelId").
		First(&target).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

func (f *FieldConfiguration) FindManyTerm(modelId string, target *[]TermModel) error {
	err := f.DB.
		Joins(`INNER JOIN modelsterms AS A on
			A.vocabularyName = ? AND
			A.field = ? AND
			A.modelName = ? AND
			A.modelId = ? AND
			A.termId = terms.id`, f.GetVocabularyName(), f.GetFieldName(), f.GetModelName(), modelId).
		Order("'order' ASC").
		// Where("modelName = ? AND modelId = ?", "modelName", "modelId").
		Find(&target).Error
	if err != nil {
		return err
	}

	return nil
}

func (f *FieldConfiguration) FindOneAssoc(modelId, termId string, target *ModelstermsModel) error {
	err := f.DB.
		Where("modelName = ? AND field = ? AND modelId = ? AND termId = ?", f.GetModelName(), f.GetFieldName(), modelId, termId).
		First(target).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

func (f *FieldConfiguration) Add(modelId, termText string) (*TermModel, *ModelstermsModel, error) {
	newTerm := TermModel{}
	err := TermFindOneByText(termText, f.GetVocabularyName(), &newTerm)
	if err != nil {
		return nil, nil, errors.Wrap(err, "FieldConfiguration.AddByText error on find new term text")
	}

	modelIdn, _ := strconv.ParseUint(modelId, 10, 64)

	newAssocRecord, _ := NewModelsterms(f.GetVocabularyName(), f.GetModelName(), f.GetFieldName(), modelIdn, newTerm.ID)
	err = newAssocRecord.Save()
	if err != nil {
		return &newTerm, &newAssocRecord, errors.Wrap(err, "FieldConfiguration.AddByText error on delete old term assoc")
	}

	return &newTerm, &newAssocRecord, nil
}

func (f *FieldConfiguration) AddMany(modelId string, texts []string) error {
	if len(texts) == 0 {
		return nil
	}

	terms := []TermModel{}

	err := TermFindManyByText(texts, f.GetVocabularyName(), &terms)
	if err != nil {
		return err
	}

	termsToCreate := []string{}

	for i := range texts {
		contains := false
		for j := range terms {
			if texts[i] == terms[j].Text {
				contains = true
				break
			}
		}

		if !contains {
			termsToCreate = append(termsToCreate, texts[i])
		}
	}

	if len(termsToCreate) > 0 {
		termsToCreateObj := []TermModel{}

		for i := range termsToCreate {
			t := TermModel{
				Text:           termsToCreate[i],
				VocabularyName: f.GetVocabularyName(),
			}

			termsToCreateObj = append(termsToCreateObj, t)
		}

		err = f.DB.Create(&termsToCreateObj).Error
		if err != nil {
			return errors.Wrap(err, "FieldConfiguration.AddMany error on create terms")
		}

		// assoc terms / refresh after create new ones
		err = TermFindManyByText(texts, f.GetVocabularyName(), &terms)
		if err != nil {
			return err
		}
	}

	modelIdn, _ := strconv.ParseUint(modelId, 10, 64)

	// create assocs
	assocsToCreate := []ModelstermsModel{}
	for i := range texts {
		var orderedTerm *TermModel

		for j := range terms {
			if terms[j].Text == texts[i] {
				orderedTerm = &terms[j]
				break
			}
		}

		if orderedTerm != nil {
			r := ModelstermsModel{
				VocabularyName: f.GetVocabularyName(),
				ModelName:      f.GetModelName(),
				Field:          f.GetFieldName(),
				ModelID:        modelIdn,
				TermID:         &orderedTerm.ID,
				Order:          i,
			}

			assocsToCreate = append(assocsToCreate, r)
		}
	}

	err = f.DB.Create(&assocsToCreate).Error
	if err != nil {
		return errors.Wrap(err, "FieldConfiguration.AddMany error on create assocs")
	}

	return nil
}

func (f *FieldConfiguration) Update(modelId string, termsText []string) error {
	var savedTerms []TermModel
	err := f.FindManyTerm(modelId, &savedTerms)
	if err != nil {
		return errors.Wrap(err, "FieldConfiguration.Update error on get field terms")
	}
	// Is already empty and the new status should be empty, skip:
	if len(termsText) == 0 && len(termsText) == len(savedTerms) {
		return nil
	}

	// filter items to delete
	var itemsToDelete []string
	for i := range savedTerms {
		if !helpers.SliceContains(termsText, savedTerms[i].Text) {
			itemsToDelete = append(itemsToDelete, savedTerms[i].Text)
		}
	}

	// filter items to add
	var itemsToAdd []string
	termsTextLen := len(termsText)
	for i := 0; i < termsTextLen; i++ {
		contains := false
		for j := range savedTerms {
			if savedTerms[j].Text == termsText[i] {
				contains = true
				break
			}
		}

		if !contains {
			itemsToAdd = append(itemsToAdd, termsText[i])
		}
	}

	// delete old items
	err = f.RemoveMany(modelId, itemsToDelete)
	if err != nil {
		return errors.Wrap(err, "UpdateFieldTermsById error on delete terms")
	}

	// create not existent terms and associate
	err = f.AddMany(modelId, itemsToAdd)
	if err != nil {
		return errors.Wrap(err, "UpdateFieldTermsById error on add new assocs")
	}

	return nil
}

func (f *FieldConfiguration) RemoveMany(modelId string, terms []string) error {
	if len(terms) == 0 {
		return nil
	}

	assocs := []ModelstermsModel{}

	termsWithIds := []TermModel{}
	err := f.DB.
		Where("vocabularyName = ? AND text IN ?", f.GetVocabularyName(), terms).
		Select("id").
		Find(&termsWithIds).Error
	if err != nil {
		return err
	}

	ids := []string{}
	for i := range termsWithIds {
		ids = append(ids, termsWithIds[i].GetIDString())
	}

	err = f.DB.
		Where("modelName = ? AND field = ? AND modelId = ? AND termId IN ?", f.GetModelName(), f.GetFieldName(), modelId, ids).
		Select("id AS id").
		Find(&assocs).Error
	if err != nil {
		return err
	}

	if len(assocs) == 0 {
		return nil
	}

	r := f.DB.
		Delete(&assocs)

	if r.Error != nil {
		return r.Error
	}

	return nil
}

// Delete all records (fiels, images, etc) associated with that record
func (f *FieldConfiguration) Clear(modelID string) error {
	return f.DB.Where("modelId = ? AND modelName = ?", modelID, f.GetModelName()).Delete(&f.AssociationModel).Error
}

func (f *FieldConfiguration) ClearField(modelID string) error {
	return f.DB.Where("modelId = ? AND field = ? AND modelName = ?", modelID, f.GetFieldName(), f.GetModelName()).Delete(&f.AssociationModel).Error
}

// Create a new field configuration with default category settings
func NewCategoryFieldConfiguration(vocabularyName, modelName, fieldName string) FieldConfigurationInterface {
	db := catu.GetDefaultDatabaseConnection()

	return &FieldConfiguration{
		DB:                db,
		VocabularyName:    vocabularyName,
		CanCreate:         false,
		FormFieldMultiple: false,
		OnlyLowercase:     false,
		ModelName:         modelName,
		FieldName:         fieldName,
		AssociationModel:  ModelstermsModel{},
		ModelToAssociate:  TermModel{},
	}
}

// Create a new field configuration with default tag settings
func NewTagFieldConfiguration(vocabularyName, modelName, fieldName string) FieldConfigurationInterface {
	db := catu.GetDefaultDatabaseConnection()

	return &FieldConfiguration{
		DB:                db,
		VocabularyName:    vocabularyName,
		CanCreate:         true,
		FormFieldMultiple: true,
		OnlyLowercase:     true,
		ModelName:         modelName,
		FieldName:         fieldName,
		AssociationModel:  ModelstermsModel{},
		ModelToAssociate:  TermModel{},
	}
}
