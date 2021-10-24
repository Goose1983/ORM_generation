//Файл сгенерирован автоматически по functional_role_permissions.go
//ищи генератор в ./gen/orm/ormGenerator.go
package dbaseModels

import "fmt"
import "strings"
import "urms/application"

type FunctionalRolePermissions []FunctionalRolePermission

func (p *FunctionalRolePermission) GetByID(app application.App, ID uint) error {
	sqlStatement := "SELECT * FROM `functional_role_permissions` WHERE id = ?"
	err := app.DB.QueryRow(sqlStatement, ID).Scan(&p.ID, &p.FunctionalRoleID, &p.PermissionID)
	return err
}

func (ps *FunctionalRolePermissions) MapByID() map[uint]FunctionalRolePermission {
	FunctionalRolePermissionsMap := map[uint]FunctionalRolePermission{}
	for i, FunctionalRolePermission := range *ps {
		FunctionalRolePermissionsMap[FunctionalRolePermission.ID] = (*ps)[i]
	}
	return FunctionalRolePermissionsMap
}

func (ps *FunctionalRolePermissions) GetAllByFunctionalRoleID(app application.App, FunctionalRoleID uint) error {
	*ps = (*ps)[:0]
	sqlStatement := "SELECT * FROM `functional_role_permissions` WHERE functional_role_id = ?"
	rows, err := app.DB.Query(sqlStatement, FunctionalRoleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		p := FunctionalRolePermission{}
		err := rows.Scan(&p.ID, &p.FunctionalRoleID, &p.PermissionID)
		if err != nil {
			return err
		}
		*ps = append(*ps, p)
	}
	return nil
}

func (ps *FunctionalRolePermissions) MapByFunctionalRoleID() map[uint]FunctionalRolePermissions {
	FunctionalRolePermissionsMap := map[uint]FunctionalRolePermissions{}
	for i, FunctionalRolePermission := range *ps {
		FunctionalRolePermissionsMap[FunctionalRolePermission.FunctionalRoleID] = append(FunctionalRolePermissionsMap[FunctionalRolePermission.FunctionalRoleID], (*ps)[i])
	}
	return FunctionalRolePermissionsMap
}

func (p *FunctionalRolePermission) Delete(app application.App) error {
	_, err := app.DB.Exec("DELETE FROM `functional_role_permissions` where ID = ?", p.ID)
	return err
}

func (ps *FunctionalRolePermissions) Delete(app application.App) error {
	sqlStatement := "DELETE FROM `functional_role_permissions` where "
	var idsToDelete []string
	for _, FunctionalRolePermission := range *ps {
		idsToDelete = append(idsToDelete, fmt.Sprint("id = ", FunctionalRolePermission.ID))
	}
	sqlStatement += strings.Join(idsToDelete, " OR ")
	_, err := app.DB.Exec(sqlStatement)
	return err
}

func (ps *FunctionalRolePermissions) Create(app application.App) error {
	if len(*ps) == 0 {
		return nil
	}
	sqlStatement := "INSERT INTO `functional_role_permissions` (functional_role_id, permission_id) values "
	var valuesBlocks []string
	for _, p := range *ps {
		values := []string{
			fmt.Sprint(p.FunctionalRoleID), fmt.Sprint(p.PermissionID),
		}

		valuesBlocks = append(valuesBlocks, "("+strings.Join(values, ",")+")")
	}
	sqlStatement += strings.Join(valuesBlocks, ",")
	_, err := app.DB.Exec(sqlStatement)
	return err
}

func (p *FunctionalRolePermission) Update(app application.App) error {
	sqlStatement := "UPDATE `functional_role_permissions` SET functional_role_id = ?, permission_id = ? WHERE id = ?"
	_, err := app.DB.Exec(sqlStatement, p.FunctionalRoleID, p.PermissionID, p.ID)
	return err
}

func (p *FunctionalRolePermission) Create(app application.App) error {
	result, err := app.DB.Exec("INSERT INTO `functional_role_permissions` (functional_role_id, permission_id) VALUES (?, ?)", p.FunctionalRoleID, p.PermissionID)
	id, _ := result.LastInsertId()
	p.ID = uint(id)
	return err
}

func (ps *FunctionalRolePermissions) GetAll(app application.App) error {
	*ps = (*ps)[:0]
	sqlStatement := "SELECT * FROM `functional_role_permissions`"
	rows, err := app.DB.Query(sqlStatement)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		p := FunctionalRolePermission{}
		err := rows.Scan(&p.ID, &p.FunctionalRoleID, &p.PermissionID)
		if err != nil {
			return err
		}
		*ps = append(*ps, p)
	}
	return nil
}
