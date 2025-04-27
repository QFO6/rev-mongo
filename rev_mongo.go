package revmongo

import (
	"context"
	"errors"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/qiniu/qmgo"
	qopts "github.com/qiniu/qmgo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Do wrap all common functions
type Do struct {
	model    interface{}
	Ctx      context.Context
	Query    bson.M
	Client   *qmgo.Client
	Coll     *qmgo.Collection
	Sort     []string
	Skip     int64
	Limit    int64
	Select   []string
	Operator string
	Reason   string
	SaveLog  bool // default is false
}

// New using mongodo Client
func New(model interface{}) *Do {
	colName := getModelName(model)
	coll := DB.Collection(colName)
	return &Do{model: model, Coll: coll, Ctx: context.Background()}
}

// NewDo initiate with input model and db name, backfoward compatible
func NewDo(dbName string, model interface{}) *Do {
	colName := getModelName(model)
	coll := Client.Database(dbName).Collection(colName)
	return &Do{model: model, Coll: coll, Ctx: context.Background()}
}

// NewWithDB initiate with input model and db name
func NewWithDBName(dbName string, model interface{}) *Do {
	colName := getModelName(model)
	coll := Client.Database(dbName).Collection(colName)
	return &Do{model: model, Coll: coll, Ctx: context.Background()}
}

// New with C, with Coll Name for Coll name diff with model name
func NewWithC(model interface{}, cName string) *Do {
	Coll := Collection(DBName, cName)
	return &Do{model: model, Coll: Coll, Ctx: context.Background()}
}

// NewWithCtx with Ctx input and model
func NewWithCtx(ctx context.Context, model interface{}) *Do {
	colName := getModelName(model)
	coll := DB.Collection(colName)
	return &Do{model: model, Coll: coll, Ctx: ctx}
}

func Collection(dbName string, m interface{}) *qmgo.Collection {
	collectionName := getModelName(m)
	return Client.Database(dbName).Collection(collectionName)
}

// Create will generate Id for model
func (m *Do) Create() error {
	timeNow := time.Now()
	id := reflect.ValueOf(m.model).Elem().FieldByName("Id")
	x := reflect.ValueOf(m.model).Elem().FieldByName("CreatedAt")
	if x.IsValid() {
		x.Set(reflect.ValueOf(timeNow))
	}
	by := reflect.ValueOf(m.model).Elem().FieldByName("CreatedBy")
	if by.IsValid() {
		by.Set(reflect.ValueOf(m.Operator))
	}
	l := reflect.ValueOf(m.model).Elem().FieldByName("LatestTime")
	if l.IsValid() {
		l.Set(reflect.ValueOf(timeNow))
	}
	isRemoved := reflect.ValueOf(m.model).Elem().FieldByName("IsRemoved")
	if isRemoved.IsValid() {
		removeValue := false
		isRemoved.Set(reflect.ValueOf(&removeValue))
	}
	// Important: make sure the sessCtx used in every operation in the whole transaction
	// start transaction
	if result, err := m.Coll.InsertOne(m.Ctx, m.model); err != nil {
		return err
	} else {
		if id.IsValid() {
			id.Set(reflect.ValueOf(result.InsertedID))
		}
	}

	if !m.SaveLog {
		return nil
	}
	time.Sleep(1 * time.Second)
	if err := m.saveLog(CREATE); err != nil {
		return err
	}

	return nil
}

// CreateWithLog record log for creation
func (m *Do) CreateWithLog() error {
	var err error
	err = m.Create()
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)
	err = m.saveLog(CREATE)
	if err != nil {
		return err
	}
	return nil
}

// Save method, upsert record with UpdatedAt as now
func (m *Do) Save() error {
	timeNow := time.Now()
	id := reflect.ValueOf(m.model).Elem().FieldByName("Id")
	x := reflect.ValueOf(m.model).Elem().FieldByName("UpdatedAt")
	if x.IsValid() {
		x.Set(reflect.ValueOf(timeNow))
	}
	by := reflect.ValueOf(m.model).Elem().FieldByName("UpdatedBy")
	if by.IsValid() {
		by.Set(reflect.ValueOf(m.Operator))
	}
	l := reflect.ValueOf(m.model).Elem().FieldByName("LatestTime")
	if l.IsValid() {
		l.Set(reflect.ValueOf(timeNow))
	}

	if err := m.Coll.UpdateOne(m.Ctx, bson.M{"_id": id.Interface()}, bson.M{"$set": m.model}, qopts.UpdateOptions{
		UpdateOptions: options.Update().SetUpsert(true),
	}); err != nil {
		return err
	}

	if !m.SaveLog {
		return nil
	}

	if err := m.saveLog(UPDATE); err != nil {
		return err
	}

	return nil
}

