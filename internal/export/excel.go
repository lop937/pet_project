package export

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SlavaNovikov/grif/internal/ds"
	"github.com/SlavaNovikov/grif/internal/product"
	"github.com/SlavaNovikov/grif/internal/project"
	"github.com/SlavaNovikov/grif/internal/system"
	"github.com/SlavaNovikov/grif/internal/user"
	"github.com/SlavaNovikov/grif/pkg/database"
	"github.com/SlavaNovikov/grif/pkg/utils"
	"github.com/xuri/excelize/v2"
)

// ExportWorks выгружает работы за период в .xlsx
func ExportWorks(w http.ResponseWriter, r *http.Request, store ds.Storage) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	var items []struct {
		ID          ds.ID   `db:"id"`
		Date        string  `db:"date"`
		ProjectID   ds.ID   `db:"project_id"`
		SystemID    ds.ID   `db:"system_id"`
		ProductID   ds.ID   `db:"product_id"`
		ServiceType string  `db:"service_type"`
		UserID      ds.ID   `db:"user_id"`
		Quantity    float64 `db:"quantity"`
		Price       float64 `db:"price"`
	}

	filter := database.SqlFilter{
		Table:   ds.WorksRoute.String(),
		Fields:  "id,date,project_id,system_id,product_id,service_type,user_id,quantity,price",
		OrderBy: "date DESC, id DESC",
	}

	if from != "" || to != "" {
		var where string
		var args []any
		if from != "" {
			where = "date >= ?"
			args = append(args, from)
		}
		if to != "" {
			if where != "" {
				where += " AND "
			}
			where += "date <= ?"
			args = append(args, to)
		}
		filter.Where = where
		filter.Args = args
	}

	if err := store.Select(r.Context(), &items, filter); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	usersMap, _ := user.GetUsersMap(r.Context(), store)
	productsMap, _ := product.GetMap(r.Context(), store)
	projectsMap, _ := project.GetMap(r.Context(), store)
	systemsMap, _ := system.GetMap(r.Context(), store)

	f := excelize.NewFile()
	sheet := "Учёт работ"
	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{"Дата", "Проект", "Система", "Товар", "Услуга", "Исполнитель", "Кол-во", "Цена", "Сумма"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for i, item := range items {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.Date)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), projectsMap[item.ProjectID].Name)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), systemsMap[item.SystemID].Name)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), productsMap[item.ProductID].Name)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), item.ServiceType)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), usersMap[item.UserID].FIO)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), item.Quantity)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), item.Price)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), item.Quantity*item.Price)
	}

	f.SetColWidth(sheet, "A", "A", 12)
	f.SetColWidth(sheet, "B", "D", 25)
	f.SetColWidth(sheet, "E", "F", 20)
	f.SetColWidth(sheet, "G", "I", 10)

	filename := fmt.Sprintf("works_%s.xlsx", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if err := f.Write(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ExportProject выгружает один проект и его системы
func ExportProject(w http.ResponseWriter, r *http.Request, store ds.Storage, id ds.ID) {
	proj, err := project.GetByID(r.Context(), store, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var systems []system.System
	_ = store.Select(r.Context(), &systems, database.SqlFilter{
		Table:  ds.SystemRoute.String(),
		Fields: "id," + utils.ExtractFields("db", system.System{}),
		Where:  "project_id = ?",
		Args:   []any{id},
	})

	f := excelize.NewFile()
	sheet := "Проект"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	f.SetCellValue(sheet, "A1", "Проект")
	f.SetCellValue(sheet, "B1", proj.Name)
	f.SetCellValue(sheet, "A2", "Дата")
	f.SetCellValue(sheet, "B2", proj.Date)
	f.SetCellValue(sheet, "A3", "Заказчик")
	f.SetCellValue(sheet, "B3", proj.Customer)
	f.SetCellValue(sheet, "A4", "Статус")
	f.SetCellValue(sheet, "B4", proj.Status)

	f.SetCellValue(sheet, "A6", "Системы")
	for i, sys := range systems {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+7), sys.Name)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+7), sys.Status)
	}

	filename := fmt.Sprintf("project_%s.xlsx", proj.Name)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(w)
}

