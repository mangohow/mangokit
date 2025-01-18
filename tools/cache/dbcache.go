package cache

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/mangohow/mangokit/tools/collection"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

type dbCache[K comparable, V any] struct {
	m   collection.ConcurrentMap[K, V]
	cfg dbCacheConfig[K, V]
}

type dbCacheConfig[K comparable, V any] struct {
	selectFn    SelectFunc[V]
	insertFn    InsertFunc[V]
	deleteFn    DeleteFunc[V]
	updateFn    UpdateFunc[V]
	keyFn       KeyFunc[K, V]
	dbConn      *sqlx.DB
	table       string
	primaryKey  string
	fields      []string
	tagToName   map[string]string
	NameMap     map[string]int
	updateKeys  []string
	updateNames map[string]struct{}
	deleteKeys  []string
	logger      *logrus.Logger
}

func NewDBCache[K comparable, V any](keyFn KeyFunc[K, V], opts ...DBCacheOption[K, V]) (DBCache[K, V], error) {
	var v V
	ft := reflect.TypeOf(v)
	if ft.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("V must be a pointer of an struct")
	}
	ft = ft.Elem()
	if ft.Kind() != reflect.Struct {
		return nil, fmt.Errorf("V must be a pointer of an struct")
	}

	fields, nameMap, tagToName, updateKeys, updateNames, deleteKeys, primary := dbFields[V]()

	c := &dbCache[K, V]{
		cfg: dbCacheConfig[K, V]{
			fields:      fields,
			NameMap:     nameMap,
			tagToName:   tagToName,
			updateKeys:  updateKeys,
			updateNames: updateNames,
			deleteKeys:  deleteKeys,
			primaryKey:  primary,
			keyFn:       keyFn,
		},
	}

	for _, opt := range opts {
		opt(&c.cfg)
	}

	if (c.cfg.selectFn == nil || c.cfg.insertFn == nil || c.cfg.updateFn == nil || c.cfg.deleteFn == nil) &&
		(c.cfg.table == "" || c.cfg.dbConn == nil) {
		return nil, fmt.Errorf("table name and dbConn required")
	}

	c.genFuncs()

	return c, nil
}

type DBCacheOption[K comparable, V any] func(c *dbCacheConfig[K, V])

func WithSelectFunc[K comparable, V any](f SelectFunc[V]) DBCacheOption[K, V] {
	return func(c *dbCacheConfig[K, V]) {
		c.selectFn = f
	}
}

func WithInsertFunc[K comparable, V any](f InsertFunc[V]) DBCacheOption[K, V] {
	return func(c *dbCacheConfig[K, V]) {
		c.insertFn = f
	}
}

func WithUpdateFunc[K comparable, V any](f UpdateFunc[V]) DBCacheOption[K, V] {
	return func(c *dbCacheConfig[K, V]) {
		c.updateFn = f
	}
}

func WithDeleteFunc[K comparable, V any](f DeleteFunc[V]) DBCacheOption[K, V] {
	return func(c *dbCacheConfig[K, V]) {
		c.deleteFn = f
	}
}

func WithTableName[K comparable, V any](name string) DBCacheOption[K, V] {
	return func(c *dbCacheConfig[K, V]) {
		c.table = name
	}
}

func WithDBConn[K comparable, V any](db *sqlx.DB) DBCacheOption[K, V] {
	return func(c *dbCacheConfig[K, V]) {
		c.dbConn = db
	}
}

func WithLogger[K comparable, V any](logger *logrus.Logger) DBCacheOption[K, V] {
	return func(c *dbCacheConfig[K, V]) {
		c.logger = logger
	}
}

func (d *dbCache[K, V]) Load() error {
	vals, err := d.cfg.selectFn()
	if err != nil {
		return fmt.Errorf("load data from db failed, err: %v", err)
	}
	m := make(map[K]V, len(vals))
	for i := range vals {
		m[d.cfg.keyFn(vals[i])] = vals[i]
	}
	d.m = collection.NewConcurrentMapFromMap(m)

	return nil
}