// SaveIf update record according to Id and Query
func (m *Do) SaveIf() error {
	timeNow := time.Now()
	id := reflect.ValueOf(m.model).Elem().FieldByName("Id")
	x := reflect.ValueOf(m.model).Elem().FieldByName("UpdatedAt")
	if x.IsValid() {
		x.Set(reflect.ValueOf(timeNow))
	}
	by := reflect.ValueOf(m.model).Elem().FieldByName("UpdatedBy")
	if by.IsValid() {
		by.Set(reflect.ValueOf(m.Operator))
	}
	l := reflect.ValueOf(m.model).Elem().FieldByName("LatestTime")
	if l.IsValid() {
		l.Set(reflect.ValueOf(timeNow))
	}

	// check query
	if m.Query == nil {
		return errors.New("Query must be defined for SaveIf")
	}
	query := m.Query
	query["_id"] = id.Interface()

	if err := m.Coll.UpdateOne(m.Ctx, query, bson.M{"$set": m.model}, qopts.UpdateOptions{
		UpdateOptions: options.Update().SetUpsert(true),
	}); err != nil {
		return err
	}

	if !m.SaveLog {
		return nil
	}
	if err := m.saveLog(UPDATE); err != nil {
		return err
	}

	return nil
}

// SaveIf update record according to Id and Query
func (m *Do) SaveIfWithLog() error {
	timeNow := time.Now()
	id := reflect.ValueOf(m.model).Elem().FieldByName("Id")
	x := reflect.ValueOf(m.model).Elem().FieldByName("UpdatedAt")
	if x.IsValid() {
		x.Set(reflect.ValueOf(timeNow))
	}
	by := reflect.ValueOf(m.model).Elem().FieldByName("UpdatedBy")
	if by.IsValid() {
		by.Set(reflect.ValueOf(m.Operator))
	}
	l := reflect.ValueOf(m.model).Elem().FieldByName("LatestTime")
	if l.IsValid() {
		l.Set(reflect.ValueOf(timeNow))
	}

	// check query
	if m.Query == nil {
		return errors.New("Query must be defined for SaveIf")
	}
	query := m.Query
	query["_id"] = id.Interface()

	if err := m.Coll.UpdateOne(m.Ctx, query, bson.M{"$set": m.model}, qopts.UpdateOptions{
		UpdateOptions: options.Update().SetUpsert(true),
	}); err != nil {
		return err
	}

	if err := m.saveLog(UPDATE); err != nil {
		return err
	}
	return nil
}

// SaveWithLog save record and inset a new changelog record
func (m *Do) SaveWithLog() error {
	if err := m.Save(); err != nil {
		return err
	}
	if err := m.saveLog(UPDATE); err != nil {
		return err
	}
	return nil
}

// Erase is hard delete according Id
func (m *Do) Erase() error {
	id := reflect.ValueOf(m.model).Elem().FieldByName("Id")
	if err := m.Coll.Remove(m.Ctx, bson.M{"_id": id.Interface()}); err != nil {
		return err
	}

	if !m.SaveLog {
		return nil
	}

	if err := m.saveLog(ERASE); err != nil {
		return err
	}
	return nil
}

// Remove is hard delete
func (m *Do) Remove() error {
	id := reflect.ValueOf(m.model).Elem().FieldByName("Id")
	if err := m.Coll.Remove(m.Ctx, bson.M{"_id": id.Interface()}); err != nil {
		return err
	}

	if !m.SaveLog {
		return nil
	}

	if err := m.saveLog(REMOVE); err != nil {
		return err
	}
	return nil
}

// EraseWithLog, hard delete record and insert a chagnelog
func (m *Do) EraseWithLog() error {
	if err := m.Erase(); err != nil {
		return err
	}

	if err := m.saveLog(ERASE); err != nil {
		return err
	}
	return nil
}

