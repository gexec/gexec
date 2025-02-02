package model

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

// EventAction defines a custom type for event actions.
type EventAction string

const (
	// EventActionCreate defines the action create.
	EventActionCreate EventAction = "create"

	// EventActionUpdate defines the action update.
	EventActionUpdate EventAction = "update"

	// EventActionDelete defines the action delete.
	EventActionDelete EventAction = "delete"
)

// EventType defines a custom type for event types.
type EventType string

const (
	// EventTypeUser defines event type for users.
	EventTypeUser EventType = "user"

	// EventTypeUserGroup defines event type for user groups.
	EventTypeUserGroup EventType = "user_group"

	// EventTypeUserProject defines event type for user projects.
	EventTypeUserProject EventType = "user_project"

	// EventTypeGroup defines event type for groups.
	EventTypeGroup EventType = "group"

	// EventTypeGroupUser defines event type for group users.
	EventTypeGroupUser EventType = "group_user"

	// EventTypeGroupProject defines event type for group projects.
	EventTypeGroupProject EventType = "group_project"

	// EventTypeProject defines event type for projects.
	EventTypeProject EventType = "project"

	// EventTypeProjectUser defines event type for project users.
	EventTypeProjectUser EventType = "project_user"

	// EventTypeProjectGroup defines event type for project groups.
	EventTypeProjectGroup EventType = "project_group"

	// EventTypeRunner defines event type for runners.
	EventTypeRunner EventType = "runner"

	// EventTypeCredential defines event type for credentials.
	EventTypeCredential EventType = "credential"

	// EventTypeRepository defines event type for repositories.
	EventTypeRepository EventType = "repository"

	// EventTypeInventory defines event type for inventories.
	EventTypeInventory EventType = "inventory"

	// EventTypeEnvironment defines event type for environments.
	EventTypeEnvironment EventType = "environment"

	// EventTypeEnvironmentSecret defines event type for environment secrets.
	EventTypeEnvironmentSecret EventType = "environment_secret"

	// EventTypeEnvironmentValue defines event type for environment vaules.
	EventTypeEnvironmentValue EventType = "environment_value"

	// EventTypeTemplate defines event type for templates.
	EventTypeTemplate EventType = "template"

	// EventTypeTemplateSurvey defines event type for template surveys.
	EventTypeTemplateSurvey EventType = "template_survey"

	// EventTypeTemplateVault defines event type for template vaults.
	EventTypeTemplateVault EventType = "template_vault"

	// EventTypeSchedule defines event type for schedules.
	EventTypeSchedule EventType = "schedule"

	// EventTypeExecution defines event type for executions.
	EventTypeExecution EventType = "execution"

	// EventTypeOutput defines event type for outputs.
	EventTypeOutput EventType = "output"
)

var (
	_ bun.BeforeAppendModelHook = (*Event)(nil)
	_ bun.AfterScanRowHook      = (*Event)(nil)
)

// Event defines the model for events table.
type Event struct {
	bun.BaseModel `bun:"table:events"`

	ID             string                 `bun:",pk,type:varchar(20)"`
	UserID         string                 `bun:"type:varchar(20)"`
	UserHandle     string                 `bun:"type:varchar(255)"`
	UserDisplay    string                 `bun:"type:varchar(255)"`
	ProjectID      string                 `bun:"type:varchar(20)"`
	ProjectDisplay string                 `bun:"type:varchar(255)"`
	ObjectID       string                 `bun:"type:varchar(20)"`
	ObjectDisplay  string                 `bun:"type:varchar(255)"`
	ObjectType     EventType              `bun:"type:varchar(255)"`
	Action         EventAction            `bun:"type:varchar(255)"`
	Attrs          map[string]interface{} `bun:"-"`
	RawAttrs       string                 `bun:"attrs,nullzero,type:text"`
	CreatedAt      time.Time              `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *Event) BeforeAppendModel(_ context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if m.ID == "" {
			m.ID = strings.ToLower(uniuri.NewLen(uniuri.UUIDLen))
		}

		m.CreatedAt = time.Now()
	case *bun.UpdateQuery:
		if m.ID == "" {
			m.ID = strings.ToLower(uniuri.NewLen(uniuri.UUIDLen))
		}
	}

	if m.Attrs != nil {
		result, err := json.Marshal(m.Attrs)

		if err != nil {
			return err
		}

		m.RawAttrs = string(result)
	}

	return nil
}

// AfterScanRow implements the bun hook interface.
func (m *Event) AfterScanRow(_ context.Context) error {
	if m.RawAttrs != "" {
		result := make(map[string]interface{})

		if err := json.Unmarshal([]byte(m.RawAttrs), &result); err != nil {
			return err
		}

		m.Attrs = result
	}

	return nil
}

// SerializeSecret ensures to encrypt all related secrets stored on the database.
func (m *Event) SerializeSecret(_ string) error {
	return nil
}

// DeserializeSecret ensures to decrypt all related secrets stored on the database.
func (m *Event) DeserializeSecret(_ string) error {
	return nil
}

// PrepareEvent prefills the event model with principal information if available.
func PrepareEvent(principal *User, event *Event) *Event {
	if principal != nil {
		event.UserID = principal.ID
		event.UserHandle = principal.Username
		event.UserDisplay = principal.Fullname
	}

	return event
}
