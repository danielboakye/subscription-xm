package data

import (
	"database/sql"
	"fmt"
	"time"
)

func TestNew(dbPool *sql.DB) Models {
	db = dbPool

	return Models{
		User: &UserTest{},
		Plan: &PlanTest{},
	}
}

type UserTest struct {
	ID        int
	Email     string
	FirstName string
	LastName  string
	Password  string
	Active    int
	IsAdmin   int
	CreatedAt time.Time
	UpdatedAt time.Time
	Plan      *Plan
}

func (u *UserTest) GetByEmail(email string) (*User, error) {
	user := User{
		ID:        1,
		Email:     "user@example.com",
		FirstName: "User",
		LastName:  "user",
		Password:  "$2a$12$dM0KXsvka4k8ahO/6s45Su2b6otbPICNDh6/DsTQLfB75SDazXcii",
		Active:    1,
		IsAdmin:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return &user, nil
}

func (u *UserTest) GetOne(id int) (*User, error) {
	return u.GetByEmail("")
}

func (u *UserTest) Update(user User) error {
	return nil
}

func (u *UserTest) Insert(user User) (int, error) {
	return 2, nil
}

func (u *UserTest) PasswordMatches(plainText string) (bool, error) {
	return true, nil
}

type PlanTest struct {
	ID                  int
	PlanName            string
	PlanAmount          int
	PlanAmountFormatted string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (p *PlanTest) GetAll() ([]*Plan, error) {
	var plans []*Plan

	plan := Plan{
		ID:         1,
		PlanName:   "Bronze Plan",
		PlanAmount: 1000,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	plans = append(plans, &plan)

	return plans, nil
}

func (p *PlanTest) GetOne(id int) (*Plan, error) {
	plan := Plan{
		ID:         1,
		PlanName:   "Bronze Plan",
		PlanAmount: 1000,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return &plan, nil
}

func (p *PlanTest) SubscribeUserToPlan(user User, plan Plan) error {
	return nil
}

func (p *PlanTest) AmountForDisplay() string {
	amount := float64(p.PlanAmount) / 100.0
	return fmt.Sprintf("$%.2f", amount)
}