// Delete is softe delete
func (m *Do) Delete() error {
	id := reflect.ValueOf(m.model).Elem().FieldByName("Id")
	x := reflect.ValueOf(m.model).Elem().FieldByName("RemovedAt")
	if x.IsValid() {
		x.Set(reflect.ValueOf(time.Now()))
	}
	by := reflect.ValueOf(m.model).Elem().FieldByName("RemovedBy")
	if by.IsValid() {
		by.Set(reflect.ValueOf(m.Operator))
	}
	isRemoved := reflect.ValueOf(m.model).Elem().FieldByName("IsRemoved")
	if isRemoved.IsValid() {
		removeValue := true
		isRemoved.Set(reflect.ValueOf(&removeValue))
	}

	// check IsLocked flag
	record := map[string]interface{}{}
	m.Coll.Find(m.Ctx, bson.D{bson.E{Key: "_id", Value: id}}).Select(bson.M{"IsLocked": 1}).One(&record)
	if record != nil {
		if v, found := record["IsLocked"]; found {
			if v.(bool) {
				return errors.New("Record locked for delete.")
			}
		}
	}

	if err := m.Coll.UpdateOne(m.Ctx, bson.M{"_id": id.Interface()}, bson.M{"$set": m.model}); err != nil {
		return err
	}

	if !m.SaveLog {
		return nil
	}

	if err := m.saveLog(DELETE); err != nil {
		return err
	}

	return nil
}

// DeleteWithLog
func (m *Do) DeleteWithLog() error {
	if err := m.Delete(); err != nil {
		return err
	}

	if err := m.saveLog(DELETE); err != nil {
		return err
	}
	return nil
}

// Erase all is hard Delete with raw condition (no predefined skip IsRemoved:true)
func (m *Do) EraseAll() error {
	_, err := m.RemoveAll()
	return err
}

