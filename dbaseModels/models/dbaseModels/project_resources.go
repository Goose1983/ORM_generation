package dbaseModels

import "urms/application"

// gen:
type ProjectResource struct {
	ID                   uint   `gen:"uindex"`
	Pkey                 string `gen:"index"`
	ResourceReference    string
	ResourceTypeID       uint
	ResourceOwnerProject string
}

func (ps *ProjectResources) GetAllByPkeyPermission(app application.App, Pkey string, Permission uint) error {
	*ps = (*ps)[:0]
	sqlStatement := "SELECT project_resources.* FROM project_resources inner join permissions p on project_resources.resource_type_id = p.resource_type_id WHERE pkey = ? and p.id = ?"
	rows, err := app.DB.Query(sqlStatement, Pkey, Permission)
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

func (ps *ProjectResources) GetAllByTypeOwnerProject(app application.App, ResourceType uint, Pkey string) error {
	*ps = (*ps)[:0]
	sqlStatement := "SELECT * FROM `project_resources` WHERE  resource_type_id = ? AND resource_owner_project = ?"
	rows, err := app.DB.Query(sqlStatement, ResourceType, Pkey)
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
