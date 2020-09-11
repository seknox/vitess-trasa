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

package preparestmt

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/seknox/trasadbproxy/vitess/go/test/endtoend/cluster"
)

// TestSelect simple select the data without any condition.
func TestSelect(t *testing.T) {
	defer cluster.PanicHandler(t)
	dbo := Connect(t)
	defer dbo.Close()
	selectWhere(t, dbo, "")
}

// TestInsertUpdateDelete validates all insert, update and
// delete method on prepared statements.
func TestInsertUpdateDelete(t *testing.T) {
	defer cluster.PanicHandler(t)
	dbo := Connect(t)
	defer dbo.Close()
	// prepare insert statement
	insertStmt := `insert into ` + tableName + ` values( ?,  ?,  ?,  ?,  ?,  ?,  ?, ?,
		?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?, ?, ?, ?,
		?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?, ?, ?, ?);`

	textValue := fake.FullName()
	largeComment := fake.Paragraph()

	location, _ := time.LoadLocation("Local")
	// inserting multiple rows into test table
	for i := 1; i <= 100; i++ {
		// preparing value for the insert testing
		insertValue := []interface{}{
			i, fmt.Sprint(i) + "21", i * 100,
			127, 1, 32767, 8388607, 2147483647, 2.55, 64.9, 55.5,
			time.Date(2009, 5, 5, 0, 0, 0, 50000, time.UTC),
			time.Date(2009, 5, 5, 0, 0, 0, 50000, location),
			time.Date(2009, 5, 5, 0, 0, 0, 50000, location),
			time.Now(),
			time.Date(2009, 5, 5, 0, 0, 0, 50000, time.UTC),
			1, 1, 1, 1, 1, 1, 1, 1, 1, jsonExample, textValue, largeComment,
			-128, 127, 1, -1,
			-32768, 32767, 1, -1,
			-8388608, 8388607, 1, -1,
			-2147483648, 2147483647, 1, -1,
			-(1 << 63), (1 << 63) - 1, 1, -1,
		}
		exec(t, dbo, insertStmt, insertValue...)

	}
	// validate inserted data count
	testcount(t, dbo, 100)

	// select data with id 1 and validate the data accordingly
	// validate row count
	data := selectWhere(t, dbo, "id = ?", testingID)
	assert.Equal(t, 1, len(data))

	// validate value of msg column in data
	assert.Equal(t, fmt.Sprintf("%d21", testingID), data[0].Msg)

	// Validate a datetime field (without micros)
	//   The 50 microsecs we inserted should have been truncated
	assert.Equal(t, time.Date(2009, 5, 5, 0, 0, 0, 0, location), data[0].DateTime)

	// Validate a datetime field (with micros)
	assert.Equal(t, time.Date(2009, 5, 5, 0, 0, 0, 50000, location), data[0].DateTimeMicros)

	testReplica(t)
	// testing record update
	updateRecord(t, dbo)

	// testing record deletion
	deleteRecord(t, dbo)

	// testing reconnection and deleted data validation
	reconnectAndTest(t)
}

func testReplica(t *testing.T) {
	replicaConn := Connect(t, "")
	require.NotNil(t, replicaConn, "unable to connect")
	_, err := replicaConn.Exec("use @replica")
	require.NoError(t, err)
	tx, err := replicaConn.Begin()
	require.NoError(t, err, "error creating replica transaction")
	data := selectWhereWithTx(t, tx, "id = ?", testingID)
	assert.Equal(t, fmt.Sprintf("%d21", testingID), data[0].Msg)
	require.NoError(t, tx.Commit())
}

// testcount validates inserted rows count with expected count.
func testcount(t *testing.T, dbo *sql.DB, except int) {
	defer cluster.PanicHandler(t)
	r, err := dbo.Query("SELECT count(1) FROM " + tableName)
	require.Nil(t, err)

	r.Next()
	var i int
	err = r.Scan(&i)
	require.Nil(t, err)
	assert.Equal(t, except, i)
}

