package revmongo

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/qiniu/qmgo"
	"github.com/revel/revel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	DBName string
	Dial   string
	Client *qmgo.Client
	DB     *qmgo.Database
)

func Init() {
	Connect()
	objID := primitive.NewObjectID()
	revel.TypeBinders[reflect.TypeOf(objID)] = ObjectIDBinder
}

func NewClient(ctx context.Context) (*qmgo.Client, error) {
	if Dial == "" {
		return nil, fmt.Errorf("Mongodb connection not defined")
	}
	return qmgo.NewClient(ctx, &qmgo.Config{Uri: Dial})
}

// Connect to database and return client
func Connect() {
	var err error
	var found bool

	defer func() {
		if Client != nil {
			DB = Client.Database(DBName)
			if err != nil {
				revel.AppLog.Errorf("Could not connect to Mongo DB. Error: %s", err)
			}
		}
	}()

	if Dial, found = revel.Config.String("mongodb.dial"); !found {
		revel.AppLog.Crit("Mongodb connection not defined")
	}

	if DBName, found = revel.Config.String("mongodb.name"); !found {
		urls := strings.Split(Dial, "/")
		DBName = urls[len(urls)-1]
	}

	if Client == nil {
		ctx := context.Background()
		Client, err = qmgo.NewClient(ctx, &qmgo.Config{Uri: Dial})
		if err != nil {
			Client = nil
			revel.AppLog.Errorf("Could not connect to Mongo DB. Error: %v", err)
			for i := 0; i < 3; i++ {
				revel.AppLog.Info("Retry connect to database ...")
				time.Sleep(3 * time.Second)
				Client, err = qmgo.NewClient(ctx, &qmgo.Config{Uri: Dial})
				if err == nil {
					break
				} else {
					revel.AppLog.Errorf("Retry time %v could not connect to Mongo DB. Error: %v", i+1, err)
				}

			}
		}
	}
}

// ObjectIDBinder do binding
var ObjectIDBinder = revel.Binder{
	// Make a ObjectId from a request containing it in string format.
	Bind: revel.ValueBinder(func(val string, typ reflect.Type) reflect.Value {
		if len(val) == 0 {
			return reflect.Zero(typ)
		}
		if objID, err := primitive.ObjectIDFromHex(val); err == nil {
			return reflect.ValueOf(objID)
		}

		revel.AppLog.Errorf("ObjectIDBinder.Bind - invalid ObjectId!")
		return reflect.Zero(typ)
	}),
	// Turns ObjectId back to hexString for reverse routing
	Unbind: func(output map[string]string, name string, val interface{}) {
		var hexStr string
		hexStr = fmt.Sprintf("%s", val.(primitive.ObjectID).Hex())
		// not sure if this is too carefull but i wouldn't want invalid ObjectIds in my App
		output[name] = hexStr
	},
}