// RemoveAll is hardDelete
func (m *Do) RemoveAll() (int64, error) {
	if m.Query == nil {
		return 0, errors.New("Cannot remove without condition")
	}

	result, err := m.Coll.RemoveAll(m.Ctx, m.Query)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// Erase all with log
func (m *Do) EraseAllWithLog() error {
	if err := m.EraseAll(); err != nil {
		return err
	}
	if err := m.saveLog(ERASE); err != nil {
		return err
	}
	return nil
}

// DirectSave method, upsert record without set UpdatedBy and UpdatedAt
func (m *Do) DirectSave() error {
	id := reflect.ValueOf(m.model).Elem().FieldByName("Id")
	// check IsLocked flag
	record := map[string]interface{}{}
	m.Coll.Find(m.Ctx, bson.D{bson.E{Key: "_id", Value: id}}).Select(bson.M{"IsLocked": 1}).One(&record)
	if record != nil {
		if v, found := record["IsLocked"]; found {
			if v.(bool) {
				return errors.New("Record is locked for update.")
			}
		}
	}

	err := m.Coll.UpdateOne(m.Ctx, bson.M{"_id": id.Interface()}, bson.M{"$set": m.model})
	return err
}

// DirectSaveWithLog save record and inset a new changelog record
func (m *Do) DirectSaveWithLog() error {
	if err := m.DirectSave(); err != nil {
		return err
	}
	if err := m.saveLog(UPDATE); err != nil {
		return err
	}
	return nil
}

// ---------- General revmongo fetch functions -----------

// GenQuery export qmgo.QueryI for further query chain
func (m *Do) Q() qmgo.QueryI {
	return m.findQ()
}

// Count
func (m *Do) Count() int64 {
	query := m.findQ()
	count, _ := query.Count()
	return int64(count)
}

// ---------retrieve functions
// FindAll except removed, i is interface address
func (m *Do) FindAll(i interface{}) error {
	return m.findQ().All(i)
}

// FindAll except removed, i is interface address
func (m *Do) FindAllIncludeRemoved(i interface{}) error {
	return m.findIncludeRemovedQ().All(i)
}

// Get will retrieve by _id
func (m *Do) Get() error {
	return m.findByIdQ().One(m.model)
}

// GetByQ get first one based on query, model will be updated
func (m *Do) GetByQ() error {
	return m.findQ().One(m.model)
}

// Fetch same as Get, but bind to another struct (uses for model name diff)
func (m *Do) Fetch(record interface{}) error {
	err := m.findByIdQ().One(record)
	return err
}

// QueryIncludeRemoved get first one based on query include isRemoved: true, model will be updated
func (m *Do) QueryIncludeRemoved() error {
	return m.findIncludeRemovedQ().One(m.model)
}

// FetchByQ match result to a structure
func (m *Do) FetchByQ(record interface{}) error {
	return m.findQ().One(record)
}

// Select query and select columns
func (m *Do) FindWithSelect(i interface{}, cols []string) error {
	sCols := bson.M{}
	for _, v := range cols {
		if strings.HasPrefix(v, "-") {
			t := v[1 : len(v)-1]
			sCols[t] = -1
		} else {
			sCols[v] = 1
		}
	}
	return m.findQ().Select(sCols).All(i)
}

// Distinct
func (m *Do) Distinct(key string, i interface{}) error {
	return m.findQ().Distinct(key, i)
}

// GetWithSelect, limit cols
func (m *Do) GetWithSelect(cols []string) error {
	sCols := bson.M{}
	for _, v := range cols {
		if strings.HasPrefix(v, "-") {
			t := v[1 : len(v)-1]
			sCols[t] = -1
		} else {
			sCols[v] = 1
		}
	}
	return m.findByIdQ().Select(sCols).One(m.model)
}

// FetchByQAndDelete find One record according to Query and mark as IsRemoved
func (m *Do) FetchByQAndDelete() error {
	colName := getModelName(m.model)
	coll := DB.Collection(colName)
	if m.Query == nil {
		m.Query = bson.M{}
	}

	m.Query["IsRemoved"] = bson.M{"$ne": true}

	err := coll.UpdateOne(m.Ctx, m.Query, bson.M{"$set": bson.M{"IsRemoved": true}})
	if err != nil {
		return err
	}
	return nil
}

// FetchByQAndUpsert find One record and update or insert
func (m *Do) FetchByQAndUpsert(setValue interface{}) error {
	colName := getModelName(m.model)
	coll := DB.Collection(colName)
	if m.Query == nil {
		return errors.New("Query cannot be nil must be defined.")
	}

	if err := coll.UpdateOne(m.Ctx, m.Query, bson.M{"$set": setValue}); err != nil {
		return err
	}
	return nil
}

// FetchByQAndUpdate find One record and update no insert
func (m *Do) FetchByQAndUpdate(setValue interface{}) error {
	colName := getModelName(m.model)
	coll := DB.Collection(colName)
	if m.Query == nil {
		return errors.New("Query cannot be nil must be defined.")
	}

	if err := coll.UpdateOne(m.Ctx, m.Query, bson.M{"$set": setValue}); err != nil {
		return err
	}
	return nil
}

// ---------- internal functions -----------

// getModelName reflect string name from model
func getModelName(m interface{}) string {
	var c string
	switch m.(type) {
	case string:
		c = m.(string)
	default:
		typ := reflect.TypeOf(m)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		c = typ.Name()
	}
	return c
}

// saveLog just copy a record to Changlog
func (m *Do) saveLog(operation string) error {
	log.Println(primitive.NewObjectID()) // fix unimported warning
	id := reflect.ValueOf(m.model).Elem().FieldByName("Id").Interface()

	cl := new(ChangeLog)
	cl.ChangeReason = m.Reason
	cl.Operation = operation
	cl.ModelObjId = id.(primitive.ObjectID)
	cl.ModelName = getModelName(m.model)
	cl.ModelValue = m.model
	cl.Operator = m.Operator
	cl.CreatedBy = m.Operator
	cl.CreatedAt = time.Now()
	do := NewWithCtx(m.Ctx, cl)
	err := do.Create()
	return err
}

// findQ conduct qmgo.QueryI, skip IsRemoved: true
func (m *Do) findQ() qmgo.QueryI {
	if m.Query == nil {
		m.Query = bson.M{}
	}
	m.Query["IsRemoved"] = bson.M{"$ne": true}

	q := m.Coll.Find(m.Ctx, m.Query)
	//sort
	if m.Sort != nil {
		q = q.Sort(m.Sort...)
	}

	//skip
	if m.Skip != 0 {
		q = q.Skip(m.Skip)
	}

	//limit
	if m.Limit != 0 {
		q = q.Limit(m.Limit)
	}

	//Select
	if m.Select != nil {
		sCols := bson.M{}
		for _, v := range m.Select {
			if strings.HasPrefix(v, "-") {
				t := v[1 : len(v)-1]
				sCols[t] = -1
			} else {
				sCols[v] = 1
			}
		}
		q = q.Select(sCols)
	}

	return q
}

// findIncludeRemovedQ conduct qmgo.QueryI, including marked as removed: isRemoved: true
func (m *Do) findIncludeRemovedQ() qmgo.QueryI {
	var query qmgo.QueryI

	query = m.Coll.Find(m.Ctx, m.Query)
	//sort
	if m.Sort != nil {
		query = query.Sort(m.Sort...)
	}

	//skip
	if m.Skip != 0 {
		query = query.Skip(m.Skip)
	}

	//limit
	if m.Limit != 0 {
		query = query.Limit(m.Limit)
	}
	return query
}

// findByIdQ, skip IsRemoved:true
func (m *Do) findByIdQ() qmgo.QueryI {
	id := reflect.ValueOf(m.model).Elem().FieldByName("Id").Interface()
	m.Query = bson.M{"_id": id}
	return m.findQ()
}
