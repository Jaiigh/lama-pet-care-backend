package repositories

import (
	"context"
	"errors"
	"fmt"
	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
	"time"
)

type paymentRepository struct {
	Context    context.Context
	Collection *db.PrismaClient
}

type IPaymentRepository interface {
	InsertPayment(user_id string, price int) (*entities.PaymentModel, error)
	FindByID(payID string) (*entities.PaymentModel, error)
	DeleteByID(payID string) (*entities.PaymentModel, error)
	UpdateByID(paymentID string, data entities.PaymentModel) (*entities.PaymentModel, error)
	FindAllPayments(month int, year int, offset, limit int) ([]*entities.PaymentModel, int, error)
	FindPaymentsByOwnerID(ownerID string, month int, year int, offset, limit int) ([]*entities.PaymentModel, int, error)
}

func NewPaymentRepository(db *ds.PrismaDB) IPaymentRepository {
	return &paymentRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *paymentRepository) InsertPayment(user_id string, price int) (*entities.PaymentModel, error) {
	createdData, err := repo.Collection.Payment.CreateOne(
		db.Payment.Price.Set(price),
		db.Payment.Status.Set(db.PaymentStatusUnpaid),
		db.Payment.Owner.Link(db.Owner.UserID.Equals(user_id)),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("payment -> InsertPayment: %v", err)
	}

	return mapToPaymentModel(createdData), nil
}

func (repo *paymentRepository) FindByID(payID string) (*entities.PaymentModel, error) {
	payment, err := repo.Collection.Payment.FindUnique(
		db.Payment.Payid.Equals(payID),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("payment -> FindByID: %v", err)
	}
	if payment == nil {
		return nil, fmt.Errorf("payment -> FindByID: payment data is nil")
	}

	return mapToPaymentModel(payment), nil
}

func (repo *paymentRepository) DeleteByID(payID string) (*entities.PaymentModel, error) {
	deletedPayment, err := repo.Collection.Payment.FindUnique(
		db.Payment.Payid.Equals(payID),
	).Delete().Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("payment -> DeleteByID: %v", err)
	}
	if deletedPayment == nil {
		return nil, fmt.Errorf("payment -> DeleteByID: payment not found")
	}

	return mapToPaymentModel(deletedPayment), nil
}

func (repo *paymentRepository) UpdateByID(paymentID string, data entities.PaymentModel) (*entities.PaymentModel, error) {
	updates := []db.PaymentSetParam{}

	if data.Status != "" {
		updates = append(updates, db.Payment.Status.Set(db.PaymentStatus(data.Status)))
	}

	if data.Type != nil && *data.Type != "" {
		updates = append(updates, db.Payment.Type.Set(*data.Type))
	}

	if data.PayDate != nil {
		updates = append(updates, db.Payment.PayDate.Set(*data.PayDate))
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("payment -> UpdateByID: no fields to update")
	}

	updatedPayment, err := repo.Collection.Payment.FindUnique(
		db.Payment.Payid.Equals(paymentID),
	).Update(updates...).Exec(repo.Context)

	if err != nil {

		if errors.Is(err, db.ErrNotFound) {
			return nil, db.ErrNotFound
		}

		return nil, fmt.Errorf("payment -> UpdateByID: %v", err)
	}

	return mapToPaymentModel(updatedPayment), nil
}

func mapToPaymentModel(model *db.PaymentModel) *entities.PaymentModel {
	paymentType, ok := model.Type()
	if !ok {
		paymentType = ""
	}
	payDate, ok := model.PayDate()
	if !ok {
		payDate = time.Time{}
	}

	return &entities.PaymentModel{
		PayID:   model.Payid,
		OwnerID: model.Oid,
		Status:  model.Status,
		Price:   model.Price,
		Type:    &paymentType,
		PayDate: &payDate,
	}
}
func mapToPaymentModels(models []db.PaymentModel) []*entities.PaymentModel {
	payments := make([]*entities.PaymentModel, len(models))
	for i := range models {
		payments[i] = mapToPaymentModel(&models[i])
	}
	return payments
}

func (repo *paymentRepository) FindAllPayments(month int, year int, offset, limit int) ([]*entities.PaymentModel, int, error) {
	params := []db.PaymentWhereParam{}

	if year > 0 {
		if month <= 0 {
			month = 1
		}
		params = addPayDateParams(params, month, year)
	}

	var sqlResult []entities.CountResult
	sql, args, err := getSqlPayment("all", "", month, year)
	if err != nil {
		return nil, 0, err
	}
	err = repo.Collection.Prisma.QueryRaw(sql, args...).Exec(repo.Context, &sqlResult)
	if err != nil {
		return nil, 0, err
	}
	total := sqlResult[0].Count

	payments, err := repo.Collection.Payment.FindMany(params...).OrderBy(
		db.Payment.PayDate.Order(db.SortOrderAsc),
	).Skip(offset).Take(limit).Exec(repo.Context)
	if err != nil {
		return nil, 0, err
	}

	return mapToPaymentModels(payments), total, nil
}

func (repo *paymentRepository) FindPaymentsByOwnerID(ownerID string, month int, year int, offset, limit int) ([]*entities.PaymentModel, int, error) {
	params := []db.PaymentWhereParam{
		db.Payment.Oid.Equals(ownerID),
	}

	if year > 0 {
		if month <= 0 {
			month = 1
		}
		params = addPayDateParams(params, month, year)
	}

	var sqlResult []entities.CountResult
	sql, args, err := getSqlPayment("owner", ownerID, month, year)
	if err != nil {
		return nil, 0, err
	}
	err = repo.Collection.Prisma.QueryRaw(sql, args...).Exec(repo.Context, &sqlResult)
	if err != nil {
		return nil, 0, err
	}
	total := sqlResult[0].Count

	payments, err := repo.Collection.Payment.FindMany(params...).OrderBy(
		db.Payment.PayDate.Order(db.SortOrderAsc),
	).Skip(offset).Take(limit).Exec(repo.Context)
	if err != nil {
		return nil, 0, err
	}

	return mapToPaymentModels(payments), total, nil
}

func addPayDateParams(params []db.PaymentWhereParam, month, year int) []db.PaymentWhereParam {

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	params = append(params,
		db.Payment.PayDate.Gte(startDate),
	)

	return params
}

func getSqlPayment(sqltype, userID string, month, year int) (string, []interface{}, error) {
	whereSQL := ""
	args := []interface{}{}
	idx := 1

	switch sqltype {
	case "owner":
		whereSQL += fmt.Sprintf(`"OID" = $%d::uuid`, idx)
		args = append(args, userID)
		idx++
	default:
	}

	if month > 0 && year > 0 {
		if whereSQL != "" {
			whereSQL += " AND "
		}
		whereSQL += fmt.Sprintf(`
			(
				pay_date >= $%d
			)
		`, idx)

		startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

		args = append(args, startDate)
		idx++
	}

	if whereSQL != "" {
		whereSQL = "WHERE " + whereSQL
	}

	sql := fmt.Sprintf(`
		SELECT CAST(COUNT(*) AS INTEGER) AS count
		FROM "Service"
		%s`, whereSQL)
	return sql, args, nil
}
