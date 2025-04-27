package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Machiel/slugify"
	"github.com/dchest/uniuri"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/validate"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/uptrace/bun"
)

// Projects provides all database operations related to projects.
type Projects struct {
	client *Store
}

// AllowedIDs returns a list of project IDs the user is allowed to access.
func (s *Projects) AllowedIDs() []string {
	result := []string{}

	if s.client.principal == nil {
		return result
	}

	for _, p := range s.client.principal.Projects {
		result = append(result, p.ProjectID)
	}

	for _, t := range s.client.principal.Groups {
		for _, p := range t.Group.Projects {
			result = append(result, p.ProjectID)
		}
	}

	return result
}

// List implements the listing of all projects.
func (s *Projects) List(ctx context.Context, params model.ListParams) ([]*model.Project, int64, error) {
	records := make([]*model.Project, 0)

	q := s.client.handle.NewSelect().
		Model(&records)

	if !s.client.principal.Admin {
		q = q.Where(
			"project.id IN (?)",
			bun.In(s.AllowedIDs()),
		)
	}

	if val, ok := s.validSort(params.Sort); ok {
		q = q.Order(strings.Join(
			[]string{
				val,
				sortOrder(params.Order),
			},
			" ",
		))
	}

	if params.Search != "" {
		q = s.client.SearchQuery(q, params.Search)
	}

	counter, err := q.Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	if params.Limit > 0 {
		q = q.Limit(int(params.Limit))
	}

	if params.Offset > 0 {
		q = q.Offset(int(params.Offset))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, int64(counter), err
	}

	return records, int64(counter), nil
}

// Show implements the details for a specific project.
func (s *Projects) Show(ctx context.Context, name string) (*model.Project, error) {
	record := &model.Project{}

	q := s.client.handle.NewSelect().
		Model(record).
		Where("project.id = ? OR project.slug = ?", name, name)

	if err := q.Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrProjectNotFound
		}

		return record, err
	}

	return record, nil
}

