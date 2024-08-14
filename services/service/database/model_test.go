package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func TestModel(t *testing.T) {
	defaultLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		Colorful:      true,
		LogLevel:      logger.Warn,
	})
	db, err := gorm.Open(mysql.Open("root:314159@tcp(127.0.0.1:3306)/cim_base?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{
		Logger: defaultLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
			NameReplacer:  strings.NewReplacer("CID", "Cid"),
		},
	})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&Group{}, &GroupMember{})
	if err != nil {
		panic(err)
	}
	//db, err = gorm.Open(mysql.Open("root:314159@tcp(127.0.0.1:3306)/cim_message?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{
	//	NamingStrategy: schema.NamingStrategy{
	//		TablePrefix:   "t_",
	//		SingularTable: true,
	//		NameReplacer:  strings.NewReplacer("CID", "Cid"),
	//	},
	//})
	//if err != nil {
	//	panic(err)
	//}
	//err = db.AutoMigrate(&MessageIndex{}, &MessageContent{})
	//if err != nil {
	//	panic(err)
	//}
}
