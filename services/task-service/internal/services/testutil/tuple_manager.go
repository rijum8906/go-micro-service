package testutil

import (
	"context"

	"github.com/openfga/go-sdk/client"
	"github.com/rijum8906/relay/packages/core/apperror"
)

type TupleManager struct {
	Writes  []client.ClientTupleKey
	Deletes []client.ClientTupleKeyWithoutCondition
}

func (m *TupleManager) Write(_ context.Context, writes []client.ClientTupleKey) *apperror.AppError {
	m.Writes = append(m.Writes, writes...)
	return nil
}

func (m *TupleManager) Delete(_ context.Context, deletes []client.ClientTupleKeyWithoutCondition) *apperror.AppError {
	m.Deletes = append(m.Deletes, deletes...)
	return nil
}

func (m *TupleManager) Read(context.Context, client.ClientReadRequest) (*client.ClientReadResponse, *apperror.AppError) {
	panic("unexpected Read call")
}

func (m *TupleManager) Check(context.Context, client.ClientCheckRequest) (*client.ClientCheckResponse, *apperror.AppError) {
	panic("unexpected Check call")
}

func (m *TupleManager) HasWrite(want client.ClientTupleKey) bool {
	for _, got := range m.Writes {
		if got.User == want.User && got.Relation == want.Relation && got.Object == want.Object {
			return true
		}
	}

	return false
}

func (m *TupleManager) HasDelete(want client.ClientTupleKeyWithoutCondition) bool {
	for _, got := range m.Deletes {
		if got.User == want.User && got.Relation == want.Relation && got.Object == want.Object {
			return true
		}
	}

	return false
}