func (d *dbCache[K, V]) Get(k K) (V, bool) {
	return d.m.Get(k)
}

func (d *dbCache[K, V]) GetBatch(ks []K) []V {
	return d.m.GetBatch(ks)
}

func (d *dbCache[K, V]) GetAll() map[K]V {
	return d.m.ToMap()
}

func (d *dbCache[K, V]) Insert(v V) error {
	if err := d.cfg.insertFn(v); err != nil {
		return fmt.Errorf("insert to db failed, err: %v", err)
	}

	d.m.Set(d.cfg.keyFn(v), v)

	return nil
}

func (d *dbCache[K, V]) Update(v V) error {
	if err := d.cfg.updateFn(v); err != nil {
		return fmt.Errorf("update to db failed, err: %v", err)
	}

	// 更新
	d.update(v)

	return nil
}

func (d *dbCache[K, V]) update(v V) {
	vv, ok := d.m.Get(d.cfg.keyFn(v))
	if !ok {
		return
	}

	rt := reflect.TypeOf(v).Elem()
	newVal := reflect.New(rt).Elem()
	rv1 := reflect.ValueOf(vv).Elem()
	rv2 := reflect.ValueOf(v).Elem()
	for i := 0; i < newVal.NumField(); i++ {
		f := newVal.Field(i)
		if !f.CanSet() {
			continue
		}
		if _, ok := d.cfg.updateNames[rt.Field(i).Name]; ok {
			f.Set(rv2.Field(i))
		} else {
			f.Set(rv1.Field(i))
		}
	}

	i := newVal.Addr().Interface().(V)
	d.m.Set(d.cfg.keyFn(i), i)
}

func (d *dbCache[K, V]) Delete(v V) error {
	if err := d.cfg.deleteFn(v); err != nil {
		return fmt.Errorf("delete from db failed, err: %v", err)
	}

	d.m.Delete(d.cfg.keyFn(v))

	return nil
}

func (d *dbCache[K, V]) genFuncs() {
	if d.cfg.selectFn == nil {
		d.genSelectFunc()
	}
	if d.cfg.insertFn == nil {
		d.genInsertFunc()
	}
	if d.cfg.updateFn == nil {
		d.genUpdateFunc()
	}
	if d.cfg.deleteFn == nil {
		d.genDeleteFunc()
	}
}

func (d *dbCache[K, V]) genSelectFunc() {
	sq := `SELECT * FROM ` + d.cfg.table
	if d.cfg.logger != nil {
		d.cfg.logger.Infof("select sql: %s", sq)
	}

	d.cfg.selectFn = func() ([]V, error) {
		res := make([]V, 0)

		if err := d.cfg.dbConn.Select(&res, sq); err != nil {
			return nil, err
		}

		return res, nil
	}
}

func (d *dbCache[K, V]) genInsertFunc() {
	fields := strings.Join(d.cfg.fields, ",")
	placeholders := strings.Repeat("?,", len(d.cfg.fields)-1) + "?"
	sq := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)", d.cfg.table, fields, placeholders)
	if d.cfg.logger != nil {
		d.cfg.logger.Infof("insert sql: %s", sq)
	}
	d.cfg.insertFn = func(v V) error {
		vals := d.getFieldValues(v)
		res, err := d.cfg.dbConn.Exec(sq, vals...)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			if d.cfg.logger != nil {
				d.cfg.logger.Errorf("get last inserted id error, err: %v", err)
			}
			return nil
		}
		// 设置主键
		d.setPrimary(v, id)
		return nil
	}
}

