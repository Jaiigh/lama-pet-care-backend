package repositories

import (
	"context"
	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"

	"fmt"
)

type paymentRepository struct {
	Context    context.Context
	Collection *db.PrismaClient
}

type IPaymentRepository interface {
	InsertPayment(user_id string) (*entities.PaymentModel, error)
	FindByID(payID string) (*entities.PaymentModel, error)
	DeleteByID(payID string) (*entities.PaymentModel, error)
	UpdateByID(data entities.PaymentModel) (*entities.PaymentModel, error)
}

func NewPaymentRepository(db *ds.PrismaDB) IPaymentRepository {
	return &paymentRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *paymentRepository) InsertPayment(user_id string) (*entities.PaymentModel, error) {
	createdData, err := repo.Collection.Payment.CreateOne(
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

func (repo *paymentRepository) UpdateByID(data entities.PaymentModel) (*entities.PaymentModel, error) {
	updates := []db.PaymentSetParam{}

	if data.Status != "" {
		updates = append(updates, db.Payment.Status.Set(data.Status))
	}
	if *data.Type != "" {
		updates = append(updates, db.Payment.Type.Set(*data.Type))
	}
	if data.PayDate != nil {
		updates = append(updates, db.Payment.PayDate.Set(*data.PayDate))
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("payment -> UpdateByID: no fields to update")
	}

	updatedPayment, err := repo.Collection.Payment.FindUnique(
		db.Payment.Payid.Equals(data.PayID),
	).Update(updates...).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("payment -> UpdateByID: %v", err)
	}
	if updatedPayment == nil {
		return nil, fmt.Errorf("payment -> UpdateByID: payment not found")
	}

	return mapToPaymentModel(updatedPayment), nil
}

func mapToPaymentModel(model *db.PaymentModel) *entities.PaymentModel {
	paymentType, _ := model.Type()
	payDate, _ := model.PayDate()

	return &entities.PaymentModel{
		PayID:   model.Payid,
		OwnerID: model.Oid,
		Status:  model.Status,
		Type:    &paymentType,
		PayDate: &payDate,
	}
}
