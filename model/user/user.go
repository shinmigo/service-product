package user

//"goshop/service-product/pkg/db"

type Users struct {
	UserId     uint64
	BusinessId uint64
	UserName   string
}

//获取表名
func GetTableName() string {
	return "users"
}

func GetField() []string {
	return []string{
		"user_id", "business_id", "user_name",
	}
}

/*
根据商户id和用户id，获取用户信息
注意：如果是简单的数据库操作或者是公共的方法，也可以封装在model中，
当然了也可以封装在servicelogic中，这个根据业务场景来决定。
*/
/*
func GetUserListByUserId(userName string, businessId uint64) (*Users, error) {
	userList := &Users{}
	err := db.Conn.Table(GetTableName()).Select(GetField()).
		Where("user_id = ? AND business_id = ?", userName, businessId).
		Find(userList).Error
	if err != nil {
		return nil, fmt.Errorf("select memeber info is fail, err: %v", err)
	}

	return userList, nil
}
*/
