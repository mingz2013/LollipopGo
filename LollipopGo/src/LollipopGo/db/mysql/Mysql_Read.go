package Mysyl_DB

import (
	"database/sql"
	"fmt"
)

/*
   插入数据库数据操作
*/

func QueryFromDB(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM userinfo")
	CheckErr(err)
	if err != nil {
		fmt.Println("error:", err)
	} else {
	}
	for rows.Next() {
		var uid string
		var username string
		var departmentname string
		var created string
		var password string
		var autid string
		CheckErr(err)
		err = rows.Scan(&uid, &username, &departmentname, &created, &password, &autid)
		fmt.Println(autid)
		fmt.Println(username)
		fmt.Println(departmentname)
		fmt.Println(created)
		fmt.Println(password)
		fmt.Println(uid)
	}
}

//------------------------------------------------------------------------------
// 查询表  select 1 from tablename where uid = 'uid' limit 1;
func (this *mysql_db) ReadUserInfoData(uid string) bool {

	rows, err := this.STdb.Query("SELECT 1 FROM t_userinfo  where uid = '" + uid + "' limit 1")
	CheckErr(err)
	if err != nil {
		fmt.Println("error:", err)
	} else {
	}

	data, _ := rows.Columns()
	if len(data) > 0 {
		return true
	}

	return false
}
