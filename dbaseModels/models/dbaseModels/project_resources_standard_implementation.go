//Файл сгенерирован автоматически по project_resources.go
//ищи генератор в ./gen/orm/ormGenerator.go
package dbaseModels

import "fmt"
import "strings"
import "urms/application"

type ProjectResources []ProjectResource

func (p *ProjectResource) GetByID(app application.App, ID uint) error {
	sqlStatement := "SELECT * FROM `project_resources` WHERE id = ?"
	err := app.DB.QueryRow(sqlStatement, ID).Scan(&p.ID, &p.Pkey, &p.ResourceReference, &p.ResourceTypeID, &p.ResourceOwnerProject)
	return err
}

func (ps *ProjectResources) MapByID() map[uint]ProjectResource {
	ProjectResourcesMap := map[uint]ProjectResource{}
	for i, ProjectResource := range *ps {
		ProjectResourcesMap[ProjectResource.ID] = (*ps)[i]
	}
	return ProjectResourcesMap
}

func (ps *ProjectResources) GetAllByPkey(app application.App, Pkey string) error {
	*ps = (*ps)[:0]
	sqlStatement := "SELECT * FROM `project_resources` WHERE pkey = ?"
	rows, err := app.DB.Query(sqlStatement, Pkey)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		p := ProjectResource{}
		err := rows.Scan(&p.ID, &p.Pkey, &p.ResourceReference, &p.ResourceTypeID, &p.ResourceOwnerProject)
		if err != nil {
			return err
		}
		*ps = append(*ps, p)
	}
	return nil
}

func (ps *ProjectResources) MapByPkey() map[string]ProjectResources {
	ProjectResourcesMap := map[string]ProjectResources{}
	for i, ProjectResource := range *ps {
		ProjectResourcesMap[ProjectResource.Pkey] = append(ProjectResourcesMap[ProjectResource.Pkey], (*ps)[i])
	}
	return ProjectResourcesMap
}

func (p *ProjectResource) Delete(app application.App) error {
	_, err := app.DB.Exec("DELETE FROM `project_resources` where ID = ?", p.ID)
	return err
}

func (ps *ProjectResources) Delete(app application.App) error {
	sqlStatement := "DELETE FROM `project_resources` where "
	var idsToDelete []string
	for _, ProjectResource := range *ps {
		idsToDelete = append(idsToDelete, fmt.Sprint("id = ", ProjectResource.ID))
	}
	sqlStatement += strings.Join(idsToDelete, " OR ")
	_, err := app.DB.Exec(sqlStatement)
	return err
}

func (ps *ProjectResources) Create(app application.App) error {
	if len(*ps) == 0 {
		return nil
	}
	sqlStatement := "INSERT INTO `project_resources` (pkey, resource_reference, resource_type_id, resource_owner_project) values "
	var valuesBlocks []string
	for _, p := range *ps {
		values := []string{
			"\"" + p.Pkey + "\"", "\"" + p.ResourceReference + "\"", fmt.Sprint(p.ResourceTypeID), "\"" + p.ResourceOwnerProject + "\"",
		}

		valuesBlocks = append(valuesBlocks, "("+strings.Join(values, ",")+")")
	}
	sqlStatement += strings.Join(valuesBlocks, ",")
	_, err := app.DB.Exec(sqlStatement)
	return err
}

func (p *ProjectResource) Update(app application.App) error {
	sqlStatement := "UPDATE `project_resources` SET pkey = ?, resource_reference = ?, resource_type_id = ?, resource_owner_project = ? WHERE id = ?"
	_, err := app.DB.Exec(sqlStatement, p.Pkey, p.ResourceReference, p.ResourceTypeID, p.ResourceOwnerProject, p.ID)
	return err
}

func (p *ProjectResource) Create(app application.App) error {
	result, err := app.DB.Exec("INSERT INTO `project_resources` (pkey, resource_reference, resource_type_id, resource_owner_project) VALUES (?, ?, ?, ?)", p.Pkey, p.ResourceReference, p.ResourceTypeID, p.ResourceOwnerProject)
	id, _ := result.LastInsertId()
	p.ID = uint(id)
	return err
}

func (ps *ProjectResources) GetAll(app application.App) error {
	*ps = (*ps)[:0]
	sqlStatement := "SELECT * FROM `project_resources`"
	rows, err := app.DB.Query(sqlStatement)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		p := ProjectResource{}
		err := rows.Scan(&p.ID, &p.Pkey, &p.ResourceReference, &p.ResourceTypeID, &p.ResourceOwnerProject)
		if err != nil {
			return err
		}
		*ps = append(*ps, p)
	}
	return nil
}
