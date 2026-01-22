package main

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username     string             `bson:"username" json:"username"`
	City         string             `bson:"city" json:"city"`
	Lat          float64            `bson:"lat,omitempty" json:"lat,omitempty"`
	Lon          float64            `bson:"lon,omitempty" json:"lon,omitempty"`
	TelegramUser string             `bson:"telegram_user,omitempty" json:"telegram_user,omitempty"`
	Notify       bool               `bson:"notify" json:"notify"`
	NotifyTime   int                `bson:"notify_time" json:"notify_time"`
	IsAdmin      bool               `bson:"is_admin" json:"is_admin"`
	PasswordHash string             `bson:"password_hash,omitempty" json:"-"`
}

func GetUserByUsername(username string) (*User, error) {
	coll := MongoDB.Collection("users")
	var user User
	err := coll.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UpsertUser(user *User) error {
	coll := MongoDB.Collection("users")
	filter := bson.M{"username": user.Username}
	update := bson.M{"$set": user}
	_, err := coll.UpdateOne(context.Background(), filter, update, optionsUpdateUpsert())
	return err
}

func optionsUpdateUpsert() *options.UpdateOptions {
	opts := options.Update()
	upsert := true
	opts.SetUpsert(upsert)
	return opts
}

// CreateUser creates a new user with hashed password. Returns error if username exists.
func CreateUser(username, password string) error {
	coll := MongoDB.Collection("users")
	// check exists
	var existing User
	err := coll.FindOne(context.Background(), bson.M{"username": username}).Decode(&existing)
	if err == nil {
		return errors.New("username already exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u := &User{Username: username, PasswordHash: string(hash), Notify: false}
	_, err = coll.InsertOne(context.Background(), u)
	return err
}

// CheckUserPassword verifies username and password, returns user on success.
func CheckUserPassword(username, password string) (*User, error) {
	user, err := GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if user.PasswordHash == "" {
		return nil, errors.New("no password set")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}
