/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package planbuilder

import (
	"github.com/seknox/trasadbproxy/vitess/go/vt/key"
	vtrpcpb "github.com/seknox/trasadbproxy/vitess/go/vt/proto/vtrpc"
	"github.com/seknox/trasadbproxy/vitess/go/vt/sqlparser"
	"github.com/seknox/trasadbproxy/vitess/go/vt/vterrors"
	"github.com/seknox/trasadbproxy/vitess/go/vt/vtgate/engine"
)

func buildPlanForBypass(stmt sqlparser.Statement, vschema ContextVSchema) (engine.Primitive, error) {
	switch vschema.Destination().(type) {
	case key.DestinationExactKeyRange:
		if _, ok := stmt.(*sqlparser.Insert); ok {
			return nil, vterrors.Errorf(vtrpcpb.Code_INVALID_ARGUMENT, "range queries not supported for inserts: %s", vschema.TargetString())
		}
	}

	keyspace, err := vschema.DefaultKeyspace()
	if err != nil {
		return nil, err
	}
	return &engine.Send{
		Keyspace:          keyspace,
		TargetDestination: vschema.Destination(),
		Query:             sqlparser.String(stmt),
		IsDML:             sqlparser.IsDMLStatement(stmt),
		SingleShardOnly:   false,
	}, nil
}