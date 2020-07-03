package database

import (
	"log"
	"sync"
	"time"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/general/utils"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type DB struct {
	conn *gorm.DB
}

var (
	dataSource        string
	maxIdleConnection int
	maxOpenConnection int
	db                DB
	lock              = new(sync.RWMutex)
)

func DBConn() *gorm.DB {
	lock.RLock()
	conn := db.conn
	lock.RUnlock()
	if conn.DB().Ping() != nil {
		lock.Lock()
		defer lock.Unlock()
		if db.conn.DB().Ping() != nil {
			log.Println("MySQL Reconnect...")
			db.conn.Close()
			db.conn, _ = gorm.Open("mysql", dataSource)
			if db.conn.DB().Ping() == nil {
				log.Println("MySQL Reconnection Successfully")
				SetMySQLConfig()
			} else {
				log.Println("MySQL Reconnection Failed")
			}
		}
		return db.conn

	}
	return conn
}

func SetMySQLConfig() {
	db.conn.DB().SetMaxIdleConns(maxIdleConnection)
	db.conn.DB().SetMaxOpenConns(maxOpenConnection)
	db.conn.DB().SetConnMaxLifetime(time.Second * 10)

	db.conn.LogMode(utils.Config().Debug.Enabled)
	db.conn.AutoMigrate(&model.Host{})
	db.conn.AutoMigrate(&model.Tag{})
}

func InitMySQL() {
	var err error
	MySQL, err := utils.MySQL()
	if err != nil {
		utils.FailOnError(err, "")
		return
	}

	dataSource = MySQL.User + ":" + MySQL.Password + "@tcp(" + MySQL.Host + ":" + MySQL.Port + ")/" + MySQL.DB +
		"?charset=" + MySQL.Charset + "&parseTime=true"

	lock.Lock()
	defer lock.Unlock()

	db.conn, err = gorm.Open("mysql", dataSource)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("MySQL Connection Successfully")

	maxIdleConnection = MySQL.MaxIdleConnection
	maxOpenConnection = MySQL.MaxOpenConnection

	SetMySQLConfig()
}
