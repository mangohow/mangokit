package cache

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"testing"
)

type CacheTest struct {
	Id      int    `db:"id,primary"`
	Name    string `db:"name"`
	Age     int    `db:"age,update"`
	Address string `db:"address,update"`
	Email   string `db:"email,update"`
}

func TestCache(t *testing.T) {
	logger := logrus.New()
	db, err := sqlx.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8")
	if err != nil {
		t.Fatal(err)
	}
	cache, err := NewDBCache[int, *CacheTest](func(c *CacheTest) int {
		return c.Id
	}, WithLogger[int, *CacheTest](logger),
		WithTableName[int, *CacheTest]("t_test"),
		WithDBConn[int, *CacheTest](db))
	if err != nil {
		t.Fatal(err)
	}
	if err = cache.Load(); err != nil {
		t.Fatal(err)
	}

	mm := cache.GetAll()
	for _, v := range mm {
		logger.Infof("%v", v)
	}
	cc := &CacheTest{
		Name:    "test",
		Age:     20,
		Address: "test",
		Email:   "test",
	}
	if err = cache.Insert(cc); err != nil {
		t.Fatal(err)
	}
	v, ok := cache.Get(cc.Id)
	if !ok {
		t.Fatal("insert error")
	}
	logger.Infof("after insert: %+v", v)
	if err := cache.Update(&CacheTest{
		Id:      cc.Id,
		Age:     25,
		Address: "china",
		Email:   "china@gmail.com",
	}); err != nil {
		t.Fatal(err)
	}
	v, ok = cache.Get(cc.Id)
	if !ok {
		t.Fatal("update error")
	}
	logger.Infof("after update: %+v", v)
	if err = cache.Delete(&CacheTest{Id: cc.Id}); err != nil {
		t.Fatal(err)
	}
	v, ok = cache.Get(cc.Id)
	if ok {
		t.Fatal("delete error")
	}
	logger.Infof("after delete: %+v", v)
}