// Create implements the create of a new project.
func (s *Projects) Create(ctx context.Context, record *model.Project) (*model.Project, error) {
	if record.Slug == "" {
		record.Slug = s.slugify(
			ctx,
			"slug",
			record.Name,
			"",
		)
	}

	if err := s.validate(ctx, record, false); err != nil {
		return nil, err
	}

	if err := s.client.handle.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().
			Model(record).
			Exec(ctx); err != nil {
			return err
		}

		if _, err := tx.NewInsert().
			Model(&model.UserProject{
				ProjectID: record.ID,
				UserID:    s.client.principal.ID,
				Perm:      model.UserProjectAdminPerm,
			}).
			Exec(ctx); err != nil {
			return err
		}

		if record.Demo {
			credential1 := &model.Credential{
				ProjectID: record.ID,
				Slug:      "none",
				Name:      "None",
				Kind:      "empty",
				Override:  false,
			}

			if err := credential1.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(credential1).
				Exec(ctx); err != nil {
				return err
			}

			credential2 := &model.Credential{
				ProjectID: record.ID,
				Slug:      "vault",
				Name:      "Vault",
				Kind:      "login",
				Override:  false,
				Login: model.CredentialLogin{
					Password: "p455w0rd",
				},
			}

			if err := credential2.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(credential2).
				Exec(ctx); err != nil {
				return err
			}

			repository1 := &model.Repository{
				ProjectID:    record.ID,
				CredentialID: credential1.ID,
				Slug:         "demo",
				Name:         "Demo",
				URL:          "https://github.com/gexec/gexec-demo.git",
				Branch:       "master",
			}

			if err := repository1.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(repository1).
				Exec(ctx); err != nil {
				return err
			}

			inventory1 := &model.Inventory{
				ProjectID:    record.ID,
				CredentialID: credential1.ID,
				Slug:         "customized",
				Name:         "Customized",
				Kind:         "static",
				Content:      `[customized]\nlocalhost ansible_connection=local`,
			}

			if err := inventory1.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(inventory1).
				Exec(ctx); err != nil {
				return err
			}

			inventory2 := &model.Inventory{
				ProjectID:    record.ID,
				RepositoryID: repository1.ID,
				CredentialID: credential1.ID,
				Slug:         "development",
				Name:         "Developmment",
				Kind:         "file",
				Content:      "ansible/environments/development/inventory.yml",
			}

			if err := inventory2.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(inventory2).
				Exec(ctx); err != nil {
				return err
			}

			inventory3 := &model.Inventory{
				ProjectID:    record.ID,
				RepositoryID: repository1.ID,
				CredentialID: credential1.ID,
				Slug:         "staging",
				Name:         "Staging",
				Kind:         "file",
				Content:      "ansible/environments/staging/inventory.yml",
			}

			if err := inventory3.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(inventory3).
				Exec(ctx); err != nil {
				return err
			}

			inventory4 := &model.Inventory{
				ProjectID:    record.ID,
				RepositoryID: repository1.ID,
				CredentialID: credential1.ID,
				Slug:         "production",
				Name:         "Production",
				Kind:         "file",
				Content:      "ansible/environments/production/inventory.yml",
			}

			if err := inventory4.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(inventory4).
				Exec(ctx); err != nil {
				return err
			}

			environment1 := &model.Environment{
				ProjectID: record.ID,
				Slug:      "empty",
				Name:      "Empty",
			}

			if err := environment1.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment1).
				Exec(ctx); err != nil {
				return err
			}

			environment2 := &model.Environment{
				ProjectID: record.ID,
				Slug:      "development",
				Name:      "Development",
			}

			if err := environment2.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment2).
				Exec(ctx); err != nil {
				return err
			}

			environment2Secret1 := &model.EnvironmentSecret{
				EnvironmentID: environment2.ID,
				Kind:          "var",
				Name:          "example_variable1",
				Content:       "s3cr37",
			}

			if err := environment2Secret1.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment2Secret1).
				Exec(ctx); err != nil {
				return err
			}

			environment2Secret2 := &model.EnvironmentSecret{
				EnvironmentID: environment2.ID,
				Kind:          "env",
				Name:          "EXAMPLE_NAME",
				Content:       "Env variable with secret value for DEV",
			}

			if err := environment2Secret2.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment2Secret2).
				Exec(ctx); err != nil {
				return err
			}

			environment2Value1 := &model.EnvironmentValue{
				EnvironmentID: environment2.ID,
				Kind:          "var",
				Name:          "plain_variable",
				Content:       "Example",
			}

			if err := environment2Value1.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment2Value1).
				Exec(ctx); err != nil {
				return err
			}

			environment2Value2 := &model.EnvironmentValue{
				EnvironmentID: environment2.ID,
				Kind:          "env",
				Name:          "SIMPLE_VARIABLE",
				Content:       "This is not a secret on DEV",
			}

			if err := environment2Value2.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment2Value2).
				Exec(ctx); err != nil {
				return err
			}

			environment3 := &model.Environment{
				ProjectID: record.ID,
				Slug:      "staging",
				Name:      "Staging",
			}

			if err := environment3.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment3).
				Exec(ctx); err != nil {
				return err
			}

			environment3Secret1 := &model.EnvironmentSecret{
				EnvironmentID: environment3.ID,
				Kind:          "var",
				Name:          "example_variable1",
				Content:       "s3cr37",
			}

			if err := environment3Secret1.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment3Secret1).
				Exec(ctx); err != nil {
				return err
			}

			environment3Secret2 := &model.EnvironmentSecret{
				EnvironmentID: environment3.ID,
				Kind:          "env",
				Name:          "EXAMPLE_NAME",
				Content:       "Env variable with secret value for STAGE",
			}

			if err := environment3Secret2.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment3Secret2).
				Exec(ctx); err != nil {
				return err
			}

			environment3Value1 := &model.EnvironmentValue{
				EnvironmentID: environment3.ID,
				Kind:          "var",
				Name:          "plain_variable",
				Content:       "Example",
			}

			if err := environment3Value1.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment3Value1).
				Exec(ctx); err != nil {
				return err
			}

			environment3Value2 := &model.EnvironmentValue{
				EnvironmentID: environment3.ID,
				Kind:          "env",
				Name:          "SIMPLE_VARIABLE",
				Content:       "This is not a secret on STAGE",
			}

			if err := environment3Value2.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment3Value2).
				Exec(ctx); err != nil {
				return err
			}

			environment4 := &model.Environment{
				ProjectID: record.ID,
				Slug:      "production",
				Name:      "Production",
			}

			if err := environment4.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment4).
				Exec(ctx); err != nil {
				return err
			}

			environment4Secret1 := &model.EnvironmentSecret{
				EnvironmentID: environment4.ID,
				Kind:          "var",
				Name:          "example_variable1",
				Content:       "s3cr37",
			}

			if err := environment4Secret1.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment4Secret1).
				Exec(ctx); err != nil {
				return err
			}

			environment4Secret2 := &model.EnvironmentSecret{
				EnvironmentID: environment4.ID,
				Kind:          "env",
				Name:          "EXAMPLE_NAME",
				Content:       "Env variable with secret value for PROD",
			}

			if err := environment4Secret2.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment4Secret2).
				Exec(ctx); err != nil {
				return err
			}

			environment4Value1 := &model.EnvironmentValue{
				EnvironmentID: environment4.ID,
				Kind:          "var",
				Name:          "plain_variable",
				Content:       "Example",
			}

			if err := environment4Value1.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment4Value1).
				Exec(ctx); err != nil {
				return err
			}

			environment4Value2 := &model.EnvironmentValue{
				EnvironmentID: environment4.ID,
				Kind:          "env",
				Name:          "SIMPLE_VARIABLE",
				Content:       "This is not a secret on PROD",
			}

			if err := environment4Value2.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(environment4Value2).
				Exec(ctx); err != nil {
				return err
			}

			template1 := &model.Template{
				ProjectID:     record.ID,
				RepositoryID:  repository1.ID,
				InventoryID:   inventory1.ID,
				EnvironmentID: environment1.ID,
				Slug:          "ping-site",
				Name:          "Ping Site",
				Description:   "This template pings the website to provide real world example of using Gexec.",
				Executor:      "ansible",
				Playbook:      "ansible/ping.yml",
			}

			if err := template1.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(template1).
				Exec(ctx); err != nil {
				return err
			}

			template2 := &model.Template{
				ProjectID:     record.ID,
				RepositoryID:  repository1.ID,
				InventoryID:   inventory1.ID,
				EnvironmentID: environment1.ID,
				Slug:          "opentofu-file",
				Name:          "OpenTofu File",
				Description:   "This template simply creates a file in the workspace with OpenTofu.",
				Executor:      "opentofu",
				Playbook:      "terraform/",
			}

			if err := template2.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(template2).
				Exec(ctx); err != nil {
				return err
			}

			template3 := &model.Template{
				ProjectID:     record.ID,
				RepositoryID:  repository1.ID,
				InventoryID:   inventory1.ID,
				EnvironmentID: environment1.ID,
				Slug:          "terraform-file",
				Name:          "Terraform File",
				Description:   "This template simply creates a file in the workspace with Terraform.",
				Executor:      "terraform",
				Playbook:      "terraform/",
			}

			if err := template3.SerializeSecret(s.client.encrypt.Passphrase); err != nil {
				return err
			}

			if _, err := tx.NewInsert().
				Model(template3).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      record.ID,
				ProjectDisplay: record.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeProject,
				Action:         model.EventActionCreate,
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.Show(ctx, record.ID)
}

