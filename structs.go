package main

import (
	"github.com/sdwolfe32/ovrstat/ovrstat"
	"time"
)

type User struct {
	Id      string               `gorethink:"id"`
	Profile *ovrstat.PlayerStats `gorethink:"profile"`
	Nick    string               `gorethink:"nick"`
	Region  string               `gorethink:"region"`
	Date    time.Time            `gorethink:"date"`
	Patron  string               `gorethink:"patron"`
}

type Change struct {
	OldVal User `gorethink:"old_val"`
	NewVal User `gorethink:"new_val"`
}

type Report struct {
	Rating int `gorethink:"rating"`
	Level  int `gorethink:"level"`
	Games  int `gorethink:"games"`
	Wins   int `gorethink:"wins"`
	Ties   int `gorethink:"ties"`
	Losses int `gorethink:"losses"`
}

type Top struct {
	Place int     `gorethink:"place"`
	Rank  float64 `gorethink:"rank"`
}
