module troop

go 1.13

require (
	github.com/DeanThompson/ginpprof v0.0.0-20190408063150-3be636683586
	github.com/axgle/mahonia v0.0.0-20180208002826-3358181d7394
	github.com/chenhg5/collection v0.0.0-20191118032303-cb21bccce4c3
	github.com/fatih/color v1.9.0
	github.com/gin-gonic/gin v1.7.7
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jinzhu/gorm v1.9.14
	github.com/mitchellh/mapstructure v1.3.2
	github.com/spf13/cobra v1.0.0
	github.com/streadway/amqp v1.0.0
	github.com/toolkits/file v0.0.0-20160325033739-a5b3c5147e07
	github.com/toolkits/net v0.0.0-20160910085801-3f39ab6fe3ce
	github.com/troopstack/troop v0.0.0-00010101000000-000000000000
	gopkg.in/ini.v1 v1.57.0
)

replace github.com/troopstack/troop => ./