// ExportSystems выгружает системы проекта
func ExportSystems(w http.ResponseWriter, r *http.Request, store ds.Storage, projectID ds.ID) {
	var systems []system.System
	_ = store.Select(r.Context(), &systems, database.SqlFilter{
		Table:  ds.SystemRoute.String(),
		Fields: "id," + utils.ExtractFields("db", system.System{}),
		Where:  "project_id = ?",
		Args:   []any{projectID},
	})

	f := excelize.NewFile()
	sheet := "Системы"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	f.SetCellValue(sheet, "A1", "Название")
	f.SetCellValue(sheet, "B1", "Статус")

	for i, sys := range systems {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+2), sys.Name)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+2), sys.Status)
	}

	filename := fmt.Sprintf("systems_%d.xlsx", projectID)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(w)
}

// ExportFullProject выгружает проект, все его системы и все работы
func ExportFullProject(w http.ResponseWriter, r *http.Request, store ds.Storage, id ds.ID) {
	proj, err := project.GetByID(r.Context(), store, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var systems []system.System
	_ = store.Select(r.Context(), &systems, database.SqlFilter{
		Table:  ds.SystemRoute.String(),
		Fields: "id," + utils.ExtractFields("db", system.System{}),
		Where:  "project_id = ?",
		Args:   []any{id},
	})

	var works []struct {
		ID          ds.ID   `db:"id"`
		Date        string  `db:"date"`
		SystemID    ds.ID   `db:"system_id"`
		ProductID   ds.ID   `db:"product_id"`
		ServiceType string  `db:"service_type"`
		UserID      ds.ID   `db:"user_id"`
		Quantity    float64 `db:"quantity"`
		Price       float64 `db:"price"`
	}
	_ = store.Select(r.Context(), &works, database.SqlFilter{
		Table:   ds.WorksRoute.String(),
		Fields:  "id,date,system_id,product_id,service_type,user_id,quantity,price",
		Where:   "project_id = ?",
		OrderBy: "date DESC",
		Args:    []any{id},
	})

	usersMap, _ := user.GetUsersMap(r.Context(), store)
	productsMap, _ := product.GetMap(r.Context(), store)
	systemsMap, _ := system.GetMap(r.Context(), store)

	f := excelize.NewFile()
	sheet := "Проект"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	f.SetCellValue(sheet, "A1", "Проект:")
	f.SetCellValue(sheet, "B1", proj.Name)
	f.SetCellValue(sheet, "A2", "Дата:")
	f.SetCellValue(sheet, "B2", proj.Date)
	f.SetCellValue(sheet, "A3", "Заказчик:")
	f.SetCellValue(sheet, "B3", proj.Customer)
	f.SetCellValue(sheet, "A4", "Статус:")
	f.SetCellValue(sheet, "B4", proj.Status)

	f.SetCellValue(sheet, "A6", "Системы:")
	for i, sys := range systems {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+7), sys.Name)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+7), sys.Status)
	}

	startRow := 7 + len(systems) + 2

	f.SetCellValue(sheet, fmt.Sprintf("A%d", startRow), "Дата")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", startRow), "Система")
	f.SetCellValue(sheet, fmt.Sprintf("C%d", startRow), "Товар")
	f.SetCellValue(sheet, fmt.Sprintf("D%d", startRow), "Услуга")
	f.SetCellValue(sheet, fmt.Sprintf("E%d", startRow), "Исполнитель")
	f.SetCellValue(sheet, fmt.Sprintf("F%d", startRow), "Кол-во")
	f.SetCellValue(sheet, fmt.Sprintf("G%d", startRow), "Цена")
	f.SetCellValue(sheet, fmt.Sprintf("H%d", startRow), "Сумма")

	for i, w := range works {
		row := startRow + 1 + i
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), w.Date)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), systemsMap[w.SystemID].Name)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), productsMap[w.ProductID].Name)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), w.ServiceType)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), usersMap[w.UserID].FIO)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), w.Quantity)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), w.Price)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), w.Quantity*w.Price)
	}

	f.SetColWidth(sheet, "A", "A", 15)
	f.SetColWidth(sheet, "B", "C", 25)
	f.SetColWidth(sheet, "D", "E", 20)
	f.SetColWidth(sheet, "F", "H", 12)

	filename := fmt.Sprintf("project_full_%s.xlsx", proj.Name)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(w)
}