// Update implements the update of an existing project.
func (s *Projects) Update(ctx context.Context, record *model.Project) (*model.Project, error) {
	if record.Slug == "" {
		record.Slug = s.slugify(
			ctx,
			"slug",
			record.Name,
			record.ID,
		)
	}

	if err := s.validate(ctx, record, true); err != nil {
		return nil, err
	}

	q := s.client.handle.NewUpdate().
		Model(record).
		Where("id = ?", record.ID)

	if _, err := q.Exec(ctx); err != nil {
		return nil, err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      record.ID,
				ProjectDisplay: record.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeProject,
				Action:         model.EventActionUpdate,
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.Show(ctx, record.ID)
}

// Delete implements the deletion of a project.
func (s *Projects) Delete(ctx context.Context, name string) error {
	record, err := s.Show(ctx, name)

	if err != nil {
		return err
	}

	q := s.client.handle.NewDelete().
		Model((*model.Project)(nil)).
		Where("id = ? OR slug = ?", name, name)

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      record.ID,
				ProjectDisplay: record.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeProject,
				Action:         model.EventActionDelete,
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// ListGroups implements the listing of all groups for a project.
func (s *Projects) ListGroups(ctx context.Context, params model.GroupProjectParams) ([]*model.GroupProject, int64, error) {
	records := make([]*model.GroupProject, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("Project").
		Relation("Group").
		Where("project_id = ?", params.ProjectID)

	if val, ok := s.validGroupSort(params.Sort); ok {
		q = q.Order(strings.Join(
			[]string{
				val,
				sortOrder(params.Order),
			},
			" ",
		))
	}

	if params.Search != "" {
		q = s.client.SearchQuery(q, params.Search)
	}

	counter, err := q.Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	if params.Limit > 0 {
		q = q.Limit(int(params.Limit))
	}

	if params.Offset > 0 {
		q = q.Offset(int(params.Offset))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, int64(counter), err
	}

	return records, int64(counter), nil
}

// AttachGroup implements the attachment of a project to a group.
func (s *Projects) AttachGroup(ctx context.Context, params model.GroupProjectParams) error {
	project, err := s.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	group, err := s.client.Groups.Show(ctx, params.GroupID)

	if err != nil {
		return err
	}

	assigned, err := s.isGroupAssigned(ctx, project.ID, group.ID)

	if err != nil {
		return err
	}

	if assigned {
		return ErrAlreadyAssigned
	}

	record := &model.GroupProject{
		ProjectID: project.ID,
		GroupID:   group.ID,
		Perm:      params.Perm,
	}

	if err := s.validatePerm(record.Perm); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(record).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       project.ID,
				ObjectDisplay:  project.Name,
				ObjectType:     model.EventTypeProjectGroup,
				Action:         model.EventActionCreate,
				Attrs: map[string]interface{}{
					"group_id":      group.ID,
					"group_display": group.Name,
					"perm":          params.Perm,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// PermitGroup implements the permission update for a group on a project.
func (s *Projects) PermitGroup(ctx context.Context, params model.GroupProjectParams) error {
	project, err := s.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	group, err := s.client.Groups.Show(ctx, params.GroupID)

	if err != nil {
		return err
	}

	unassigned, err := s.isGroupUnassigned(ctx, project.ID, group.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	q := s.client.handle.NewUpdate().
		Model((*model.GroupProject)(nil)).
		Set("perm = ?", params.Perm).
		Where("project_id = ?", project.ID).
		Where("group_id = ?", group.ID)

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       project.ID,
				ObjectDisplay:  project.Name,
				ObjectType:     model.EventTypeProjectGroup,
				Action:         model.EventActionUpdate,
				Attrs: map[string]interface{}{
					"group_id":      group.ID,
					"group_display": group.Name,
					"perm":          params.Perm,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// DropGroup implements the removal of a project from a group.
func (s *Projects) DropGroup(ctx context.Context, params model.GroupProjectParams) error {
	project, err := s.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	group, err := s.client.Groups.Show(ctx, params.GroupID)

	if err != nil {
		return err
	}

	unassigned, err := s.isGroupUnassigned(ctx, project.ID, group.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	q := s.client.handle.NewDelete().
		Model((*model.GroupProject)(nil)).
		Where("project_id = ?", project.ID).
		Where("group_id = ?", group.ID)

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       project.ID,
				ObjectDisplay:  project.Name,
				ObjectType:     model.EventTypeProjectGroup,
				Action:         model.EventActionDelete,
				Attrs: map[string]interface{}{
					"group_id":      group.ID,
					"group_display": group.Name,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Projects) isGroupAssigned(ctx context.Context, projectID, groupID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.GroupProject)(nil)).
		Where("project_id = ? AND group_id = ?", projectID, groupID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Projects) isGroupUnassigned(ctx context.Context, projectID, groupID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.GroupProject)(nil)).
		Where("project_id = ? AND group_id = ?", projectID, groupID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count < 1, nil
}

// ListUsers implements the listing of all users for a project.
func (s *Projects) ListUsers(ctx context.Context, params model.UserProjectParams) ([]*model.UserProject, int64, error) {
	records := make([]*model.UserProject, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("Project").
		Relation("User").
		Where("project_id = ?", params.ProjectID)

	if val, ok := s.validUserSort(params.Sort); ok {
		q = q.Order(strings.Join(
			[]string{
				val,
				sortOrder(params.Order),
			},
			" ",
		))
	}

	if params.Search != "" {
		q = s.client.SearchQuery(q, params.Search)
	}

	counter, err := q.Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	if params.Limit > 0 {
		q = q.Limit(int(params.Limit))
	}

	if params.Offset > 0 {
		q = q.Offset(int(params.Offset))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, int64(counter), err
	}

	return records, int64(counter), nil
}

// AttachUser implements the attachment of a project to an user.
func (s *Projects) AttachUser(ctx context.Context, params model.UserProjectParams) error {
	project, err := s.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	user, err := s.client.Users.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	assigned, err := s.isUserAssigned(ctx, project.ID, user.ID)

	if err != nil {
		return err
	}

	if assigned {
		return ErrAlreadyAssigned
	}

	record := &model.UserProject{
		ProjectID: project.ID,
		UserID:    user.ID,
		Perm:      params.Perm,
	}

	if err := s.validatePerm(record.Perm); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(record).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       project.ID,
				ObjectDisplay:  project.Name,
				ObjectType:     model.EventTypeProjectUser,
				Action:         model.EventActionCreate,
				Attrs: map[string]interface{}{
					"user_id":      user.ID,
					"user_display": user.Username,
					"perm":         params.Perm,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// PermitUser implements the permission update for an user on a project.
func (s *Projects) PermitUser(ctx context.Context, params model.UserProjectParams) error {
	project, err := s.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	user, err := s.client.Users.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	unassigned, err := s.isUserUnassigned(ctx, project.ID, user.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	q := s.client.handle.NewUpdate().
		Model((*model.UserProject)(nil)).
		Set("perm = ?", params.Perm).
		Where("project_id = ?", project.ID).
		Where("user_id = ?", user.ID)

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       project.ID,
				ObjectDisplay:  project.Name,
				ObjectType:     model.EventTypeProjectUser,
				Action:         model.EventActionUpdate,
				Attrs: map[string]interface{}{
					"user_id":      user.ID,
					"user_display": user.Username,
					"perm":         params.Perm,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// DropUser implements the removal of a project from an user.
func (s *Projects) DropUser(ctx context.Context, params model.UserProjectParams) error {
	project, err := s.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	user, err := s.client.Users.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	unassigned, err := s.isUserUnassigned(ctx, project.ID, user.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	q := s.client.handle.NewDelete().
		Model((*model.UserProject)(nil)).
		Where("project_id = ?", project.ID).
		Where("user_id = ?", user.ID)

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       project.ID,
				ObjectDisplay:  project.Name,
				ObjectType:     model.EventTypeProjectUser,
				Action:         model.EventActionDelete,
				Attrs: map[string]interface{}{
					"user_id":      user.ID,
					"user_display": user.Username,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Projects) isUserAssigned(ctx context.Context, projectID, userID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.UserProject)(nil)).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Projects) isUserUnassigned(ctx context.Context, projectID, userID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.UserProject)(nil)).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count < 1, nil
}

func (s *Projects) validatePerm(perm string) error {
	if err := validation.Validate(
		perm,
		validation.In("user", "admin"),
	); err != nil {
		return validate.Errors{
			Errors: []validate.Error{
				{
					Field: "perm",
					Error: fmt.Errorf("invalid permission value"),
				},
			},
		}
	}

	return nil
}

func (s *Projects) validate(ctx context.Context, record *model.Project, _ bool) error {
	errs := validate.Errors{}

	if err := validation.Validate(
		record.Slug,
		validation.Required,
		validation.Length(3, 255),
		validation.By(s.uniqueValueIsPresent(ctx, "slug", record.ID)),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "slug",
			Error: err,
		})
	}

	if err := validation.Validate(
		record.Name,
		validation.Required,
		validation.Length(3, 255),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "name",
			Error: err,
		})
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (s *Projects) uniqueValueIsPresent(ctx context.Context, key, id string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		q := s.client.handle.NewSelect().
			Model((*model.Project)(nil)).
			Where("? = ?", bun.Ident(key), val)

		if id != "" {
			q = q.Where(
				"id != ?",
				id,
			)
		}

		exists, err := q.Exists(ctx)

		if err != nil {
			return err
		}

		if exists {
			return errors.New("is already taken")
		}

		return nil
	}
}

func (s *Projects) slugify(ctx context.Context, column, value, id string) string {
	var (
		slug string
	)

	for i := 0; true; i++ {
		if i == 0 {
			slug = slugify.Slugify(value)
		} else {
			slug = slugify.Slugify(
				fmt.Sprintf("%s-%s", value, uniuri.NewLen(6)),
			)
		}

		query := s.client.handle.NewSelect().
			Model((*model.Project)(nil)).
			Where("? = ?", bun.Ident(column), slug)

		if id != "" {
			query = query.Where(
				"id != ?",
				id,
			)
		}

		if count, err := query.Count(
			ctx,
		); err == nil && count == 0 {
			break
		}
	}

	return slug
}

func (s *Projects) validSort(val string) (string, bool) {
	if val == "" {
		return "project.name", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"slug":    "project.slug",
		"name":    "project.name",
		"created": "project.created_at",
		"updated": "project.updated_at",
	} {
		if val == key {
			return name, true
		}
	}

	return "project.name", true
}

func (s *Projects) validGroupSort(val string) (string, bool) {
	if val == "" {
		return "group.name", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"slug": "group.slug",
		"name": "group.name",
	} {
		if val == key {
			return name, true
		}
	}

	return "group.name", true
}

func (s *Projects) validUserSort(val string) (string, bool) {
	if val == "" {
		return "user.username", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"username": "user.username",
		"email":    "user.email",
		"fullname": "user.fullname",
		"admin":    "user.admin",
		"active":   "user.active",
	} {
		if val == key {
			return name, true
		}
	}

	return "user.username", true
}
