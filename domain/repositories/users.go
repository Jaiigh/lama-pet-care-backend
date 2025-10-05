package repositories

import (
	"context"
	ds "lama-backend/domain/datasources"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"

	"fmt"
	"time"
)

type usersRepository struct {
	Context    context.Context
	Collection *db.PrismaClient
}

type IUsersRepository interface {
	InsertUser(role string, data entities.CreatedUserModel) (*entities.UserDataModel, error)
	FindByEmailAndRole(email string, role string) (*entities.LoginUserResponseModel, error)
	FindByID(userID string) (*entities.UserDataModel, error)
	DeleteByID(userID string) (*entities.UserDataModel, error)
	UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error)
}

func NewUsersRepository(db *ds.PrismaDB) IUsersRepository {
	return &usersRepository{
		Context:    db.Context,
		Collection: db.PrismaDB,
	}
}

func (repo *usersRepository) InsertUser(role string, data entities.CreatedUserModel) (*entities.UserDataModel, error) {
	createdData, err := repo.Collection.Users.CreateOne(
		db.Users.Email.Set(data.Email),
		db.Users.Password.Set(data.Password),
		db.Users.Name.Set(data.Name),
		db.Users.Birthdate.Set(data.BirthDate),
		db.Users.TelephoneNumber.Set(data.TelephoneNumber),
		db.Users.Address.Set(data.Address),
		db.Users.Role.Set(db.Role(role)),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> InsertUser: %v", err)
	}

	return &entities.UserDataModel{
		UserID:          createdData.ID,
		CreatedAt:       createdData.CreatedAt,
		UpdatedAt:       createdData.UpdatedAt,
		Email:           createdData.Email,
		Password:        createdData.Password,
		Role:            createdData.Role,
		Name:            createdData.Name,
		BirthDate:       createdData.Birthdate,
		TelephoneNumber: createdData.TelephoneNumber,
		Address:         createdData.Address,
	}, nil
}

func (repo *usersRepository) FindByEmailAndRole(email string, role string) (*entities.LoginUserResponseModel, error) {
	user, err := repo.Collection.Users.FindUnique(
		db.Users.UsersEmailRoleKey(
			db.Users.Email.Equals(email),
			db.Users.Role.Equals(db.Role(role)),
		),
	).Exec(repo.Context)
	if err != nil {
		return nil, fmt.Errorf("users -> FindByEmail: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByEmail: user data is nil")
	}
	return &entities.LoginUserResponseModel{
		UserID:   user.ID,
		Email:    user.Email,
		Password: user.Password,
		Role:     user.Role,
	}, nil
}

func (repo *usersRepository) FindByID(userID string) (*entities.UserDataModel, error) {
	user, err := repo.Collection.Users.FindUnique(
		db.Users.ID.Equals(userID),
	).With(
		db.Users.Caretaker.Fetch(),
		db.Users.Doctor.Fetch(),
		db.Users.Owner.Fetch(),
	).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> FindByID: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("users -> FindByID: user data is nil")
	}

	return MapToEntities(user), nil
}

func (repo *usersRepository) DeleteByID(userID string) (*entities.UserDataModel, error) {
	deletedUser, err := repo.Collection.Users.FindUnique(
		db.Users.ID.Equals(userID),
	).Delete().Exec(repo.Context) // ลบและคืนค่าที่ถูกลบเลย

	if err != nil {
		return nil, fmt.Errorf("users -> DeleteByID: %v", err)
	}
	if deletedUser == nil {
		return nil, fmt.Errorf("users -> DeleteByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:          deletedUser.ID,
		CreatedAt:       deletedUser.CreatedAt,
		UpdatedAt:       deletedUser.UpdatedAt,
		Email:           deletedUser.Email,
		Password:        deletedUser.Password,
		Role:            deletedUser.Role,
		Name:            deletedUser.Name,
		BirthDate:       deletedUser.Birthdate,
		TelephoneNumber: deletedUser.TelephoneNumber,
		Address:         deletedUser.Address,
	}, nil
}

func (repo *usersRepository) UpdateByID(userID string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	// Start with an empty list of updates
	updates := []db.UsersSetParam{}

	if data.Email != nil {
		updates = append(updates, db.Users.Email.Set(*data.Email))
	}
	if data.Password != nil {
		updates = append(updates, db.Users.Password.Set(*data.Password))
	}
	if data.Name != nil {
		updates = append(updates, db.Users.Name.Set(*data.Name))
	}
	if data.BirthDate != nil {
		updates = append(updates, db.Users.Birthdate.Set(*data.BirthDate))
	}
	if data.TelephoneNumber != nil {
		updates = append(updates, db.Users.TelephoneNumber.Set(*data.TelephoneNumber))
	}
	if data.Address != nil {
		updates = append(updates, db.Users.Address.Set(*data.Address))
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("users -> UpdateByID: no fields to update")
	}

	// Execute update
	updatedUser, err := repo.Collection.Users.FindUnique(
		db.Users.ID.Equals(userID),
	).Update(updates...).Exec(repo.Context)

	if err != nil {
		return nil, fmt.Errorf("users -> UpdateByID: %v", err)
	}
	if updatedUser == nil {
		return nil, fmt.Errorf("users -> UpdateByID: user not found")
	}

	return &entities.UserDataModel{
		UserID:          updatedUser.ID,
		CreatedAt:       updatedUser.CreatedAt,
		UpdatedAt:       updatedUser.UpdatedAt,
		Email:           updatedUser.Email,
		Password:        updatedUser.Password,
		Role:            updatedUser.Role,
		Name:            updatedUser.Name,
		BirthDate:       updatedUser.Birthdate,
		TelephoneNumber: updatedUser.TelephoneNumber,
		Address:         updatedUser.Address,
	}, nil
}

func MapToEntities(user *db.UsersModel) *entities.UserDataModel {
	var licenseNumber, specialization string
	var startDate db.DateTime
	var startWorkingTime, endWorkingTime time.Time
	var rating, totalSpending db.Decimal
	doctor, ok := user.Doctor()
	if ok {
		licenseNumber = doctor.LicenseNumber
		startDate = doctor.StartDate
		startWorkingTime = doctor.StartWorkingTime
		endWorkingTime = doctor.EndWorkingTime
	}
	caretaker, ok := user.Caretaker()
	if ok {
		specialization, _ = caretaker.Specialties()
		rating, _ = caretaker.Rating()
		startWorkingTime = caretaker.StartWorkingTime
		endWorkingTime = caretaker.EndWorkingTime
	}
	owner, ok := user.Owner()
	if ok {
		totalSpending = owner.TotalSpending
	}

	return &entities.UserDataModel{
		UserID:          user.ID,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		Email:           user.Email,
		Password:        user.Password,
		Role:            user.Role,
		Name:            user.Name,
		BirthDate:       user.Birthdate,
		TelephoneNumber: user.TelephoneNumber,
		Address:         user.Address,
		LicenseNumber:   &licenseNumber,
		StartDate:       &startDate,
		StartWorkTime:   &startWorkingTime,
		EndWorkTime:     &endWorkingTime,
		Specialization:  &specialization,
		Rating:          &rating,
		TotalSpending:   &totalSpending,
	}
}
