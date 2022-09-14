package tags

// func TestUpdateFieldTermById_Category(t *testing.T) {
// 	assert := assert.New(t)

// 	app := GetAppInstance()

// 	var categoryFieldCfg = NewCategoryFieldConfiguration("category", "content", "cat")

// 	// var tagFieldCfg = NewTagFieldConfiguration("tags", "content", "tagis")

// 	t.Run("Should Return error if try to associate with a not existent category", func(t *testing.T) {
// 		modelId, termText := "10", "somethingUnknow"

// 		term, assoc, err := UpdateFieldTermById(modelId, termText, &categoryFieldCfg)
// 		assert.NotNil(err)
// 		assert.Equal("UpdateFieldTermById term dont exists and cant create for this field", err.Error())

// 		assert.Nil(term)
// 		assert.Nil(assoc)
// 	})

// 	t.Run("Should associate model field with a valid category", func(t *testing.T) {
// 		modelId, termText := "2", "Gaming"

// 		savedCat := TermModel{
// 			Text:           termText,
// 			VocabularyName: categoryFieldCfg.VocabularyName,
// 		}

// 		err := savedCat.Save()
// 		assert.Nil(err)

// 		term, assoc, err := UpdateFieldTermById(modelId, termText, &categoryFieldCfg)
// 		assert.Nil(err)

// 		assert.NotNil(term)
// 		assert.NotNil(assoc)

// 		assert.Equal(savedCat.ID, term.ID)
// 		assert.Equal(savedCat.Text, term.Text)
// 		assert.Equal(savedCat.ID, *assoc.TermID)
// 	})

// 	t.Run("Should remove old association and associate the new one", func(t *testing.T) {
// 		modelId, termText, oldTermText := "3", "Health", "Tech"

// 		savedCat := TermModel{
// 			Text:           termText,
// 			VocabularyName: categoryFieldCfg.VocabularyName,
// 		}

// 		err := savedCat.Save()
// 		assert.Nil(err)

// 		savedCat2 := TermModel{
// 			Text:           oldTermText,
// 			VocabularyName: categoryFieldCfg.VocabularyName,
// 		}

// 		err = savedCat2.Save()
// 		assert.Nil(err)

// 		assocSaved := ModelstermsModel{
// 			ModelName:      categoryFieldCfg.ModelName,
// 			ModelID:        3,
// 			Field:          categoryFieldCfg.FieldName,
// 			VocabularyName: categoryFieldCfg.VocabularyName,
// 			TermID:         &savedCat2.ID,
// 		}

// 		err = assocSaved.Save()
// 		assert.Nil(err)

// 		var oldSavedTerm TermModel
// 		err = TermFindOneInRecord(categoryFieldCfg.ModelName, categoryFieldCfg.FieldName, modelId, &oldSavedTerm)
// 		assert.Nil(err)

// 		assert.Equal(oldTermText, oldSavedTerm.Text)
// 		assert.Equal(savedCat2.ID, oldSavedTerm.ID)

// 		term, assoc, err := UpdateFieldTermById(modelId, termText, &categoryFieldCfg)
// 		assert.Nil(err)

// 		assert.NotNil(term)
// 		assert.NotNil(assoc)

// 		assert.Equal(savedCat.ID, term.ID)
// 		assert.Equal(savedCat.Text, term.Text)
// 		assert.Equal(savedCat.ID, *assoc.TermID)
// 	})

// 	t.Cleanup(func() {
// 		db := app.DB
// 		r := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ModelstermsModel{})
// 		if r.Error != nil {
// 			log.Println("Error on delete db modelsTerms", r.Error, r.RowsAffected)
// 		}
// 	})
// }

// func TestUpdateFieldTermById_Term(t *testing.T) {
// 	assert := assert.New(t)
// 	app := GetAppInstance()
// 	var cfg = NewTagFieldConfiguration("tags", "content", "tagis")

// 	t.Run("Should Return create a new term and associate", func(t *testing.T) {
// 		modelId := "11"
// 		terms := []string{"somethingNew1", "meToo2"}

