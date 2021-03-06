// Copyright 2015, Cyrill @ Schumacher.fm and the CoreStore contributors
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

package directory_test

import (
	"testing"

	"github.com/corestoreio/csfw/config"
	"github.com/corestoreio/csfw/directory"
	"github.com/corestoreio/csfw/utils/log"
)

func init() {
	log.Set(log.NewStdLogger())
	log.SetLevel(log.StdLevelDebug)
}

func TestSourceCurrencyAll(t *testing.T) {

	r := config.NewMockReader(
		config.MockString(func(path string) string {
			t.Log(path)
			switch path {
			case config.MockPathScopeStore(1, directory.PathDefaultLocale):
				return "de_CH"
			}
			return "Not Found"
		}),
	)

	s := config.ScopeID(1)

	sca := directory.NewSourceCurrencyAll(config.ModelConstructor{
		ConfigReader: r,
		Scope:        s,
	})

	t.Logf("\n%#v\n", sca.Options())

}