// TestAutoIncColumns test insertion of row without passing
// the value of auto increment columns (here it is id).
func TestAutoIncColumns(t *testing.T) {
	defer cluster.PanicHandler(t)
	dbo := Connect(t)
	defer dbo.Close()
	// insert a row without id
	insertStmt := "INSERT INTO " + tableName + ` (
		msg,keyspace_id,tinyint_unsigned,bool_signed,smallint_unsigned,
		mediumint_unsigned,int_unsigned,float_unsigned,double_unsigned,
		decimal_unsigned,t_date,t_datetime,t_datetime_micros,t_time,t_timestamp,c8,c16,c24,
		c32,c40,c48,c56,c63,c64,json_col,text_col,data,
		tinyint_min,tinyint_max,tinyint_pos,tinyint_neg,
		smallint_min,smallint_max,smallint_pos,smallint_neg,
		medint_min,medint_max,medint_pos,medint_neg,
		int_min,int_max,int_pos,int_neg,
		bigint_min,bigint_max,bigint_pos,bigint_neg
) VALUES (?,  ?,  ?,  ?,  ?, ?,
		  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?, ?, ?, ?,
		  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?,  ?, ?, ?, ?);`
	insertValue := []interface{}{
		"21", 0,
		127, 1, 32767, 8388607, 2147483647, 2.55, 64.9, 55.5,
		time.Date(2009, 5, 5, 0, 0, 0, 50000, time.UTC),
		time.Date(2009, 5, 5, 0, 0, 0, 50000, time.UTC),
		time.Date(2009, 5, 5, 0, 0, 0, 50000, time.UTC),
		time.Now(),
		time.Date(2009, 5, 5, 0, 0, 0, 50000, time.UTC),
		1, 1, 1, 1, 1, 1, 1, 1, 1, jsonExample, fake.DomainName(), fake.Paragraph(),
		-(1 << 7), (1 << 7) - 1, 1, -1,
		-(1 << 15), (1 << 15) - 1, 1, -1,
		-(1 << 23), (1 << 23) - 1, 1, -1,
		-(1 << 31), (1 << 31) - 1, 1, -1,
		-(1 << 63), (1 << 63) - 1, 1, -1,
	}

	exec(t, dbo, insertStmt, insertValue...)
}

// deleteRecord test deletion operation corresponds to the testingID.
func deleteRecord(t *testing.T, dbo *sql.DB) {
	// delete the record with id 1
	exec(t, dbo, "DELETE FROM "+tableName+" WHERE id = ?;", testingID)

	data := selectWhere(t, dbo, "id = ?", testingID)
	assert.Equal(t, 0, len(data))

}

// updateRecord test update operation corresponds to the testingID.
func updateRecord(t *testing.T, dbo *sql.DB) {
	// update the record with id 1
	updateData := "new data value"
	updateTextCol := "new text col value"
	updateQuery := "update " + tableName + " set data = ? , text_col = ? where id = ?;"

	exec(t, dbo, updateQuery, updateData, updateTextCol, testingID)

	// validate the updated value
	// validate row count
	data := selectWhere(t, dbo, "id = ?", testingID)
	assert.Equal(t, 1, len(data))

	// validate value of msg column in data
	assert.Equal(t, updateData, data[0].Data)
	assert.Equal(t, updateTextCol, data[0].TextCol)

}

// reconnectAndTest creates new connection with database and validate.
func reconnectAndTest(t *testing.T) {
	// reconnect and try to select the record with id 1
	dbo := Connect(t)
	defer dbo.Close()
	data := selectWhere(t, dbo, "id = ?", testingID)
	assert.Equal(t, 0, len(data))

}

// TestColumnParameter query database using column
// parameter.
func TestColumnParameter(t *testing.T) {
	defer cluster.PanicHandler(t)
	dbo := Connect(t)
	defer dbo.Close()

	id := 1000
	parameter1 := "param1"
	message := "TestColumnParameter"
	insertStmt := "INSERT INTO " + tableName + " (id, msg, keyspace_id) VALUES (?, ?, ?);"
	values := []interface{}{
		id,
		message,
		2000,
	}
	exec(t, dbo, insertStmt, values...)

	var param, msg string
	var recID int

	selectStmt := "SELECT COALESCE(?, id), msg FROM " + tableName + " WHERE msg = ? LIMIT ?"

	results1, err := dbo.Query(selectStmt, parameter1, message, 1)
	require.Nil(t, err)
	require.True(t, results1.Next())

	results1.Scan(&param, &msg)
	assert.Equal(t, parameter1, param)
	assert.Equal(t, message, msg)

	results2, err := dbo.Query(selectStmt, nil, message, 1)
	require.Nil(t, err)
	require.True(t, results2.Next())

	results2.Scan(&recID, &msg)
	assert.Equal(t, id, recID)
	assert.Equal(t, message, msg)
}

// TestWrongTableName query database using invalid
// tablename and validate error.
func TestWrongTableName(t *testing.T) {
	defer cluster.PanicHandler(t)
	dbo := Connect(t)
	defer dbo.Close()
	execWithError(t, dbo, []uint16{1146}, "select * from teseting_table;")
}