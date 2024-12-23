package scim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	"github.com/genexec/genexec/pkg/config"
	"github.com/genexec/genexec/pkg/model"
	"github.com/genexec/genexec/pkg/store"
	"github.com/rs/zerolog/log"
)

func TestGroupsCreate(t *testing.T) {
	dbStore := store.MustStore(config.Database{
		Driver: "sqlite3",
		Name:   "file::memory:?cache=share",
	})

	if _, err := dbStore.Open(); err != nil {
		panic(err)
	}

	defer dbStore.Close()

	if _, err := dbStore.Migrate(context.Background()); err != nil {
		panic(err)
	}

	handler := &groupHandlers{
		config: config.Scim{},
		store:  dbStore.Handle(),
		logger: log.Logger,
	}

	attributes := scim.ResourceAttributes{
		"externalId":  "group-ext-123",
		"displayName": "Test Group",
	}

	body, _ := json.Marshal(attributes)
	req, _ := http.NewRequest("POST", "/scim/v2/Groups", bytes.NewBuffer(body))

	resource, err := handler.Create(req, attributes)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resource.ExternalID != optional.NewString("group-ext-123") {
		t.Errorf("Expected external ID 'group-ext-123', got %v", resource.ExternalID)
	}

	if resource.Attributes["displayName"] != "Test Group" {
		t.Errorf("Expected group displayName 'Test Group', got %v", resource.Attributes["displayName"])
	}
}

func TestUpdateGroup(t *testing.T) {
	dbStore := store.MustStore(config.Database{
		Driver: "sqlite3",
		Name:   "file::memory:?cache=share",
	})

	if _, err := dbStore.Open(); err != nil {
		panic(err)
	}

	defer dbStore.Close()

	if _, err := dbStore.Migrate(context.Background()); err != nil {
		panic(err)
	}

	expected := &model.Team{
		Name: "Group 1",
		Auths: []*model.TeamAuth{
			{
				Provider: "scim",
				Ref:      "group-ext-123",
				Name:     "Group 1",
			},
			{
				Provider: "github",
				Ref:      "group-ext-321",
				Name:     "Group 1",
			},
		},
	}

	// if err := dbStore.Handle().Create(
	// 	expected,
	// ).Error; err != nil {
	// 	panic(err)
	// }

	handler := &groupHandlers{
		config: config.Scim{},
		store:  dbStore.Handle(),
		logger: log.Logger,
	}

	attributes := scim.ResourceAttributes{
		"externalId":  "group-ext-123",
		"displayName": "Group 1",
	}

	jsonBody, _ := json.Marshal(attributes)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/scim/v2/Groups/%s", expected.ID), bytes.NewBuffer(jsonBody))

	resource, err := handler.Replace(
		req,
		expected.ID,
		attributes,
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resource.Attributes["displayName"] != "Group 1" {
		t.Errorf("Expected group displayName 'Group 1', got %v", resource.Attributes["displayName"])
	}
}