func (d *dbCache[K, V]) genUpdateFunc() {
	format := "UPDATE %s SET " + strings.Repeat("%s=?,", len(d.cfg.updateKeys)-1) + "%s=? WHERE " + "%s=?"
	args := make([]any, 0, len(d.cfg.updateKeys)+2)
	args = append(args, d.cfg.table)
	for i := range d.cfg.updateKeys {
		args = append(args, d.cfg.updateKeys[i])
	}
	args = append(args, d.cfg.primaryKey)
	sq := fmt.Sprintf(format, args...)
	if d.cfg.logger != nil {
		d.cfg.logger.Infof("update sql: %s", sq)
	}
	d.cfg.updateFn = func(v V) error {
		args := d.getFieldValuesByTagName(v, d.cfg.updateKeys)
		_, err := d.cfg.dbConn.Exec(sq, append(args, d.cfg.keyFn(v))...)
		return err
	}
}

func (d *dbCache[K, V]) genDeleteFunc() {
	sq := `DELETE FROM ` + d.cfg.table + ` WHERE ` + d.cfg.primaryKey + "=?"
	if d.cfg.logger != nil {
		d.cfg.logger.Infof("delete sql: %s", sq)
	}
	d.cfg.deleteFn = func(v V) error {
		_, err := d.cfg.dbConn.Exec(sq, d.getFieldValuesByTagName(v, []string{d.cfg.primaryKey})[0])
		return err
	}
}

func (d *dbCache[K, V]) getFieldValues(v any) []any {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rt := rv.Type()

	res := make([]any, rv.NumField())
	n := 0
	for i := 0; i < rv.NumField(); i++ {
		vv := rv.Field(i)
		if idx, ok := d.cfg.NameMap[rt.Field(i).Name]; ok {
			res[idx] = vv.Interface()
			n++
		}
	}

	return res[:n]
}

func (d *dbCache[K, V]) getFieldValuesByTagName(v any, tags []string) []any {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	res := make([]any, len(tags))
	for i := range tags {
		vv := rv.FieldByName(d.cfg.tagToName[tags[i]])
		res[i] = vv.Interface()
	}

	return res
}

func (d *dbCache[K, V]) setPrimary(v any, id int64) {
	if d.cfg.primaryKey == "" {
		return
	}
	rv := reflect.ValueOf(v).Elem()
	vv := rv.FieldByName(d.cfg.tagToName[d.cfg.primaryKey])
	if !vv.CanSet() {
		return
	}

	vv.SetInt(id)
}

// 获取db tag中的数据库字段名称、更新字段、删除字段、以及主键
func dbFields[V any]() ([]string, map[string]int, map[string]string, []string, map[string]struct{}, []string, string) {
	var v V
	ft := reflect.TypeOf(v)
	for ft.Kind() == reflect.Ptr {
		ft = ft.Elem()
	}

	if ft.Kind() != reflect.Struct {
		return nil, nil, nil, nil, nil, nil, ""
	}

	var (
		primary     string
		fields      = make([]string, 0, ft.NumField())
		nameMap     = make(map[string]int, ft.NumField())
		tagToName   = make(map[string]string, ft.NumField())
		updateKeys  = make([]string, 0, ft.NumField())
		updateNames = make(map[string]struct{}, ft.NumField())
		deleteKeys  = make([]string, 0, ft.NumField())
	)

	for i := 0; i < ft.NumField(); i++ {
		fv := ft.Field(i)
		tag := fv.Tag.Get("db")
		if tag == "" {
			continue
		}

		name, _, found := strings.Cut(tag, ",")
		if !found {
			name = tag
		}
		if strings.Contains(tag, "primary") {
			primary = name
		} else {
			fields = append(fields, name)
			nameMap[fv.Name] = len(fields) - 1
		}

		tagToName[name] = fv.Name
		if strings.Contains(tag, "update") {
			updateKeys = append(updateKeys, name)
			updateNames[fv.Name] = struct{}{}
		}
		if strings.Contains(tag, "delete") {
			deleteKeys = append(deleteKeys, name)
		}

	}

	return fields, nameMap, tagToName, updateKeys, updateNames, deleteKeys, primary
}
