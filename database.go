package main

import (
	"errors"
	r "gopkg.in/gorethink/gorethink.v3"
	"os"
)

func InitConnectionPool() {
	var err error

	dbUrl := os.Getenv("DB")
	if dbUrl == "" {
		log.Fatal("db env variable not specified")
	}

	session, err = r.Connect(r.ConnectOpts{
		Address:    dbUrl,
		InitialCap: 10,
		MaxOpen:    10,
		Database:   "OverStatsNext",
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	res, err := r.Table("users").Changes().Run(session)
	if err != nil {
		log.Fatal(err)
	}

	var change Change
	for res.Next(&change) {
		SessionReport(change)
	}
}

func GetUser(ID int) (User, error) {
	res, err := r.Table("users").Get(ID).Run(session)
	if err != nil {
		return User{}, err
	}

	var user User
	err = res.One(&user)
	if err == r.ErrEmptyResult {
		return User{}, errors.New("db: row not found")
	}
	if err != nil {
		return User{}, err
	}

	defer res.Close()
	return user, nil
}

func GetRatingTop(platform string, limit int) ([]User, error) {
	var (
		res *r.Cursor
		err error
	)

	if platform == "console" {
		res, err = r.Table("users").Filter(r.Row.Field("region").Eq("psn").Or(r.Row.Field("region").Eq("xbl"))).OrderBy(r.Desc(r.Row.Field("profile").Field("Rating"))).Limit(limit).Run(session)
	} else {
		res, err = r.Table("users").Filter(r.Row.Field("region").Ne("psn").And(r.Row.Field("region").Ne("xbl"))).OrderBy(r.Desc(r.Row.Field("profile").Field("Rating"))).Limit(limit).Run(session)
	}

	if err != nil {
		return []User{}, err
	}

	var top []User
	err = res.All(&top)
	if err != nil {
		return []User{}, err
	}

	defer res.Close()
	return top, nil
}

func GetRank(id int64, hero string, fieldType string, field string) (float64, error) {
	res, err := r.Do(
		r.Table("users").Count(),
		r.Table("users").OrderBy(r.Asc(r.Row.Field("profile").Field("CompetitiveStats").Field("CareerStats").Field(hero).Field(fieldType).Field(field))).OffsetsOf(r.Row.Field("id").Eq(id)).Nth(0),
		func(count r.Term, position r.Term) r.Term {
			return position.Div(count).Mul(100)
		},
	).Run(session)

	var rank float64
	err = res.One(&rank)
	if err != nil {
		return 0, err
	}

	return rank, nil
}

func InsertUser(user User) (r.WriteResponse, error) {
	newDoc := map[string]interface{}{
		"id":      user.Id,
		"profile": user.Profile,
		"nick":    user.Nick,
		"region":  user.Region,
		"date":    r.Now(),
	}

	res, err := r.Table("users").Insert(newDoc, r.InsertOpts{
		Conflict: "replace",
	}).RunWrite(session)
	if err != nil {
		return r.WriteResponse{}, err
	}

	return res, nil
}