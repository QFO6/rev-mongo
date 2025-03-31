package revmongo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	BaseModel `bson:",inline"`
	Identity  string `bson:"Identity,omitempty"`
	Name      string `bson:"Name,omitempty"`
	Age       int    `bson:"Age,omitempty"`
}

var (
	dbName = "revmongo_test"
	dial   = "mongodb://localhost:27018"
)

func TestTextIndex(t *testing.T) {
	Dial = dial
	DBName = dbName
	MongoClient, err := qmgo.NewClient(context.Background(), &qmgo.Config{Uri: Dial})
	if err != nil {
		fmt.Println("Create qmgo client failed: ", err)
	}

	DB = MongoClient.Database(DBName)

	user := new(User)
	_, err = CreateTextIndex(user, []string{"Identity", "Name"})
	if err != nil {
		fmt.Println("Create text index error", err)
	}
}

func TestIndex(t *testing.T) {
	Dial = dial
	DBName = dbName
	MongoClient, err := qmgo.NewClient(context.Background(), &qmgo.Config{Uri: Dial})
	if err != nil {
		fmt.Println("Create qmgo client failed: ", err)
	}
	DB = MongoClient.Database(DBName)

	user := new(User)
	out, err := CreateIndex(user, []string{"Identity"}, true)
	if err != nil {
		fmt.Printf("CreateIndex with output: %v. Error: %v\n", out, err)
	}

	a := new(User)
	a.Identity = "E00000001"
	do := New(a)
	do.Operator = "E00000SYS"
	err = do.Create()
	if err != nil {
		fmt.Printf("Create document with identity %v error: %v\n", a.Identity, err)
	}

	b := new(User)
	b.Identity = "E00000002"
	do = New(b)
	do.Operator = "E00000SYS"
	err = do.CreateWithLog()
	if err != nil {
		fmt.Printf("CreateWithLog document with identity %v error: %v\n", b.Identity, err)
	}

	c := new(User)
	c.Identity = "E00000003"
	do = New(c)
	do.Operator = "E00000SYS"
	err = do.Create()
	if err != nil {
		fmt.Printf("Create document with identity %v error: %v\n", c.Identity, err)
	}

	d := new(User)
	do = New(d)
	do.Query = bson.M{"Identity": "E00000001"}
	do.GetByQ()
	err = do.DeleteWithLog()
	if err != nil {
		fmt.Printf("DeleteWithLog document with identity %v error: %v\n", d.Identity, err)
	}

	// create again
	e := new(User)
	e.Identity = "E00000001"
	e.Name = "Recreated user"
	do = New(e)
	err = do.CreateWithLog()
	if err != nil {
		fmt.Printf("CreateWithLog document with identity %v error: %v\n", e.Identity, err)
	}

	result, err := CollectionIndexes(&User{})
	if err != nil {
		fmt.Printf("Collection indexes: %v\n", result)
	}
}

func TestCRUD(t *testing.T) {
	Dial = dial
	DBName = dbName
	MongoClient, err := qmgo.NewClient(context.Background(), &qmgo.Config{Uri: Dial})
	if err != nil {
		fmt.Println("Create client failed: ", err)
	}
	DB = MongoClient.Database(DBName)

	f := new(User)
	do := New(f)
	do.Query = bson.M{"Identity": "E00000001"}
	err = do.GetByQ()
	if err != nil {
		fmt.Printf("Fetch user with query: %v, user: %v, err: %v\n", do.Query, f, err)
	}

	f.Name = "Updated User Name: " + time.Now().String()
	err = do.SaveWithLog()
	if err != nil {
		fmt.Println("Update document err:", err)
	}

	g := new(User)
	g.Id, _ = primitive.ObjectIDFromHex("64103c2e1e1537a49a433b11")
	do = New(g)
	err = do.Get()
	if err != nil {
		fmt.Printf("Fetch user: %v, err: %v\n", g, err)
	}

	h := new(User)
	h.Identity = "E00000005"
	do = New(h)
	err = do.CreateWithLog()
	if err != nil {
		fmt.Printf("Create %v with Id %v, error: %v\n", h.Identity, h.Id, err)
	}

	err = do.EraseWithLog()
	if err != nil {
		fmt.Println("EraseWithLog document err:", err)
	}

	// remove all documents
	// j := new(User)
	// do = New(j)
	// do.Query = bson.M{"Name": "Test User"}
	// _, err = do.RemoveAll()
	// if err != nil {
	// 	fmt.Printf("Remove users with query: %v, err: %v\n", do.Query, err)
	// }
}