// 		err := UpdateFieldTermsById(modelId, terms, &cfg)
// 		assert.Nil(err)

// 		afterSaveTerms := []TermModel{}
// 		err = TermFindManyInRecord(cfg.VocabularyName, cfg.ModelName, cfg.FieldName, modelId, &afterSaveTerms)
// 		assert.Nil(err)
// 		assert.Equal(2, len(afterSaveTerms))
// 		assert.Equal(terms[0], afterSaveTerms[0].Text)
// 		assert.Equal(terms[1], afterSaveTerms[1].Text)
// 	})

// 	t.Run("Should remove 1 terms and add 2 new ones", func(t *testing.T) {
// 		modelId := "12"
// 		oldTerms := []string{"oldOne1"}
// 		terms := []string{"somethingNew2", "meToo3"}

// 		err := UpdateFieldTermsById(modelId, oldTerms, &cfg)
// 		assert.Nil(err)

// 		afterSave1Terms := []TermModel{}
// 		err = TermFindManyInRecord(cfg.VocabularyName, cfg.ModelName, cfg.FieldName, modelId, &afterSave1Terms)
// 		assert.Nil(err)
// 		assert.Equal(1, len(afterSave1Terms))
// 		assert.Equal(oldTerms[0], afterSave1Terms[0].Text)

// 		err = UpdateFieldTermsById(modelId, terms, &cfg)
// 		assert.Nil(err)

// 		afterSave2Terms := []TermModel{}
// 		err = TermFindManyInRecord(cfg.VocabularyName, cfg.ModelName, cfg.FieldName, modelId, &afterSave2Terms)
// 		assert.Nil(err)
// 		assert.Equal(2, len(afterSave2Terms))
// 		assert.Equal(terms[0], afterSave2Terms[0].Text)
// 		assert.Equal(terms[1], afterSave2Terms[1].Text)
// 	})

// 	t.Run("Should remove 2, keep 2 terms and add 2 new ones", func(t *testing.T) {
// 		modelId := "13"
// 		oldTerms := []string{"oldOne0", "oldOne1", "oldOne2", "oldOne3"}
// 		terms := []string{"oldOne2", "oldOne3", "somethingNew2", "meToo3"}

// 		err := UpdateFieldTermsById(modelId, oldTerms, &cfg)
// 		assert.Nil(err)

// 		afterSave1Terms := []TermModel{}
// 		err = TermFindManyInRecord(cfg.VocabularyName, cfg.ModelName, cfg.FieldName, modelId, &afterSave1Terms)
// 		assert.Nil(err)

// 		assert.Equal(4, len(afterSave1Terms))
// 		assert.Equal(oldTerms[0], afterSave1Terms[0].Text)
// 		assert.Equal(oldTerms[1], afterSave1Terms[1].Text)
// 		assert.Equal(oldTerms[2], afterSave1Terms[2].Text)
// 		assert.Equal(oldTerms[3], afterSave1Terms[3].Text)

// 		err = UpdateFieldTermsById(modelId, terms, &cfg)
// 		assert.Nil(err)

// 		afterSave2Terms := []TermModel{}
// 		err = TermFindManyInRecord(cfg.VocabularyName, cfg.ModelName, cfg.FieldName, modelId, &afterSave2Terms)
// 		assert.Nil(err)

// 		assert.Equal(4, len(afterSave2Terms))
// 		assert.Equal(terms[0], afterSave2Terms[0].Text)
// 		assert.Equal(terms[1], afterSave2Terms[1].Text)
// 		assert.Equal(terms[2], afterSave2Terms[2].Text)
// 		assert.Equal(terms[3], afterSave2Terms[3].Text)
// 	})

// 	t.Cleanup(func() {
// 		db := app.DB
// 		r := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ModelstermsModel{})
// 		if r.Error != nil {
// 			log.Println("Error on delete db modelsTerms", r.Error, r.RowsAffected)
// 		}
// 	})
// }