// ExportSingleWork выгружает одну работу
func ExportSingleWork(w http.ResponseWriter, r *http.Request, store ds.Storage, id ds.ID) {
	var items []struct {
		ID          ds.ID   `db:"id"`
		Date        string  `db:"date"`
		ProjectID   ds.ID   `db:"project_id"`
		SystemID    ds.ID   `db:"system_id"`
		ProductID   ds.ID   `db:"product_id"`
		ServiceType string  `db:"service_type"`
		UserID      ds.ID   `db:"user_id"`
		Quantity    float64 `db:"quantity"`
		Price       float64 `db:"price"`
	}
	err := store.Select(r.Context(), &items, database.SqlFilter{
		Table:  ds.WorksRoute.String(),
		Fields: "id,date,project_id,system_id,product_id,service_type,user_id,quantity,price",
		Where:  "id = ?",
		Limit:  "1",
		Args:   []any{id},
	})
	if err != nil || len(items) == 0 {
		http.Error(w, "Работа не найдена", http.StatusNotFound)
		return
	}
	item := items[0]

	usersMap, _ := user.GetUsersMap(r.Context(), store)
	productsMap, _ := product.GetMap(r.Context(), store)
	projectsMap, _ := project.GetMap(r.Context(), store)
	systemsMap, _ := system.GetMap(r.Context(), store)

	f := excelize.NewFile()
	sheet := "Работа"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	rows := [][]interface{}{
		{"Дата", item.Date},
		{"Проект", projectsMap[item.ProjectID].Name},
		{"Система", systemsMap[item.SystemID].Name},
		{"Товар", productsMap[item.ProductID].Name},
		{"Услуга", item.ServiceType},
		{"Исполнитель", usersMap[item.UserID].FIO},
		{"Кол-во", item.Quantity},
		{"Цена", item.Price},
		{"Сумма", item.Quantity * item.Price},
	}

	for i, row := range rows {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+1), row[0])
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+1), row[1])
	}

	f.SetColWidth(sheet, "A", "A", 20)
	f.SetColWidth(sheet, "B", "B", 40)

	filename := fmt.Sprintf("work_%d.xlsx", id)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	f.Write(w)
}

