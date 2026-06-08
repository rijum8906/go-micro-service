package db_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	taskdb "github.com/rijum8906/relay/services/task-service/internal/db"
)

func TestCreateProjectAndGetProject(t *testing.T) {
	ctx, q := newTestQueries(t)

	createdBy := uuid.New()

	project, err := q.CreateProject(ctx, taskdb.CreateProjectParams{
		CreatedBy:   createdBy,
		Name:        "Platform",
		Description: "Core workstream",
	})
	if err != nil {
		t.Fatal(err)
	}

	if project.ID == uuid.Nil {
		t.Fatal("expected project id")
	}
	if project.Status != "active" {
		t.Fatalf("expected active status, got %q", project.Status)
	}

	got, err := q.GetProject(ctx, project.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != project.ID {
		t.Fatalf("expected %s, got %s", project.ID, got.ID)
	}
	if got.Name != "Platform" {
		t.Fatalf("unexpected name: %q", got.Name)
	}
}

func TestDeleteProjectHidesProjectFromGet(t *testing.T) {
	ctx, q := newTestQueries(t)

	createdBy := uuid.New()
	deletedBy := uuid.New()

	project, err := q.CreateProject(ctx, taskdb.CreateProjectParams{
		CreatedBy:   createdBy,
		Name:        "Platform",
		Description: "Core workstream",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = q.DeleteProject(ctx, taskdb.DeleteProjectParams{
		ID: project.ID,
		DeletedBy: pgUUID(deletedBy),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = q.GetProject(ctx, project.ID)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
}
