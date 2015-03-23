// Copyright 2015 CoreStore Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Generates code
package main

import (
	"fmt"
	"io/ioutil"

	"github.com/corestoreio/csfw/eav"
	"github.com/corestoreio/csfw/storage/dbr"
	"github.com/corestoreio/csfw/tools"
	"github.com/juju/errgo"
)

func materializeEntityType(ctx *context) {
	defer ctx.wg.Done()
	type dataContainer struct {
		ETypeData     eav.EntityTypeSlice
		EntityTypeMap tools.EntityTypeMap
		Package, Tick string
	}

	etData, err := getEntityTypeData(ctx.dbrConn.NewSession(nil))
	tools.LogFatal(err)

	tplData := &dataContainer{
		ETypeData:     etData,
		EntityTypeMap: tools.ConfigEntityType,
		Package:       tools.ConfigEntityTypeMaterialization.Package,
		Tick:          "`",
	}

	formatted, err := tools.GenerateCode(tools.ConfigEntityTypeMaterialization.Package, tplEav, tplData)
	if err != nil {
		fmt.Printf("\n%s\n", formatted)
		tools.LogFatal(err)
	}

	ioutil.WriteFile(tools.ConfigEntityTypeMaterialization.OutputFile, formatted, 0600)
}

func getEntityTypeData(dbrSess *dbr.Session) (etc eav.EntityTypeSlice, err error) {

	s, err := eav.GetTableStructure(eav.TableEntityType)
	if err != nil {
		return nil, errgo.Mask(err)
	}

	_, err = dbrSess.
		Select(s.AllColumnAliasQuote(s.Name)...).
		From(s.Name).
		Where("entity_type_code IN ?", tools.ConfigEntityType.Keys()).
		LoadStructs(&etc)
	if err != nil {
		return nil, errgo.Mask(err)
	}

	for typeCode, mapData := range tools.ConfigEntityType {
		// map the fields from the config struct to the data retrieved from the database.
		et, err := etc.GetByCode(typeCode)
		tools.LogFatal(err)
		et.EntityModel = mapData.EntityModel
		et.AttributeModel.String = mapData.AttributeModel
		et.EntityTable.String = mapData.EntityTable
		et.IncrementModel.String = mapData.IncrementModel
		et.AdditionalAttributeTable.String = mapData.AdditionalAttributeTable
		et.EntityAttributeCollection.String = mapData.EntityAttributeCollection
	}

	return etc, nil
}