// ExportFormattedProject выгружает проект в формате строительной сметы
func ExportFormattedProject(w http.ResponseWriter, r *http.Request, store ds.Storage, id ds.ID) {
	proj, err := project.GetByID(r.Context(), store, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var systems []system.System
	_ = store.Select(r.Context(), &systems, database.SqlFilter{
		Table:  ds.SystemRoute.String(),
		Fields: "id," + utils.ExtractFields("db", system.System{}),
		Where:  "project_id = ?",
		Args:   []any{id},
	})

	productsMap, _ := product.GetMap(r.Context(), store)

	f := excelize.NewFile()
	sheet := "Смета"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	boldStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	borderStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	headers := []string{"№", "Наименование", "ед. изм.", "кол.", "м2 за 1 ед.", "итого м2/шт.", "цена за ед.", "итого сумма"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+2, 2)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, boldStyle)
	}

	f.SetCellValue(sheet, "C3", proj.Name)
	f.SetCellStyle(sheet, "C3", "C3", boldStyle)

	row := 4
	var grandTotal float64

	for _, sys := range systems {
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), sys.Name)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("I%d", row), boldStyle)
		row++

		var works []struct {
			ProductID   ds.ID   `db:"product_id"`
			Quantity    float64 `db:"quantity"`
			Price       float64 `db:"price"`
			ServiceType string  `db:"service_type"`
		}
		_ = store.Select(r.Context(), &works, database.SqlFilter{
			Table:   ds.WorksRoute.String(),
			Fields:  "product_id, quantity, price, service_type",
			Where:   "project_id = ? AND system_id = ?",
			OrderBy: "id",
			Args:    []any{id, sys.ID},
		})

		var sysTotal float64

		for i, w := range works {
			pr := productsMap[w.ProductID]
			name := fmt.Sprintf("%s (%s)", pr.Name, w.ServiceType)
			unit := pr.Unit
			qty := w.Quantity
			square := pr.Square

			var totalSquare float64
			if square > 0 {
				totalSquare = square * qty
			} else {
				totalSquare = qty
			}

			sum := qty * w.Price
			sysTotal += sum

			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), i+1)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), name)
			f.SetCellValue(sheet, fmt.Sprintf("D%d", row), unit)

			if qty > 0 {
				f.SetCellValue(sheet, fmt.Sprintf("E%d", row), qty)
			}
			if square > 0 {
				f.SetCellValue(sheet, fmt.Sprintf("F%d", row), square)
			}
			if totalSquare > 0 {
				f.SetCellValue(sheet, fmt.Sprintf("G%d", row), totalSquare)
			}

			f.SetCellValue(sheet, fmt.Sprintf("H%d", row), w.Price)
			f.SetCellValue(sheet, fmt.Sprintf("I%d", row), sum)

			f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("I%d", row), borderStyle)
			row++
		}

		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), "Итого по системе:")
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), sysTotal)
		f.SetCellStyle(sheet, fmt.Sprintf("H%d", row), fmt.Sprintf("I%d", row), boldStyle)
		grandTotal += sysTotal
		row += 2
	}

	f.SetCellValue(sheet, fmt.Sprintf("H%d", row), "ВСЕГО:")
	f.SetCellValue(sheet, fmt.Sprintf("I%d", row), grandTotal)
	f.SetCellStyle(sheet, fmt.Sprintf("H%d", row), fmt.Sprintf("I%d", row), boldStyle)

	f.SetColWidth(sheet, "B", "B", 5)
	f.SetColWidth(sheet, "C", "C", 50)
	f.SetColWidth(sheet, "D", "G", 12)
	f.SetColWidth(sheet, "H", "I", 15)

	filename := fmt.Sprintf("smeta_%s.xlsx", proj.Name)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := f.Write(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ExportSelectedWorks(w http.ResponseWriter, r *http.Request, store ds.Storage) {
	// Читаем из URL: /export/smeta-selected?ids=1,2,3
	idsStr := r.URL.Query().Get("ids")
	if idsStr == "" {
		http.Error(w, "Не выбрано ни одной работы", http.StatusBadRequest)
		return
	}
	// ... дальше ваш код без r.ParseForm() ...

	idParts := strings.Split(idsStr, ",")
	ids := make([]any, len(idParts))
	for i, p := range idParts {
		id, _ := strconv.Atoi(strings.TrimSpace(p))
		ids[i] = id
	}

	var works []struct {
		ID          ds.ID   `db:"id"`
		ProjectID   ds.ID   `db:"project_id"`
		SystemID    ds.ID   `db:"system_id"`
		ProductID   ds.ID   `db:"product_id"`
		Quantity    float64 `db:"quantity"`
		Price       float64 `db:"price"`
		ServiceType string  `db:"service_type"`
	}

	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]

	query := fmt.Sprintf(`SELECT id, project_id, system_id, product_id, quantity, price, service_type 
		FROM works WHERE id IN (%s) ORDER BY project_id, system_id, id`, placeholders)

	if err := store.SelectX(r.Context(), &works, query, ids...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	productsMap, _ := product.GetMap(r.Context(), store)
	projectsMap, _ := project.GetMap(r.Context(), store)
	systemsMap, _ := system.GetMap(r.Context(), store)

	f := excelize.NewFile()
	sheet := "Смета"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	boldStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	borderStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	headers := []string{"№", "Наименование", "ед. изм.", "кол.", "м2 за 1 ед.", "итого м2/шт.", "цена за ед.", "итого сумма"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+2, 2)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, boldStyle)
	}

	row := 3
	var grandTotal float64

	currentProjectID := ds.ID(0)
	currentSystemID := ds.ID(0)
	var sysTotal float64
	workIndex := 1

	for _, w := range works {
		if w.ProjectID != currentProjectID {
			if currentSystemID != 0 {
				f.SetCellValue(sheet, fmt.Sprintf("H%d", row), "Итого по системе:")
				f.SetCellValue(sheet, fmt.Sprintf("I%d", row), sysTotal)
				f.SetCellStyle(sheet, fmt.Sprintf("H%d", row), fmt.Sprintf("I%d", row), boldStyle)
				grandTotal += sysTotal
				row += 2
			}

			proj := projectsMap[w.ProjectID]
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), proj.Name)
			f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), boldStyle)
			row++
			currentProjectID = w.ProjectID
			currentSystemID = 0
		}

		if w.SystemID != currentSystemID {
			if currentSystemID != 0 {
				f.SetCellValue(sheet, fmt.Sprintf("H%d", row), "Итого по системе:")
				f.SetCellValue(sheet, fmt.Sprintf("I%d", row), sysTotal)
				f.SetCellStyle(sheet, fmt.Sprintf("H%d", row), fmt.Sprintf("I%d", row), boldStyle)
				grandTotal += sysTotal
				row += 2
			}

			sys := systemsMap[w.SystemID]
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), sys.Name)
			f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("I%d", row), boldStyle)
			row++
			currentSystemID = w.SystemID
			sysTotal = 0
			workIndex = 1
		}

		pr := productsMap[w.ProductID]
		name := fmt.Sprintf("%s (%s)", pr.Name, w.ServiceType)
		totalSquare := w.Quantity
		if pr.Square > 0 {
			totalSquare = pr.Square * w.Quantity
		}
		sum := w.Quantity * w.Price
		sysTotal += sum

		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), workIndex)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), name)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), pr.Unit)
		if w.Quantity > 0 {
			f.SetCellValue(sheet, fmt.Sprintf("E%d", row), w.Quantity)
		}
		if pr.Square > 0 {
			f.SetCellValue(sheet, fmt.Sprintf("F%d", row), pr.Square)
		}
		if totalSquare > 0 {
			f.SetCellValue(sheet, fmt.Sprintf("G%d", row), totalSquare)
		}
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), w.Price)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), sum)

		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("I%d", row), borderStyle)
		row++
		workIndex++
	}

	if currentSystemID != 0 {
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), "Итого по системе:")
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), sysTotal)
		f.SetCellStyle(sheet, fmt.Sprintf("H%d", row), fmt.Sprintf("I%d", row), boldStyle)
		grandTotal += sysTotal
		row += 2
	}

	f.SetCellValue(sheet, fmt.Sprintf("H%d", row), "ВСЕГО:")
	f.SetCellValue(sheet, fmt.Sprintf("I%d", row), grandTotal)
	f.SetCellStyle(sheet, fmt.Sprintf("H%d", row), fmt.Sprintf("I%d", row), boldStyle)

	f.SetColWidth(sheet, "B", "B", 5)
	f.SetColWidth(sheet, "C", "C", 50)
	f.SetColWidth(sheet, "D", "G", 12)
	f.SetColWidth(sheet, "H", "I", 15)

	filename := fmt.Sprintf("smeta_selected_%s.xlsx", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := f.Write(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
